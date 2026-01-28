package handlers

import (
	"encoding/json"
	"errors"
	"forum/internal/app"
	"forum/internal/auth"
	"forum/internal/models"
	"log"

	"net/http"
)

type userRegistrationForm struct {
	Username    string
	Password    string
	Email       string
	Avatar      string
	FieldErrors map[string]string
	app.Validator
}

func Register(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseMultipartForm(10 << 20)
		if err != nil {
			f.ErrorLog.Printf("Form parsing error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		form := &userRegistrationForm{}
		form.Username = r.PostForm.Get("username")
		form.Email = r.PostForm.Get("email")
		form.Password = r.PostForm.Get("password")
		confirmPassword := r.PostForm.Get("confirm_password")

		// Checking form fields
		form.CheckField(app.NotBlank(form.Username), "username", "This field cannot be blank")
		form.CheckField(app.MaxChars(form.Username, 10), "username", "Username too long")
		form.CheckField(app.NotBlank(form.Email), "email", "This field cannot be blank")
		form.CheckField(app.ValidEmail(form.Email), "email", "Invalid Email format")
		form.CheckField(app.NotBlank(form.Password), "password", "This field cannot be blank")
		form.CheckField(app.MinChars(form.Password, 6), "password", "Minimum 6 characters")
		form.CheckField(app.IsIdentical(form.Password, confirmPassword), "password", "Passwords aren't identical")

		if !form.Valid() {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]any{
				"error":  "Validation failed",
				"fields": form.FieldErrors,
			})
			return
		}

		// Avatar Upload
		file, header, err := r.FormFile("avatar")
		if err != nil {
			if err == http.ErrMissingFile {
				form.Avatar = "default-avatar.jpg"
			} else {
				f.ErrorLog.Printf("error reading avatar file: %v", err)
				form.Avatar = "default-avatar.jpg"
			}
		} else {
			defer file.Close()
			if header == nil || header.Filename == "" {
				form.Avatar = "default-avatar.jpg"
			} else {
				form.Avatar, err = app.UploadImage(file, *header, "avatars")
				if err != nil {
					log.Println(err)
					form.Avatar = "default-avatar.jpg"
				}
			}
		}

		uuid, err := f.Users.Register(form.Username, form.Email, form.Password, form.Avatar, 0)
		if err != nil {
			if errors.Is(err, models.ErrDuplicateRecord) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusConflict)
				json.NewEncoder(w).Encode(map[string]string{
					"error": "An account with this email or username already exists",
				})
				return
			}
			f.ErrorLog.Printf("User registration failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		err = f.Sessions.CreateSession(w, r, uuid)
		if err != nil {
			f.ErrorLog.Printf("Session creation failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("New user registered with UUID: %s", uuid)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "User registered successfully", "user_id": uuid})
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

		// Handle POST
		emailOrUsername := r.PostForm.Get("username")
		password := r.PostForm.Get("password")

		id, err := f.Users.Authenticate(emailOrUsername, password)
		if err != nil {
			f.ErrorLog.Printf("User authentication failed: %v", err)
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		// Create a session for the user.
		err = f.Sessions.CreateSession(w, r, id)
		if err != nil {
			f.ErrorLog.Printf("Session creation failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		f.InfoLog.Printf("User with ID %s logged in successfully", id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Login successful", "user_id": id})
	}
}

func Logout(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := f.Sessions.DestroySession(w, r)
		if err != nil {
			f.ErrorLog.Printf("%v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Logged out successfully"})
	}
}

func GoogleLogin(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := auth.GoogleConfig.GetAuthURL("state-token") // In production, use a random state
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func GoogleCallback(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := auth.GoogleConfig.Exchange(code)
		if err != nil {
			f.ErrorLog.Printf("Google exchange error: %v", err)
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		userInfo, err := auth.GoogleConfig.GetUserInfo(token)
		if err != nil {
			f.ErrorLog.Printf("Google user info error: %v", err)
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		handleOAuthLogin(w, r, f, userInfo)
	}
}

func GitHubLogin(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		url := auth.GitHubConfig.GetAuthURL("state-token")
		http.Redirect(w, r, url, http.StatusTemporaryRedirect)
	}
}

func GitHubCallback(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		code := r.URL.Query().Get("code")
		token, err := auth.GitHubConfig.Exchange(code)
		if err != nil {
			f.ErrorLog.Printf("GitHub exchange error: %v", err)
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		userInfo, err := auth.GitHubConfig.GetUserInfo(token)
		if err != nil {
			f.ErrorLog.Printf("GitHub user info error: %v", err)
			http.Error(w, "Failed to get user info", http.StatusInternalServerError)
			return
		}

		handleOAuthLogin(w, r, f, userInfo)
	}
}

func handleOAuthLogin(w http.ResponseWriter, r *http.Request, f *app.Application, userInfo *auth.UserInfo) {
	// Check if user exists
	user, err := f.Users.GetByEmail(userInfo.Email)
	var userID string

	if err != nil {
		if errors.Is(err, models.ErrNoRecords) {
			// Create new user
			userID, err = f.Users.CreateOAuthUser(userInfo.Name, userInfo.Email, userInfo.Picture)
			if err != nil {
				f.ErrorLog.Printf("Failed to create OAuth user: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		} else {
			f.ErrorLog.Printf("Database error: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	} else {
		userID = user.ID
	}

	// Create session
	err = f.Sessions.CreateSession(w, r, userID)
	if err != nil {
		f.ErrorLog.Printf("Session creation failed: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}
