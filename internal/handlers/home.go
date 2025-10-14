package handlers

import (
	"forum/internal/app"
	"log"
	"net/http"
)

func Home(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if f.InfoLog != nil {
			f.InfoLog.Println("home page")
		}
		test, err := f.Users.Register("Cudder", "bla@bla.col", "fsdfdsf", "avatar.jpg", 1)
		if err != nil {
			log.Println(err)
		}
		log.Println(test)
		w.Write([]byte("Welcome to the Forum!"))
	}
}

func RegisterPage(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		render(w, r, f, "register.html")
	}
}
