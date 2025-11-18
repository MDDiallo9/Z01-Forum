package handlers

import (
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"net/http"
	"strconv"
)

func LikePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleReaction(w, r, f, true, 1)
	}
}

func DislikePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleReaction(w, r, f, true, -1)
	}
}

func LikeComment(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleReaction(w, r, f, false, 1)
	}
}

func DislikeComment(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		handleReaction(w, r, f, false, -1)
	}
}

func handleReaction(w http.ResponseWriter, r *http.Request, f *app.Application, isPost bool, reactionType int) {
	currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	var postID, commentID *int
	if isPost {
		postID = &id
	} else {
		commentID = &id
	}

	err = f.Likes.AddLike(currentUser.ID, postID, commentID, reactionType)
	if err != nil {
		f.ErrorLog.Printf("Error adding reaction: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Create Notification if it's a post reaction
	if isPost {
		post, err := f.Posts.Get(id)
		if err != nil {
			f.ErrorLog.Printf("Error getting post for notification: %v", err)
			// Continue without notification if post not found, or handle as needed
		} else if post != nil && post.AuthorID != currentUser.ID { // Don't notify self
			var notificationType string
			if reactionType == 1 {
				notificationType = "like"
			} else {
				notificationType = "dislike"
			}
			postIDInt := post.ID
			f.Notifications.Create(post.AuthorID, currentUser.ID, notificationType, &postIDInt, nil)
		}
	}

	// Redirect back to the page
	http.Redirect(w, r, r.Header.Get("Referer"), http.StatusSeeOther)
}
