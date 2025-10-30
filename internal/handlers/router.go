package handlers

import (
	"forum/internal/app"
	"forum/internal/middleware"
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
	auth := middleware.AuthRequired(f.Sessions, f.Users)
	adminOnly := func(next http.Handler) http.Handler {
		return auth(middleware.RequireAdmin(next))
	}

	// PROTECTED ROUTES ALL GO HERE FOLLOWING THE PATTERN
	// Protected handler to test our sessions
	mux.Handle("GET /post/create", auth(CreatePostPage(f)))
	mux.Handle("POST /post", auth(CreatePost(f)))
	mux.Handle("DELETE /post/delete/{id}", auth(DeletePost(f)))
	mux.Handle("PUT /post/update/{id}", auth(UpdatePost(f)))
	mux.Handle("GET /logout", auth(LogoutPopUp(f)))
	mux.Handle("POST /logout", auth(Logout(f)))

	// Report Creation Routes for only logged in users
	mux.Handle("POST /posts/{id}/report", auth(CreateReport(f)))
	mux.Handle("GET /posts/{id}/report", auth(CreateReportPage(f)))

	// ADMIN ONLY ROUTES
	// User Management
	// We use PUT for the updates
	mux.Handle("PUT /admin/users/{id}/promote", adminOnly(PromoteUser(f)))
	mux.Handle("PUT /admin/users/{id}/demote", adminOnly(DemoteUser(f)))

	// Report Management
	mux.Handle("GET /admin/reports", adminOnly(ListReports(f)))
	mux.Handle("POST /admin/reports/{id}/resolve", adminOnly(ResolveReport(f))) // Changed to POST for simplicity from a form/link

	return mux
}
