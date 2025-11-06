package main

import (
	"forum/internal/app"
	"forum/internal/handlers"
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

	// Initialize services(SessionManager)
	sessionManager := &services.SessionManager{
		DB:         db,
		CookieName: "forum_session",
		LifeTime:   1 * time.Hour,
		HardMax:    24 * time.Hour,
	}

	forum := app.NewApplication(info, errLog, usersModel, postsModel, categoriesModel, attachmentsModel, reportsModel, sessionManager)
	mux := handlers.Routes(forum)
	srv := app.Server(forum, mux)

	// Server configuration

	log.Printf("Starting Forum server on http://localhost%s\n", srv.Addr)

	err = srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}
