package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"forum/internal/models"
	"net/http"
	"strconv"
)

func ListReports(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		status := r.URL.Query().Get("status")
		reports, err := f.Reports.List(status)
		if err != nil {
			f.ErrorLog.Printf("Failed to retrieve reports: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(reports)
	}
}

func ResolveReport(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		action := r.PostForm.Get("action") // "resolve" or "dismiss"
		status := "resolved"
		if action == "dismiss" {
			status = "dismissed"
		}

		err = f.Reports.UpdateStatus(id, status)
		if err != nil {
			f.ErrorLog.Printf("Failed to update report status: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Report status updated", "status": status})
	}
}

func PromoteUser(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if userID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		err := f.Users.UpdateRole(userID, models.RoleModerator)
		if err != nil {
			f.ErrorLog.Printf("Failed to promote user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "User promoted successfully"})
	}
}

func DemoteUser(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID := r.PathValue("id")
		if userID == "" {
			http.Error(w, "Missing user ID", http.StatusBadRequest)
			return
		}

		err := f.Users.UpdateRole(userID, models.RoleNormal)
		if err != nil {
			f.ErrorLog.Printf("Failed to demote user: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "User demoted successfully"})
	}
}
