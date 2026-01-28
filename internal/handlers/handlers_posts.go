package handlers

import (
	"encoding/json"
	"errors"
	"strconv"

	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"

	"net/http"
)

type postForm struct {
	Title        string
	Content      string
	Author_id    string
	Category_ids []int
	FieldErrors  map[string]string
	app.Validator
}

func CreatePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: currentUser.ID,
		}

		catStrings := r.PostForm["categories"]
		for _, catStr := range catStrings {
			catID, err := strconv.Atoi(catStr)
			if err == nil {
				form.Category_ids = append(form.Category_ids, catID)
			}
		}

		form.CheckField(app.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Title, 30), "title", "Title cannot exceed 30 chars")
		form.CheckField(app.NotBlank(form.Content), "content", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Content, 1000), "content", "Content cannot exceed 1000 chars")
		if len(form.Category_ids) == 0 {
			form.AddFieldError("categories", "At least one category must be selected")
		}

		if !form.Valid() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":  "Validation failed",
				"fields": form.FieldErrors,
			})
			return
		}

		var imageURL string
		files := r.MultipartForm.File["attachments"]
		if len(files) > 0 && files[0] != nil && files[0].Filename != "" {
			file, err := files[0].Open()
			if err != nil {
				f.ErrorLog.Printf("Error opening uploaded file: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			filename, err := app.UploadImage(file, *files[0], "posts")
			if err != nil {
				f.ErrorLog.Printf("Error saving uploaded file: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			imageURL = "/static/posts/" + filename
		}

		post := models.Post{
			Title:      form.Title,
			Content:    form.Content,
			AuthorID:   currentUser.ID,
			ImageURL:   imageURL,
			Categories: form.Category_ids,
		}

		id, err := f.Posts.CreateNewPostDB(post)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateRecord) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{"error": "Duplicate post detected"})
				return
			}
			f.ErrorLog.Printf("Post creation failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("New post created with ID: %v", id)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]any{"message": "Post created successfully", "post_id": id})
	}
}

func DeletePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := f.Posts.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		curentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		isModeratorOrAdmin := curentUser.Role == models.RoleModerator || curentUser.Role == models.RoleAdmin
		isAuthor := curentUser.ID == post.AuthorID

		if !isModeratorOrAdmin && !isAuthor {
			f.ErrorLog.Printf("User %s attempted to delete post %d without permission", curentUser.ID, id)
			http.Error(w, "You do not have permission to delete this post", http.StatusForbidden)
			return
		}

		err = f.Posts.DeletePostDB(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.Error(w, "Post not found", http.StatusNotFound)
			} else {
				f.ErrorLog.Printf("Failed to delete post #%d: %v", id, err)
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			}
			return
		}

		f.InfoLog.Printf("Deleted post #%d from database", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Post deleted successfully"})
	}
}

func UpdatePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := f.Posts.Get(id)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		currentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		isModeratorOrAdmin := currentUser.Role == models.RoleModerator || currentUser.Role == models.RoleAdmin
		isAuthor := currentUser.ID == post.AuthorID

		if !isModeratorOrAdmin && !isAuthor {
			http.Error(w, "You do not have permission to edit this post", http.StatusForbidden)
			return
		}

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: post.AuthorID,
		}

		catStrings := r.PostForm["categories"]
		for _, catStr := range catStrings {
			catID, err := strconv.Atoi(catStr)
			if err == nil {
				form.Category_ids = append(form.Category_ids, catID)
			}
		}

		form.CheckField(app.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Title, 30), "title", "Title cannot exceed 30 chars")
		form.CheckField(app.NotBlank(form.Content), "content", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Content, 1000), "content", "Content cannot exceed 1000 chars")
		if len(form.Category_ids) == 0 {
			form.AddFieldError("categories", "At least one category must be selected")
		}

		if !form.Valid() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":  "Validation failed",
				"fields": form.FieldErrors,
			})
			return
		}

		err = f.Posts.UpdatePostDB(form.Title, form.Content, form.Author_id, form.Category_ids, id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.Error(w, "Post not found", http.StatusNotFound)
			} else if errors.Is(err, models.ErrDuplicateRecord) {
				http.Error(w, "Update resulted in a duplicate record", http.StatusConflict)
			} else {
				f.ErrorLog.Printf("Post update failed for ID #%d: %v", id, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		f.InfoLog.Printf("Updated post #%d", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Post updated successfully"})
	}
}

func GetPost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		post, err := f.Posts.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
			} else {
				f.ErrorLog.Printf("Error fetching post #%d: %v", id, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		comments, err := f.Comments.GetByPostID(id)
		if err != nil {
			f.ErrorLog.Printf("Error fetching comments for post #%d: %v", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"post":     post,
			"comments": comments,
		})
	}
}
