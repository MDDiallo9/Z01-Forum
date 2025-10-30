package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"net/http"
	"strconv"
)

func ListPosts(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// A method that calls 20 random posts
		posts, err := f.Posts.ListRandom(20)
		if err != nil {
			f.ErrorLog.Println("Failed to get posts: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Send JSON data to the browser using (Content-Type header)
		w.Header().Set("Content-Type", "application/json")

		// Encode the posts into JSON and write it to the response
		err = json.NewEncoder(w).Encode(posts)
		if err != nil {
			f.ErrorLog.Println("Failed to encode posts to JSON: ", err)
		}
	}
}

func ListPostsByCategory(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get the category ID from the URL path, and convert to int
		categoryID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.Error(w, "invalid category ID", http.StatusBadRequest)
			return
		}

		// Call the ListByCategory model
		posts, err := f.Posts.ListByCategory(categoryID, 50)
		if err != nil {
			f.ErrorLog.Println("Failed to get posts by category:", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// Encode and send the JSON response.
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(posts)
		if err != nil {
			f.ErrorLog.Println("Failed to encode posts to JSON:", err)
		}
	}
}

func ListCategories(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		categories, err := f.Categories.ListAll()
		if err != nil {
			f.ErrorLog.Println("Failed to get categories: ", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		err = json.NewEncoder(w).Encode(categories)
		if err != nil {
			f.ErrorLog.Println("Failed to encode categories to JSON: ", err)
		}
	}
}
