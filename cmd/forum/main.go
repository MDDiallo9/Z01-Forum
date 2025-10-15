package main

import (
	"forum/internal/app"
	"forum/internal/handlers"
	"forum/internal/models"
	"log"
	"os"
)

func main() {
	// Initialize DB and schema
	db, err := models.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	info := log.New(os.Stdout, "INFO\t", log.Ldate|log.Ltime)
	errLog := log.New(os.Stderr, "ERROR\t", log.Ldate|log.Ltime)
	usersModel := &models.UsersModel{DB: db}

	forum := app.NewApplication(info, errLog, usersModel)
	mux := handlers.Routes(forum)
	srv := app.Server(forum, mux)

	// Server configuration

	log.Printf("Starting Forum server on http://localhost%s\n", srv.Addr)

	err = srv.ListenAndServe()
	app.ErrorLog.Fatal(err)
}
