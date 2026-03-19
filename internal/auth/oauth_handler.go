package auth

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type OAuthHandler struct {
	service     *OAuthService
	frontendURL string
}

func NewOAuthHandler(service *OAuthService, frontendURL string) *OAuthHandler {
	return &OAuthHandler{service: service, frontendURL: frontendURL}
}

func (h *OAuthHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")

	stateBytes := make([]byte, 16)
	_, _ = rand.Read(stateBytes)
	state := hex.EncodeToString(stateBytes)

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		MaxAge:   600,
		HttpOnly: true,
	})

	url, err := h.service.GetAuthURL(provider, state)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func (h *OAuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != state {
		http.Error(w, "invalid state", http.StatusBadRequest)
		return
	}

	tokenPair, err := h.service.HandleCallback(r.Context(), provider, code)
	if err != nil {
		http.Error(w, "oauth failed", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, h.frontendURL+"/auth/callback?access_token="+tokenPair.AccessToken+"&refresh_token="+tokenPair.RefreshToken, http.StatusTemporaryRedirect)
}
