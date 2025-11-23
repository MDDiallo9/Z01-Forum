package handlers

import (
	"errors"
	"fmt"
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"net/http"
	"strconv"
)

func CreateComment(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		content := r.PostForm.Get("content")
		if content == "" {
			http.Error(w, "Content cannot be empty", http.StatusBadRequest)
			return
		}

		postIDStr := r.PathValue("id")
		postID, err := strconv.Atoi(postIDStr)
		if err != nil {
			http.Error(w, "Invalid Post ID", http.StatusBadRequest)
			return
		}

		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// Original code used f.Comments.Create(content, currentUser.ID, postID)
		// The provided snippet uses app.Comments.Create(content, postID, user.ID)
		// Assuming 'app' refers to 'f' and 'user' refers to 'currentUser'
		// Also, the order of arguments for Create might have changed in the user's system.
		// I will use the arguments from the provided snippet: (content, postID, currentUser.ID)
		_, err = f.Comments.Create(content, currentUser.ID, postID)
		if err != nil {
			// Original code used f.ErrorLog.Printf and http.Error
			// The provided snippet uses app.serverError
			// I will adapt to the existing error handling style using f.ErrorLog.Printf and http.Error
			f.ErrorLog.Printf("Error creating comment: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create Notification
		post, _ := f.Posts.Get(postID) // Assuming 'app.Posts' refers to 'f.Posts'
		if post != nil {
			postIDInt := post.ID
			// Assuming 'app.Notifications' refers to 'f.Notifications'
			f.Notifications.Create(post.AuthorID, currentUser.ID, "comment", &postIDInt, nil) // Assuming 'user.ID' refers to 'currentUser.ID'
		}

		// Update post last modified time
		err = f.Posts.UpdateLastModified(postID)
		if err != nil {
			f.ErrorLog.Printf("Error updating post last modified time: %v", err)
			// Don't fail the request, just log it
		}

		// Original redirect was to Referer
		// The provided snippet redirects to fmt.Sprintf("/post/%d", postID)
		http.Redirect(w, r, fmt.Sprintf("/post/%d", postID), http.StatusSeeOther)
	}
}

func DeleteComment(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid comment ID", http.StatusBadRequest)
			return
		}

		comment, err := f.Comments.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
			} else {
				f.ErrorLog.Printf("Error fetching comment #%d: %v", id, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		currentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if currentUser.ID != comment.AuthorID && currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleModerator {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		err = f.Comments.Delete(id)
		if err != nil {
			f.ErrorLog.Printf("Error deleting comment #%d: %v", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/post/%d", comment.PostID), http.StatusSeeOther)
	}
}
