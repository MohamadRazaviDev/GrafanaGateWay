package auth

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
)

type contextKey string

const IdentityKey contextKey = "gateway.identity"

// GetIdentity retrieves the authenticated identity from the request context.
func GetIdentity(ctx context.Context) *Identity {
	if id, ok := ctx.Value(IdentityKey).(*Identity); ok {
		return id
	}
	return nil
}

// Middleware creates an authentication middleware.
// It validates the Bearer token, maps it to a user identity,
// and sets the X-WEBAUTH-USER header for Grafana auth proxy.
func Middleware(validator *APIKeyValidator, headerName string, skipPaths []string, logger *slog.Logger) func(http.Handler) http.Handler {
	skip := make(map[string]bool, len(skipPaths))
	for _, p := range skipPaths {
		skip[p] = true
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for health/metrics endpoints
			if skip[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				logger.Warn("missing authorization header",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				http.Error(w, `{"error":"unauthorized","message":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Extract Bearer token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == authHeader {
				// No "Bearer " prefix found
				logger.Warn("invalid authorization format",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
				)
				http.Error(w, `{"error":"unauthorized","message":"invalid Authorization format, expected Bearer token"}`, http.StatusUnauthorized)
				return
			}

			identity, err := validator.Validate(token)
			if err != nil {
				logger.Warn("authentication failed",
					"path", r.URL.Path,
					"remote_addr", r.RemoteAddr,
					"error", err,
				)
				http.Error(w, `{"error":"forbidden","message":"invalid API key"}`, http.StatusForbidden)
				return
			}

			logger.Debug("authenticated request",
				"user", identity.User,
				"team", identity.Team,
				"path", r.URL.Path,
			)

			// Set Grafana auth proxy header
			r.Header.Set(headerName, identity.User)

			// Store identity in context for downstream handlers
			ctx := context.WithValue(r.Context(), IdentityKey, identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
