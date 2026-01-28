package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"net/http"
)

func UserCreatedPosts(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		posts, err := f.Posts.ListByAuthor(currentUser.ID)
		if err != nil {
			f.ErrorLog.Printf("Error fetching user posts: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}

func UserLikedPosts(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		posts, err := f.Posts.ListLikedByUser(currentUser.ID)
		if err != nil {
			f.ErrorLog.Printf("Error fetching liked posts: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(posts)
	}
}

func UserComments(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		comments, err := f.Comments.ListByUser(currentUser.ID)
		if err != nil {
			f.ErrorLog.Printf("Error fetching user comments: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(comments)
	}
}
