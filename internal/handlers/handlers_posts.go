package handlers

import (
	"errors"
	"fmt"
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

		// Handle multiple file uploads (Attachments) - Optional, if you want to keep this logic separate from the main image
		// For now, we are focusing on the main post image handled above.
		// If you want to support additional attachments, keep the logic here.
		var imageURL string
		files := r.MultipartForm.File["attachments"] // Assuming input name is "attachments"
		if len(files) > 0 && files[0] != nil && files[0].Filename != "" {
			file, err := files[0].Open()
			if err != nil {
				f.ErrorLog.Printf("Error opening uploaded file: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			defer file.Close()

			imageURL, err = app.UploadImage(file, *files[0], "posts") // Save to a 'posts' subdirectory
			if err != nil {
				f.ErrorLog.Printf("Error saving uploaded file: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
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
				// Handle duplicate record if necessary
			}

			f.ErrorLog.Printf("Post creation failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("New post created with ID: %v", id)
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

func EditPostPage(f *app.Application) http.HandlerFunc {
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

		// Get the currently logged-in user
		currentUser := r.Context().Value(middleware.ContextKeyUser).(*models.User)
		if currentUser.ID != post.AuthorID && currentUser.Role != models.RoleAdmin && currentUser.Role != models.RoleModerator {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		categories, err := f.Categories.ListAll()
		if err != nil {
			f.ErrorLog.Printf("Error fetching categories: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Convert post categories to ID slice for the form
		var categoryIDs []int
		// post.Categories is []int based on previous edits to models/posts.go
		categoryIDs = post.Categories

		form := &postForm{
			Title:        post.Title,
			Content:      post.Content,
			Category_ids: categoryIDs,
		}

		type PageData struct {
			Categories []*models.Category
			Form       *postForm
			PostID     int
		}

		data := &app.TemplateData{
			Form: &PageData{
				Categories: categories,
				Form:       form,
				PostID:     post.ID,
			},
		}

		render(w, r, f, "edit_post.html", data)
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

		err = r.ParseForm()
		if err != nil {
			http.Error(w, "Bad Request", http.StatusBadRequest)
			return
		}

		form := &postForm{
			Title:     r.PostForm.Get("title"),
			Content:   r.PostForm.Get("content"),
			Author_id: post.AuthorID, // Keep original author
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
				PostID     int
			}

			data := &app.TemplateData{
				Form: &PageData{
					Categories: categories,
					Form:       form,
					PostID:     id,
				},
			}
			render(w, r, f, "edit_post.html", data)
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
		http.Redirect(w, r, fmt.Sprintf("/post/%d", id), http.StatusSeeOther)
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
