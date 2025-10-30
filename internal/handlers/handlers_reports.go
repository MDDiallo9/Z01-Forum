package handlers

import (
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

func CreateReportPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "create_report.html", nil)
	}
}

func CreateReport(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract the user trying to make the report from context
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Couldn't retrieve the current user", http.StatusInternalServerError)
			return
		}

		// Check the role authorization
		if currentUser.Role != models.RoleNormal && currentUser.Role != models.RoleModerator && currentUser.Role != models.RoleAdmin {
			http.Error(w, "You do not have permission to make a report", http.StatusForbidden)
			return
		}

		// Parse the report form details
		err := r.ParseForm()
		if err != nil {
			f.ErrorLog.Println("Error parsing form", err)
			http.Error(w, "Bad Request", http.StatusBadRequest)
		}

		postIDStr := r.PathValue("id")
		// userIDstr := r.FormValue("user_id")
		reason := r.FormValue("reason")

		// Convert the integer of post_id and user_id to strings
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "invalid post id in url", http.StatusBadRequest)
			return
		}

		// Verify that the post actually exists
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

		// Assign form values
		form := &reportForm{
			PostID: postID,
			UserID: currentUser.ID,
			Reason: reason,
		}

		form.CheckField(app.NotBlank(form.Reason), "reason", "This field cannot be blank")
		if !form.Valid() {
			form.FieldErrors = form.Validator.FieldErrors
			data := &app.TemplateData{Form: form}
			render(w, r, f, "create_report.html", data)
			return
		}

		// Create the report in DB
		err = f.Reports.Create(form.PostID, form.UserID, form.Reason)
		if err != nil {
			f.ErrorLog.Printf("failed to create report: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Response
		f.InfoLog.Printf("User '%s' created a report on post #%d", currentUser.ID, form.PostID)
		// Redirect user back to the post
		http.Redirect(w, r, "/post/"+postIDStr, http.StatusSeeOther)
	}
}

func ListReports(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Couldn't retrieve the current user", http.StatusInternalServerError)
			return
		}

		if currentUser.Role != models.RoleAdmin {
			http.Error(w, "You do not have the permission to view reports", http.StatusForbidden)
			return
		}

		// Get the status filter from the URL query string
		statusFilter := r.URL.Query().Get("status")

		// Return list of reports (all reports!)
		reports, err := f.Reports.List(statusFilter)
		if err != nil {
			f.ErrorLog.Printf("Failed to retrieve reports: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Render the reports fetched
		data := &app.TemplateData{
			Reports: reports,
		}
		render(w, r, f, "admin_reports.html", data)
	}
}

func ResolveReport(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Couldn't retrieve the current user", http.StatusInternalServerError)
			return
		}

		if currentUser.Role != models.RoleAdmin {
			http.Error(w, "You do not have the permission to perform this task", http.StatusForbidden)
			return
		}

		reportIDStr := r.PathValue("id")
		newStatus := "resolved"

		// Convert reportIDStr to an integer
		reportID, err := strconv.Atoi(reportIDStr)
		if err != nil {
			http.Error(w, "Invalid report ID provided", http.StatusBadRequest)
			return
		}

		err = f.Reports.UpdateStatus(reportID, newStatus)
		if err != nil {
			http.Error(w, "Failed to update report status", http.StatusInternalServerError)
			return
		}

		// Redirect to lists
		http.Redirect(w, r, "/admin/reports", http.StatusSeeOther)
	}
}
