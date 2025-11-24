package handlers

import (
	"forum/internal/app"
	"forum/internal/models"
	"net/http"
)

func Home(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := f.Posts.ListAll() // Fetch all posts sorted by activity
		if err != nil {
			f.ErrorLog.Printf("Error fetching posts: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		categories, err := f.Categories.ListAll()
		if err != nil {
			f.ErrorLog.Printf("Error fetching categories: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		userCount, err := f.Users.Count()
		if err != nil {
			f.ErrorLog.Printf("Error fetching user count: %v", err)
			userCount = 0
		}

		postCount, err := f.Posts.Count()
		if err != nil {
			f.ErrorLog.Printf("Error fetching post count: %v", err)
			postCount = 0
		}

		type PageData struct {
			Posts      []*models.Post
			Categories []*models.Category
			UserCount  int
			PostCount  int
		}

		data := &app.TemplateData{
			Form: &PageData{
				Posts:      posts,
				Categories: categories,
				UserCount:  userCount,
				PostCount:  postCount,
			},
		}

		render(w, r, f, "home.html", data)
	}
}

func RegisterPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "register.html")
	}
}

func LoginPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "login.html")
	}
}

func LogoutPopUp(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "logout_popup.html")
	}
}
