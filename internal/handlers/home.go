package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"net/http"
)

func Home(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		posts, err := f.Posts.ListAll()
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"posts":      posts,
			"categories": categories,
			"message":    "Welcome to the Forum API",
		})
	}
}
