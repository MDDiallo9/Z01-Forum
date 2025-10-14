package handlers

import (
	"forum/internal/app"
	"net/http"
)

func Home(f *app.Application) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
        if f.InfoLog != nil {
            f.InfoLog.Println("home page")
        }
        w.Write([]byte("Welcome to the Forum!"))
    }
}