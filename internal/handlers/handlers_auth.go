package handlers

import (
	"fmt"
	"forum/internal/app"
	"net/http"
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

		fmt.Println(username, email)

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

func Login(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		emailOrUsername := r.PostForm.Get("emailORUsername")
		password := r.PostForm.Get("password")

		// TODO: Add validation for the form fields.

		id, err := f.Users.Authenticate(emailOrUsername, password)
		if err != nil {
			f.ErrorLog.Printf("User authentication failed: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// TODO: Create a session for the user.

		f.InfoLog.Printf("User with ID %s logged in successfully", id)
		w.Write([]byte("Login successful!"))
	}
}
