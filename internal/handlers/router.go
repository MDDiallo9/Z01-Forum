package handlers

import (
	"forum/internal/app"
	"forum/internal/middleware"
	"net/http"
)

// Register routes and all handlerson ServeMux; mount static, compose middleware

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*") // Allow any origin
		next.ServeHTTP(w, r)
	})
}
func Routes(f *app.Application) *http.ServeMux {
	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir("./ui/templates/static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Create an instance of the authentication middleware
	// Then pass f.Sessions because it meets the SessionManager perequisites
	auth := middleware.AuthRequired(f.Sessions, f.Users)
	maybeAuth := middleware.LoadSession(f.Sessions, f.Users)
	adminOnly := func(next http.Handler) http.Handler {
		return auth(middleware.RequireAdmin(next))
	}

	mux.Handle("GET /{$}", maybeAuth(Home(f)))
	mux.Handle("GET /categories/{id}", maybeAuth(CategoryPosts(f)))
	mux.Handle("GET /register", maybeAuth(RegisterPage(f)))
	mux.Handle("POST /register", maybeAuth(Register(f)))
	mux.Handle("GET /login", maybeAuth(LoginPage(f)))
	mux.Handle("POST /login", maybeAuth(Login(f)))

	// API Routes that carry data from database to the frontend
	mux.Handle("GET /api/posts", enableCORS(ListPosts(f)))
	mux.Handle("GET /api/categories", enableCORS(ListCategories(f)))
	mux.Handle("GET /api/categories/{id}/posts", enableCORS(ListPostsByCategory(f)))

	// PROTECTED ROUTES ALL GO HERE FOLLOWING THE PATTERN
	// Protected handler to test our sessions
	mux.Handle("GET /post/create", auth(CreatePostPage(f)))
	mux.Handle("POST /post/create", auth(CreatePost(f)))
	mux.Handle("GET /post/{id}", auth(GetPost(f)))
	mux.Handle("DELETE /post/delete/{id}", auth(DeletePost(f)))
	mux.Handle("PUT /post/update/{id}", auth(UpdatePost(f)))
	mux.Handle("GET /logout", auth(LogoutPopUp(f)))
	mux.Handle("POST /logout", auth(Logout(f)))

	// Report Creation Routes for only logged in users
	mux.Handle("POST /posts/{id}/report", auth(CreateReport(f)))
	mux.Handle("GET /posts/{id}/report", auth(CreateReportPage(f)))

	// Likes and Comments
	mux.Handle("POST /post/{id}/like", auth(LikePost(f)))
	mux.Handle("POST /post/{id}/dislike", auth(DislikePost(f)))
	mux.Handle("POST /comment/{id}/like", auth(LikeComment(f)))
	mux.Handle("POST /comment/{id}/dislike", auth(DislikeComment(f)))
	mux.Handle("POST /post/{id}/comment", auth(CreateComment(f)))

	// User Profile
	mux.Handle("GET /user/posts", auth(UserCreatedPosts(f)))
	mux.Handle("GET /user/liked", auth(UserLikedPosts(f)))

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
