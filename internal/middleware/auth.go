package middleware

import (
	"context"
	"forum/internal/app"
	"net/http"
)

// bcrypt hash/verify

// Authentiation Middleware for sessions
func AuthRequired(f *app.Application) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID, err := f.Sessions.GetUserFromRequest(r)
			if err != nil || userID == "" {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}

			ctx := context.WithValue(r.Context(), "userID", userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
