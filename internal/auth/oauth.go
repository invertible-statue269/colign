package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/uptrace/bun"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
	"golang.org/x/oauth2/google"

	"github.com/gobenpark/colign/internal/models"
)

type OAuthConfig struct {
	GitHubClientID     string
	GitHubClientSecret string
	GoogleClientID     string
	GoogleClientSecret string
	RedirectBaseURL    string
}

type OAuthService struct {
	db         *bun.DB
	jwtManager *JWTManager
	orgJoiner  OrgJoiner
	github     *oauth2.Config
	google     *oauth2.Config
}

type ProviderStatus struct {
	GitHub bool `json:"github"`
	Google bool `json:"google"`
}

func NewOAuthService(db *bun.DB, jwtManager *JWTManager, cfg OAuthConfig, orgJoiner OrgJoiner) *OAuthService {
	return &OAuthService{
		db:         db,
		jwtManager: jwtManager,
		orgJoiner:  orgJoiner,
		github: &oauth2.Config{
			ClientID:     cfg.GitHubClientID,
			ClientSecret: cfg.GitHubClientSecret,
			Endpoint:     github.Endpoint,
			RedirectURL:  cfg.RedirectBaseURL + "/api/auth/github/callback",
			Scopes:       []string{"user:email"},
		},
		google: &oauth2.Config{
			ClientID:     cfg.GoogleClientID,
			ClientSecret: cfg.GoogleClientSecret,
			Endpoint:     google.Endpoint,
			RedirectURL:  cfg.RedirectBaseURL + "/api/auth/google/callback",
			Scopes:       []string{"openid", "email", "profile"},
		},
	}
}

func (s *OAuthService) GetAuthURL(provider, state string) (string, error) {
	if !s.isProviderEnabled(provider) {
		return "", fmt.Errorf("provider not enabled: %s", provider)
	}

	cfg, err := s.getConfig(provider)
	if err != nil {
		return "", err
	}
	return cfg.AuthCodeURL(state), nil
}

func (s *OAuthService) HandleCallback(ctx context.Context, provider, code string) (*TokenPair, bool, error) {
	if !s.isProviderEnabled(provider) {
		return nil, false, fmt.Errorf("provider not enabled: %s", provider)
	}

	cfg, err := s.getConfig(provider)
	if err != nil {
		return nil, false, err
	}

	token, err := cfg.Exchange(ctx, code)
	if err != nil {
		return nil, false, fmt.Errorf("oauth exchange failed: %w", err)
	}

	userInfo, err := s.fetchUserInfo(ctx, provider, token)
	if err != nil {
		return nil, false, err
	}

	user, orgID, isNewUser, err := s.findOrCreateUser(ctx, provider, userInfo, token)
	if err != nil {
		return nil, false, err
	}

	tokenPair, err := s.createSession(ctx, user, orgID)
	if err != nil {
		return nil, false, err
	}
	return tokenPair, isNewUser, nil
}

type oauthUserInfo struct {
	ProviderID string
	Email      string
	Name       string
	AvatarURL  string
}

func (s *OAuthService) fetchUserInfo(ctx context.Context, provider string, token *oauth2.Token) (*oauthUserInfo, error) {
	client := oauth2.NewClient(ctx, oauth2.StaticTokenSource(token))

	switch provider {
	case "github":
		return s.fetchGitHubUser(client)
	case "google":
		return s.fetchGoogleUser(client)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *OAuthService) EnabledProviders() ProviderStatus {
	return ProviderStatus{
		GitHub: s.isProviderEnabled("github"),
		Google: s.isProviderEnabled("google"),
	}
}

func (s *OAuthService) fetchGitHubUser(client *http.Client) (*oauthUserInfo, error) {
	resp, err := client.Get("https://api.github.com/user")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	name := data.Name
	if name == "" {
		name = data.Login
	}

	// If email is private, fetch from emails endpoint
	email := data.Email
	if email == "" {
		resp2, err := client.Get("https://api.github.com/user/emails")
		if err == nil {
			defer func() { _ = resp2.Body.Close() }()
			var emails []struct {
				Email   string `json:"email"`
				Primary bool   `json:"primary"`
			}
			if json.NewDecoder(resp2.Body).Decode(&emails) == nil {
				for _, e := range emails {
					if e.Primary {
						email = e.Email
						break
					}
				}
			}
		}
	}

	return &oauthUserInfo{
		ProviderID: fmt.Sprintf("%d", data.ID),
		Email:      email,
		Name:       name,
		AvatarURL:  data.AvatarURL,
	}, nil
}

func (s *OAuthService) fetchGoogleUser(client *http.Client) (*oauthUserInfo, error) {
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var data struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, err
	}

	return &oauthUserInfo{
		ProviderID: data.ID,
		Email:      data.Email,
		Name:       data.Name,
		AvatarURL:  data.Picture,
	}, nil
}

