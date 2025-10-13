package main

import (
	"forum/internals/models"
	"log"
	"net/http"
)

func main() {
	// Initialize DB and schema
	db, err := models.InitDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	log.Println("✓ Database initialized successfully!")

	// Create a new ServeMux (router)
	mux := http.NewServeMux()

	// Static files - serves CSS, JS, images from ./static directory
	fileServer := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static", fileServer))

	// Routes
	mux.HandleFunc("/", home)

	// Server configuration
	port := ":8080"
	log.Printf("Starting Forum server on http://localhost%s\n", port)

	// Start server (note: using different variable name to avoid redeclaration)
	serverErr := http.ListenAndServe(port, mux)
	if serverErr != nil {
		log.Fatal("Server failed to start:", serverErr)
	}
}

func home(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Welcome to the Forum!"))
}
