package handlers

import (
	"errors"
	"forum/internal/app"
	"forum/internal/models"
	"strconv"

	"net/http"
)

type postForm struct {
	Title       string
	Content     string
	Author_id   string
	Category_id int
	FieldErrors map[string]string
	app.Validator
}

func CreatePostPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "create_post.html", nil)
	}
}

func CreatePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: r.PostForm.Get("author_id"),
		}
		form.Category_id, _ = strconv.Atoi(r.PostForm.Get("category_id"))

		form.CheckField(app.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Title, 30), "title", "Title cannot exceed 30 chars")
		form.CheckField(app.NotBlank(form.Content), "content", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Content, 1000), "content", "Content cannot exceed 1000 chars")
		// TODO : Add more tests

		if !form.Valid() {
			form.FieldErrors = form.Validator.FieldErrors
			data := &app.TemplateData{Form: form}
			render(w, r, f, "posts.html", data)
			return
		}

		id, err := f.Posts.CreateNewPostDB(form.Title, form.Content, form.Author_id, form.Category_id)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateRecord) {

			}

			f.ErrorLog.Printf("Post creation failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Handle multiple file uploads
		files := r.MultipartForm.File["attachments"] // Assuming input name is "attachments"
		for _, header := range files {
			// Skip empty file inputs
			if header == nil || header.Filename == "" {
				continue
			}

			file, err := header.Open()
			if err != nil {
				f.ErrorLog.Printf("Error opening uploaded file: %v", err)
				continue // Or handle error more gracefully
			}

			// Save the file and get its path
			filePath, err := app.UploadImage(file, *header, "posts") // Save to a 'posts' subdirectory
			if err != nil {
				f.ErrorLog.Printf("Error saving uploaded file: %v", err)
				file.Close()
				continue
			}
			file.Close()

			// Save the attachment record to the database
			err = f.Attachments.CreateForPost(filePath, int64(id))
			if err != nil {
				f.ErrorLog.Printf("Failed to create attachment record for post #%d: %v", id, err)
			}
		}

		f.InfoLog.Printf("New post created with ID: %v", id)
		w.Write([]byte("Post successful!"))

	}
}

func DeletePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
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
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Post deleted successfully"))
	}
}

// WIP : Not sure if the route should handle the ID or if should be sent from the edit form
func UpdatePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		err = r.ParseForm()
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: r.PostForm.Get("author_id"),
		}
		form.Category_id, _ = strconv.Atoi(r.PostForm.Get("category_id"))

		form.CheckField(app.NotBlank(form.Title), "title", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Title, 30), "title", "Title cannot exceed 30 chars")
		form.CheckField(app.NotBlank(form.Content), "content", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Title, 1000), "title", "Title cannot exceed 1000 chars")
		// TODO : Add more tests

		if !form.Valid() {
			form.FieldErrors = form.Validator.FieldErrors
			data := &app.TemplateData{Form: form}
			render(w, r, f, "posts.html", data)
			return
		}

		err = f.Posts.UpdatePostDB(form.Title, form.Content, form.Author_id, form.Category_id, id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.Error(w, "Post not found", http.StatusNotFound)
			} else if errors.Is(err, models.ErrDuplicateRecord) {
				// This case is unlikely for an update unless you're changing to a title that already exists
				// and have a UNIQUE constraint on it.
				http.Error(w, "Update resulted in a duplicate record", http.StatusConflict)
			} else {
				f.ErrorLog.Printf("Post update failed for ID #%d: %v", id, err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		f.InfoLog.Printf("Updated post #%d", id)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Post updated successfully"))
	}
}
