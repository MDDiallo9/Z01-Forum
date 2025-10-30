package handlers

import (
	"encoding/json"
	"forum/internal/app"
	"net/http"
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
