package models

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	dbFile := "./forum.db"
	// Check if the database file exists
	_, err := os.Stat(dbFile)
	needsInit := os.IsNotExist(err)

	DB, err = sql.Open("sqlite3", dbFile)
	if err != nil {
		log.Fatal("Failed to open database:", err)
		return nil, err
	}

	if needsInit {
		log.Println("Database file not found, creating and initializing schema...")
		// Read and execute the schema to create tables
		schemaPath := "migrations/schema.sql"
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			log.Printf("FATAL: Could not read schema file to initialize database: %v", err)
			return nil, err
		}

		_, err = DB.Exec(string(schema))
		if err != nil {
			log.Printf("FATAL: Failed to execute schema.sql: %v", err)
			return nil, err
		}
		log.Println("Database initialized successfully.")
	} else {
		log.Println("Existing database found.")
	}

	if _, err := DB.Exec("PRAGMA foreign_keys = ON;"); err != nil {
		log.Printf("Warning: Could not enable foreign key constraints: %v", err)
	}

	return DB, nil
}
