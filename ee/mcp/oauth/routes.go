package oauth

import (
	"net/http"

	"github.com/uptrace/bun"

	"github.com/gobenpark/colign/internal/apitoken"
	"github.com/gobenpark/colign/internal/auth"
)

// RegisterRoutes adds OAuth endpoints to the mux and wraps the MCP handler with auth middleware.
// Returns a middleware that should wrap the MCP handler.
func RegisterRoutes(mux *http.ServeMux, db *bun.DB, jwtManager *auth.JWTManager, apiTokenService *apitoken.Service, baseURL string) func(http.Handler) http.Handler {
	// Discovery endpoints
	mux.HandleFunc("GET /.well-known/oauth-protected-resource", ProtectedResourceMetadata(baseURL))
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", AuthorizationServerMetadata(baseURL))

	// OAuth endpoints
	authorizeHandler := NewAuthorizeHandler(db, jwtManager, baseURL)
	mux.Handle("/oauth/authorize", authorizeHandler)

	tokenHandler := NewTokenHandler(db, apiTokenService)
	mux.Handle("/oauth/token", tokenHandler)

	registerHandler := NewRegisterHandler(db)
	mux.Handle("/oauth/register", registerHandler)

	// Return middleware for MCP endpoint
	return func(next http.Handler) http.Handler {
		return MCPAuthMiddleware(baseURL, next)
	}
}
