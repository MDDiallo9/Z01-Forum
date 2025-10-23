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

	// Create an instance of the authentication middleware
	// Then pass f.Sessions because it meets the SessionManager perequisites
	// auth := middleware.AuthRequired(f.Sessions)

	// PROTECTED ROUTES ALL GO HERE FOLLOWING THE PATTERN
	// Protected handler to test our sessions
	// mux.Handle("GET /dashboard", auth(Dashboard(f)))

	return mux
}