func (s *OAuthService) findOrCreateUser(ctx context.Context, provider string, info *oauthUserInfo, token *oauth2.Token) (*models.User, int64, bool, error) {
	// Check if account already exists
	account := new(models.Account)
	err := s.db.NewSelect().Model(account).
		Relation("User").
		Where("a.provider = ?", provider).
		Where("a.provider_account_id = ?", info.ProviderID).
		Scan(ctx)

	if err == nil {
		// Update tokens
		account.AccessToken = token.AccessToken
		account.RefreshToken = token.RefreshToken
		if !token.Expiry.IsZero() {
			account.ExpiresAt = token.Expiry.Unix()
		}
		if _, err := s.db.NewUpdate().Model(account).WherePK().Exec(ctx); err != nil {
			slog.Warn("failed to update oauth account tokens", "account_id", account.ID, "error", err)
		}
		// Accept pending invitations on every login (without domain-based auto-join)
		if s.orgJoiner != nil {
			if _, err := s.orgJoiner.AcceptPendingInvitations(ctx, account.User.ID, account.User.Email); err != nil {
				slog.Warn("failed to accept pending invitations on oauth login", "user_id", account.User.ID, "error", err)
			}
		}
		orgID := s.getDefaultOrgID(ctx, account.User.ID)
		return account.User, orgID, false, nil
	}

	if !errors.Is(err, sql.ErrNoRows) {
		return nil, 0, false, err
	}

	// Check if user with same email exists
	user := new(models.User)
	isNewUser := false
	err = s.db.NewSelect().Model(user).Where("email = ?", info.Email).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		isNewUser = true
		// Create new user
		user = &models.User{
			Email:         info.Email,
			Name:          info.Name,
			AvatarURL:     info.AvatarURL,
			EmailVerified: true,
		}
		if _, err := s.db.NewInsert().Model(user).Exec(ctx); err != nil {
			return nil, 0, false, err
		}
	} else if err != nil {
		return nil, 0, false, err
	}

	// Link account
	newAccount := &models.Account{
		UserID:            user.ID,
		Provider:          provider,
		ProviderAccountID: info.ProviderID,
		AccessToken:       token.AccessToken,
		RefreshToken:      token.RefreshToken,
	}
	if !token.Expiry.IsZero() {
		newAccount.ExpiresAt = token.Expiry.Unix()
	}
	if _, err := s.db.NewInsert().Model(newAccount).Exec(ctx); err != nil {
		return nil, 0, false, err
	}

	// For new users: domain auto-join + invitations; for existing users: invitations only
	if s.orgJoiner != nil {
		if isNewUser {
			if _, err := s.orgJoiner.AutoJoinOrgs(ctx, user.ID, info.Email); err != nil {
				slog.Warn("failed to auto-join orgs for new oauth user", "user_id", user.ID, "error", err)
			}
		} else {
			if _, err := s.orgJoiner.AcceptPendingInvitations(ctx, user.ID, info.Email); err != nil {
				slog.Warn("failed to accept pending invitations on oauth login", "user_id", user.ID, "error", err)
			}
		}
	}

	var orgID int64
	if isNewUser {
		orgID = s.getDefaultOrgID(ctx, user.ID)

		// If no org was joined, create a personal workspace
		if orgID == 0 {
			org := &models.Organization{
				Name: fmt.Sprintf("%s's Workspace", info.Name),
				Slug: fmt.Sprintf("%s-oauth", info.ProviderID),
			}
			if _, err := s.db.NewInsert().Model(org).Exec(ctx); err != nil {
				return nil, 0, false, err
			}
			om := &models.OrganizationMember{
				OrganizationID: org.ID,
				UserID:         user.ID,
				Role:           models.OrgRoleOwner,
			}
			if _, err := s.db.NewInsert().Model(om).Exec(ctx); err != nil {
				return nil, 0, false, err
			}
			orgID = org.ID
		}
	} else {
		orgID = s.getDefaultOrgID(ctx, user.ID)
	}

	return user, orgID, isNewUser, nil
}

func (s *OAuthService) getDefaultOrgID(ctx context.Context, userID int64) int64 {
	om := new(models.OrganizationMember)
	err := s.db.NewSelect().Model(om).
		Where("user_id = ?", userID).
		OrderExpr("created_at ASC").
		Limit(1).
		Scan(ctx)
	if err != nil {
		return 0
	}
	return om.OrganizationID
}

func (s *OAuthService) createSession(ctx context.Context, user *models.User, orgID int64) (*TokenPair, error) {
	tokenPair, err := s.jwtManager.GenerateTokenPair(user.ID, user.Email, user.Name, orgID)
	if err != nil {
		return nil, err
	}

	session := &models.Session{
		UserID:       user.ID,
		OrgID:        orgID,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    time.Now().Add(RefreshTokenDuration),
	}
	if _, err := s.db.NewInsert().Model(session).Exec(ctx); err != nil {
		return nil, err
	}

	return tokenPair, nil
}

func (s *OAuthService) getConfig(provider string) (*oauth2.Config, error) {
	switch provider {
	case "github":
		return s.github, nil
	case "google":
		return s.google, nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (s *OAuthService) isProviderEnabled(provider string) bool {
	cfg, err := s.getConfig(provider)
	if err != nil || cfg == nil {
		return false
	}

	return strings.TrimSpace(cfg.ClientID) != "" && strings.TrimSpace(cfg.ClientSecret) != ""
}
