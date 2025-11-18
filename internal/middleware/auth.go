package middleware

import (
	"context"
	"forum/internal/models"
	"net/http"
)

// bcrypt hash/verify

// Use a custom type for context keys to avoid collissioons
type contextKey string

// contextKeyUser is the key used to store the user object in the request context
const ContextKeyUser = contextKey("user")

// SessionManager defines the dependency needed by AuthRequired
type SessionManager interface {
	GetUserFromRequest(r *http.Request) (string, error)
}

// Authentiation Middleware for sessions
func AuthRequired(sessions SessionManager, userModel *models.UsersModel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Get user ID from the session
			userID, err := sessions.GetUserFromRequest(r)
			if err != nil || userID == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			// Use the user ID to fetch the full user object from the DB
			user, err := userModel.Get(userID)
			if err != nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// LoadSession checks for a session and loads the user if present, but doesn't require it.
func LoadSession(sessions SessionManager, userModel *models.UsersModel) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := sessions.GetUserFromRequest(r)
			if err == nil && userID != "" {
				user, err := userModel.Get(userID)
				if err == nil {
					ctx := context.WithValue(r.Context(), ContextKeyUser, user)
					next.ServeHTTP(w, r.WithContext(ctx))
					return
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}
