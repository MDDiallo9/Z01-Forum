package handlers

import (
	"errors"
	"forum/internal/app"
	"forum/internal/middleware"
	"forum/internal/models"
	"strconv"

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

func CreatePostPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Fetch categories to display in the form
		categories, err := f.Categories.ListAll()
		if err != nil {
			f.ErrorLog.Printf("Error fetching categories: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Categories []*models.Category
			Form       *postForm
		}

		render(w, r, f, "create_post.html", &app.TemplateData{Form: &PageData{Categories: categories, Form: &postForm{}}})
	}
}

func CreatePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Form parsing knowledge
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// The context we created in midlleware (auth.go) bears the user,
		// it's ID will be defined for use by the Author_id
		currentUser, ok := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if !ok {
			http.Error(w, "Could not retrieve user from context", http.StatusInternalServerError)
			return
		}

		form := &postForm{
			Title:   r.PostForm.Get("title"),
			Content: r.PostForm.Get("content"),
			// Realistically, Author_id isn't goten from the form, but from the authenticated user (sessions)
			Author_id: currentUser.ID,
		}

		// Parse categories
		// r.PostForm["categories"] should give a slice of strings if multiple checkboxes have name="categories"
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
			form.FieldErrors = form.Validator.FieldErrors

			// Re-fetch categories for re-rendering
			categories, _ := f.Categories.ListAll()
			type PageData struct {
				Categories []*models.Category
				Form       *postForm
			}

			data := &app.TemplateData{Form: &PageData{Categories: categories, Form: form}}
			render(w, r, f, "create_post.html", data)
			return
		}

		id, err := f.Posts.CreateNewPostDB(form.Title, form.Content, form.Author_id, form.Category_ids)
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
		// w.Write([]byte("Post successful!"))
		http.Redirect(w, r, "/", http.StatusSeeOther)

	}
}

func DeletePost(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		// Get the post author,, as the id would be needed to help moderator/admin rights for deleteing
		post, err := f.Posts.Get(id)
		if err != nil {
			if errors.Is(err, models.ErrNoRecords) {
				http.NotFound(w, r)
			} else {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			}
			return
		}

		// Get current user with our context
		curentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)

		// AUTHORIZATION CHECK
		isModeratorOrAdmin := curentUser.Role == models.RoleModerator || curentUser.Role == models.RoleAdmin
		isAuthor := curentUser.ID == post.AuthorID

		if !isModeratorOrAdmin && !isAuthor {
			f.ErrorLog.Printf("User %s attempted to delete post %d without permission", curentUser.ID, id)
			http.Error(w, "You do not have permission to delete this post", http.StatusForbidden)
			return
		}

		// If our code excutes up to this point, then the person trying too delete has been authenticated and authorized
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

		post, err := f.Posts.Get(id)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Get the currently logged-in user
		currentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)

		// AUTHORIZATION Check
		isModeratorOrAdmin := currentUser.Role == models.RoleModerator || currentUser.Role == models.RoleAdmin
		isAuthor := currentUser.ID == post.AuthorID

		if !isModeratorOrAdmin && !isAuthor {
			http.Error(w, "You do not have permission to edit this post", http.StatusForbidden)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: r.PostForm.Get("author_id"),
		}

		// Parse categories
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
		form.CheckField(app.MaxChars(form.Title, 1000), "title", "Title cannot exceed 1000 chars")
		if len(form.Category_ids) == 0 {
			form.AddFieldError("categories", "At least one category must be selected")
		}

		if !form.Valid() {
			form.FieldErrors = form.Validator.FieldErrors

			// Re-fetch categories for re-rendering
			categories, _ := f.Categories.ListAll()
			type PageData struct {
				Categories []*models.Category
				Form       *postForm
			}

			data := &app.TemplateData{Form: &PageData{Categories: categories, Form: form}}
			render(w, r, f, "post.html", data)
			return
		}

		err = f.Posts.UpdatePostDB(form.Title, form.Content, form.Author_id, form.Category_ids, id)
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

		type PageData struct {
			Post     *models.Post
			Comments []*models.Comment
		}

		data := &app.TemplateData{
			Form: &PageData{
				Post:     post,
				Comments: comments,
			},
		}

		render(w, r, f, "post.html", data)
	}
}

func CategoryPosts(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "Invalid category ID", http.StatusBadRequest)
			return
		}

		// Fetch posts for this category
		posts, err := f.Posts.ListByCategory(id, 100) // Limit 100 for now
		if err != nil {
			f.ErrorLog.Printf("Error fetching posts for category #%d: %v", id, err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Fetch all categories for the sidebar
		categories, err := f.Categories.ListAll()
		if err != nil {
			f.ErrorLog.Printf("Error fetching categories: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		type PageData struct {
			Posts      []*models.Post
			Categories []*models.Category
			CurrentCat int
		}

		data := &app.TemplateData{
			Form: &PageData{
				Posts:      posts,
				Categories: categories,
				CurrentCat: id,
			},
		}

		render(w, r, f, "home.html", data)
	}
}
