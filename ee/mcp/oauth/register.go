package oauth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"

	"github.com/uptrace/bun"
)

type RegisterHandler struct {
	db *bun.DB
}

func NewRegisterHandler(db *bun.DB) *RegisterHandler {
	return &RegisterHandler{db: db}
}

// ServeHTTP handles POST /oauth/register — Dynamic Client Registration (RFC 7591).
func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		ClientName   string   `json:"client_name"`
		RedirectURIs []string `json:"redirect_uris"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ClientName == "" {
		req.ClientName = "MCP Client"
	}

	clientID, err := generateClientID()
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	client := &OAuthClient{
		ClientID:     clientID,
		ClientName:   req.ClientName,
		RedirectURIs: req.RedirectURIs,
	}

	if _, err := h.db.NewInsert().Model(client).Exec(r.Context()); err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]any{
		"client_id":     clientID,
		"client_name":   req.ClientName,
		"redirect_uris": req.RedirectURIs,
	})
}

func generateClientID() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "colign_" + hex.EncodeToString(b), nil
}
