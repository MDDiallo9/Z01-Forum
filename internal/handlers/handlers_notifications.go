package handlers

import (
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

		data := &app.TemplateData{
			Form: notifications,
		}

		render(w, r, f, "notifications.html", data)
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
		id, err := strconv.Atoi(idStr)
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

		// Redirect back to notifications page
		http.Redirect(w, r, "/notifications", http.StatusSeeOther)
	}
}
