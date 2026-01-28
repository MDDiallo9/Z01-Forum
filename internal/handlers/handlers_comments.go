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

		_, err = f.Comments.Create(content, currentUser.ID, postID)
		if err != nil {
			f.ErrorLog.Printf("Error creating comment: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Create Notification
		post, _ := f.Posts.Get(postID) // Assuming 'app.Posts' refers to 'f.Posts'
		if post != nil {
			postIDInt := post.ID
			f.Notifications.Create(post.AuthorID, currentUser.ID, "comment", &postIDInt, nil)
		}

		// Update post last modified time
		err = f.Posts.UpdateLastModified(postID)
		if err != nil {
			f.ErrorLog.Printf("Error updating post last modified time: %v", err)
			// Don't fail the request, just log it
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Comment created successfully"})
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Comment deleted successfully"})
	}
}
