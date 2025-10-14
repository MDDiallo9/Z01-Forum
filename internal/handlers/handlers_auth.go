package handlers

import (
	"forum/internal/app"
	"net/http"
	"fmt"
)

func Register(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		username := r.PostForm.Get("username")
		email := r.PostForm.Get("email")
		password := r.PostForm.Get("password")

		fmt.Println(username,email)

		// TODO: Add validation for the form fields.

		// TODO: Add real avatar support. 
		uuid, err := f.Users.Register(username, email, password, "default-avatar.jpg", 1)
		if err != nil {
			f.ErrorLog.Printf("User registration failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("New user registered with UUID: %s", uuid)
		w.Write([]byte("Registration successful!"))
	}
}
