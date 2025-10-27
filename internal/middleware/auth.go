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
const contextKeyUser = contextKey("user")

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

			ctx := context.WithValue(r.Context(), contextKeyUser, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireModerator checks if the user is a moderator or not
func RequireModerator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the context
		user, ok := r.Context().Value(contextKeyUser).(*models.User)
		if !ok {
			// This should not happen if AuthRequired is used first.
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}
		// Check if the user has the required role
		// Admins can do everything a moderator can do
		if user.Role != models.RoleModerator && user.Role != models.RoleAdmin {
			http.Error(w, "You don't have permission to do that", http.StatusForbidden)
			return
		}
		// User has permission, call the next handler
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get the user from the context
		user, ok := r.Context().Value(contextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Check if user has the required admin role.
		if user.Role != models.RoleAdmin {
			http.Error(w, "You must be an admin to do that", http.StatusForbidden)
			return
		}

		// User has admin status
		next.ServeHTTP(w, r)
	})
}
