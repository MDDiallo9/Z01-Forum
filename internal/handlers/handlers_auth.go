package handlers

import (
	"errors"
	"forum/internal/app"
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
		form.CheckField(app.MinChars(form.Password, 6), "password", "Minimum 8 characters")
		form.CheckField(app.IsIdentical(form.Password, confirmPassword), "password", "Passwords aren't identical")

		if !form.Valid() {
			form.FieldErrors = form.Validator.FieldErrors
			data := &app.TemplateData{Form: form}
			render(w, r, f, "register.html", data)
			return
		}

		// Avatar Upload

		file, header, err := r.FormFile("avatar")
		if err != nil {
			// no file uploaded -> use default avatar
			if err == http.ErrMissingFile {
				form.Avatar = "default-avatar.jpg"
			} else {
				// real error reading the uploaded file
				f.ErrorLog.Printf("error reading avatar file: %v", err)
				form.Avatar = "default-avatar.jpg"
			}
		} else {
			defer file.Close()
			// user left the file input empty => filename may be empty
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
				// TODO Could add a box for this error instead of sticking it to a form field
				form.AddFieldError("email", "An account with this email or username already exists")
				form.AddFieldError("username", "An account with this email or username already exists")
				form.FieldErrors = form.Validator.FieldErrors
				data := &app.TemplateData{Form: form}
				render(w, r, f, "register.html", data)
				return
			}

			f.ErrorLog.Printf("User registration failed: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		// TODO Session Management and Cookie

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
		// w.Write([]byte("Login successful!"))

		// Redirect to dashboard
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}
