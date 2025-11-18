package main

import (
	"forum/internal/app"
	"forum/internal/handlers"
	"forum/internal/middleware"
	"forum/internal/models"
	"forum/internal/services"
	"log"
	"os"
	"time"
)

func main() {
	// Initialize DB and schema
	db, err := models.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Initialize models
	info := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
	usersModel := &models.UsersModel{DB: db}
	postsModel := &models.PostsModel{DB: db}
	categoriesModel := &models.CategoriesModel{DB: db}
	attachmentsModel := &models.AttachmentsModel{DB: db}
	reportsModel := &models.ReportsModel{DB: db}
	likesModel := &models.LikesModel{DB: db}
	commentsModel := &models.CommentsModel{DB: db}
	notificationsModel := &models.NotificationsModel{DB: db}

	// Initialize services(SessionManager)
	sessionManager := &services.SessionManager{
		DB:         db,
		CookieName: "forum_session",
		LifeTime:   1 * time.Hour,
		HardMax:    24 * time.Hour,
	}

	forum := app.NewApplication(info, errLog, usersModel, postsModel, categoriesModel, attachmentsModel, reportsModel, likesModel, commentsModel, notificationsModel, sessionManager)
	mux := handlers.Routes(forum)
	// Middleware Chain
	// We wrap the mux with our middleware
	handler := middleware.SecureHeaders(mux)
	handler = middleware.RateLimit(handler)

	srv := app.Server(forum, handler)

	// Server configuration

	log.Printf("Starting Forum server on https://localhost%s\n", srv.Addr)

	// Generate self-signed certs if they don't exist:
	// openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem -days 365 -nodes -subj "/C=US/ST=Denial/L=Springfield/O=Dis/CN=localhost"

	err = srv.ListenAndServeTLS("cert.pem", "key.pem")
	if err != nil {
		// Fallback to HTTP if certs are missing (for development convenience, or error out)
		log.Printf("Failed to start HTTPS server: %v. Falling back to HTTP.", err)
		srv.Addr = ":8080" // Ensure port is correct for HTTP if needed, or keep same
		log.Printf("Starting Forum server on http://localhost%s\n", srv.Addr)
		err = srv.ListenAndServe()
		app.ErrorLog.Fatal(err)
	}
}
