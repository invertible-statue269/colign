package oauth

import (
	"fmt"
	"net/http"
)

// MCPAuthMiddleware wraps the MCP handler to return 401 + WWW-Authenticate
// when no Bearer token is present, enabling OAuth discovery.
func MCPAuthMiddleware(baseURL string, next http.Handler) http.Handler {
	resourceMetadataURL := baseURL + "/.well-known/oauth-protected-resource"

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || len(auth) < 8 {
			w.Header().Set("WWW-Authenticate", fmt.Sprintf(
				`Bearer resource_metadata="%s"`, resourceMetadataURL,
			))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
