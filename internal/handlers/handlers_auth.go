package handlers

import (
	"errors"
	"forum/internal/app"
	"forum/internal/models"
	"net/http"
)

type userRegistrationForm struct {
	Username    string
	Password    string
	Email       string
	FieldErrors map[string]string
	app.Validator
}

func Register(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := r.ParseForm()
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

		// TODO: Add real avatar support.
		uuid, err := f.Users.Register(form.Username, form.Email, form.Password, "default-avatar.jpg", 1)
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

		f.InfoLog.Printf("New user registered with UUID: %s", uuid)
		w.Write([]byte("Registration successful!"))
	}
}
