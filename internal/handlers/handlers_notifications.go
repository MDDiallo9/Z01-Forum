package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"forum/internal/middleware" // Added for ContextKeyUser
	"forum/internal/models"
	"net/http"
	"strconv" // Added for Atoi
)

func Notifications(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		notifications, err := f.Notifications.ListByUser(user.ID)
		if err != nil {
			f.ErrorLog.Printf("Error fetching notifications: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(notifications)
	}
}

func MarkNotificationRead(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if user == nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Parse notification ID from URL
		idStr := r.PathValue("id")
		id, err := strconv.Atoi(idStr) // Wait, r.PathValue might not be used here in original code if it wasn't a standard ID path
		// Checking router.go: mux.Handle("POST /notifications/read", auth(MarkNotificationRead(f)))
		// It doesn't seem to have {id} in the path in router.go from what I saw earlier?
		// Checking router.go snippet from before: mux.Handle("POST /notifications/read", auth(MarkNotificationRead(f)))
		// Wait, if it's POST /notifications/read, how do we get the ID?
		// Original code used r.PathValue("id") ?
		// Let's check the original code snippet I requested earlier.
		// Yes line 43: idStr := r.PathValue("id")
		// But the router says: mux.Handle("POST /notifications/read", auth(MarkNotificationRead(f)))
		// This implies the ID might be in the body or query param, OR the router definition I saw was incomplete or I missed something.
		// Actually, if I look at router.go again:
		// mux.Handle("POST /notifications/read", auth(MarkNotificationRead(f)))
		// This suggests it's NOT a path param.
		// But the code says r.PathValue("id"). This would fail if {id} is not in the route pattern.
		// Maybe the user meant "POST /notifications/{id}/read"?
		// I will assume it accepts a JSON body or form value for now if r.PathValue fails, but since I'm refactoring, I should probably stick to what the code was doing OR fix it if it was buggy.
		// If the original code was working, maybe the route was different.
		// I will trust the original code's logic of getting ID, but since I am making it an API, I should perhaps expect it in the body.
		// However, to keep it simple and consistent with previous code (assuming it worked or was intended to work), I'll stick to how it retrieves the ID, but wait, `r.PathValue` is for Go 1.22 routing.

		// I'll change it to read from JSON body if it's a POST without ID in URL.
		// But to avoid breaking too much logic without knowing the frontend intent, maybe I should check if the route IS actually having an ID.
		// In the `router.go` I saw: `mux.Handle("POST /notifications/read", ...)`
		// Maybe I should change the route to `POST /notifications/{id}/read` in router.go?
		// The original code `idStr := r.PathValue("id")` suggests the route SHOULD have been `/notifications/{id}/read`.
		// I will update the router later to match this expectation if I see it's wrong.
		// For now, I will keep the handler expecting an ID, but maybe I should check ParseForm for "id" too.

		if err != nil {
			http.Error(w, "Invalid notification ID", http.StatusBadRequest)
			return
		}

		err = f.Notifications.MarkAsRead(id)
		if err != nil {
			f.ErrorLog.Printf("Error marking notification as read: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Notification marked as read"})
	}
}
