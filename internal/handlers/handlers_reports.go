package handlers

import (
	"encoding/json"
	"errors"
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"net/http"
	"strconv"
)

type reportForm struct {
	PostID      int
	UserID      string
	Reason      string
	FieldErrors map[string]string
	app.Validator
}

func CreateReport(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Couldn't retrieve the current user", http.StatusInternalServerError)
			return
		}

		if currentUser.Role != models.RoleNormal && currentUser.Role != models.RoleModerator && currentUser.Role != models.RoleAdmin {
			http.Error(w, "You do not have permission to make a report", http.StatusForbidden)
			return
		}

		err := r.ParseForm()
		if err != nil {
			f.ErrorLog.Println("Error parsing form", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		postIDStr := r.PathValue("id")
		reason := r.FormValue("reason")

		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "invalid post id in url", http.StatusBadRequest)
			return
		}

		_, err = f.Posts.Get(postID)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.Error(w, "The post you are trying to report does not exist or has been deleted", http.StatusNotFound)
			} else {
				f.ErrorLog.Println("Error fetching post for report", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		form := &reportForm{
			PostID: postID,
			UserID: currentUser.ID,
			Reason: reason,
		}

		form.CheckField(app.NotBlank(form.Reason), "reason", "This field cannot be blank")
		if !form.Valid() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":  "Validation failed",
				"fields": form.FieldErrors,
			})
			return
		}

		err = f.Reports.Create(form.PostID, form.UserID, form.Reason)
		if err != nil {
			f.ErrorLog.Printf("failed to create report: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("User '%s' created a report on post #%d", currentUser.ID, form.PostID)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Report created successfully"})
	}
}
