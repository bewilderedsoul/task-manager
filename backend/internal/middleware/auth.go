// Package middleware provides authentication/authorization middleware.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"taskmanager/internal/auth"
	"taskmanager/internal/httpx"
)

type ctxKey string

const userCtxKey ctxKey = "user"

// AuthUser is the minimal identity attached to the request context.
type AuthUser struct {
	ID   uuid.UUID
	Role string
}

// RequireAuth validates the Bearer token and stores the identity in context.
// Requests without a valid token receive 401.
func RequireAuth(jwtManager *auth.Manager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
				httpx.Error(w, http.StatusUnauthorized, "missing or malformed Authorization header")
				return
			}

			claims, err := jwtManager.Parse(parts[1])
			if err != nil {
				httpx.Error(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), userCtxKey, AuthUser{ID: claims.UserID, Role: claims.Role})
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// UserFromContext returns the authenticated user stored by RequireAuth.
func UserFromContext(ctx context.Context) (AuthUser, bool) {
	u, ok := ctx.Value(userCtxKey).(AuthUser)
	return u, ok
}
