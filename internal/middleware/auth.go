package middleware

import (
	"context"
	"net/http"
)

// bcrypt hash/verify

// SessionManager defines the dependency needed by AuthRequired
type SessionManager interface {
	GetUserFromRequest(r *http.Request) (string, error)
}

// Authentiation Middleware for sessions
func AuthRequired(sessions SessionManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := sessions.GetUserFromRequest(r)
			if err != nil || userID == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
