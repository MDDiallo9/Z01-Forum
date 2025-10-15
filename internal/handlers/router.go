package handlers

import (
	"forum/internal/app"
	"net/http"
)

// Register routes and all handlerson ServeMux; mount static, compose middleware

func Routes(f *app.Application) *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/templates/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	mux.HandleFunc("GET /{$}", Home(f))
	mux.HandleFunc("GET /register", RegisterPage(f))
	mux.HandleFunc("POST /register", Register(f))
	mux.HandleFunc("GET /login", LoginPage(f))
	mux.HandleFunc("POST /login", Login(f))
	return mux
}
