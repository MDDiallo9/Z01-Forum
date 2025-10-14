package handlers

import (
	"forum/internal/app"
	"net/http"
	"log"
)

func Home(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        if f.InfoLog != nil {
            f.InfoLog.Println("home page")
        }
		test,err := f.Users.Register("Cudder","bla@bla.col","fsdfdsf","avatar.jpg",1)
		if err != nil {
			log.Println(err)
		}
		log.Println(test)
        w.Write([]byte("Welcome to the Forum!"))
    }
}