package handlers

import (
	"errors"
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"net/http"
)

func PromoteUser(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve the current user", http.StatusInternalServerError)
			return
		}

		// Make sure only admins can get past this point
		if currentUser.Role != models.RoleAdmin {
			http.Error(w, "Permission denied", http.StatusForbidden)
			return
		}

		// Get the target userId to be promoted from the query or path
		targetID := r.PathValue("id")
		if targetID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		// Fetch the target user through their ID
		targetUser, err := f.Users.Get(targetID)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check to prevent trying to promote already existing admin
		if targetUser.Role == models.RoleAdmin {
			http.Error(w, "User is already an admin", http.StatusBadRequest)
			return
		}

		// Determine taget user's new role
		newRole := models.RoleModerator
		if targetUser.Role == models.RoleModerator {
			newRole = models.RoleAdmin
		}

		// Update the user's role in the database
		err = f.Users.UpdateRole(targetID, newRole)
		if err != nil {
			http.Error(w, "Failed to update role", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User promoted successfully"))
	}
}

func DemoteUser(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve the current user", http.StatusNotFound)
			return
		}

		// Ensure only admins can get past this point
		if currentUser.Role != models.RoleAdmin {
			http.Error(w, "Permoission denied", http.StatusForbidden)
			return
		}

		// Get the target userId to be demoted from the query or path
		targetID := r.PathValue("id")
		if targetID == "" {
			http.Error(w, "Missing user id", http.StatusBadRequest)
			return
		}

		// Fetch the target user through their ID
		targetUser, err := f.Users.Get(targetID)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		// Check to prevent the demoting a normal user
		if targetUser.Role == models.RoleNormal {
			http.Error(w, "User cannot be demoted", http.StatusBadRequest)
			return
		}

		// Determine target user's role
		newRole := models.RoleNormal
		if targetUser.Role == models.RoleAdmin {
			newRole = models.RoleModerator
		}

		// Update the user's role in the database
		err = f.Users.UpdateRole(targetID, newRole)
		if err != nil {
			http.Error(w, "Failed to update role", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("User promoted successfully"))
	}
}
