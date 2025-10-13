package models

import (
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB() (*sql.DB, error) {
	var err error
	DB, err = sql.Open("sqlite3", "./forum.db")
	if err != nil {
		log.Fatal("Failed to open database:", err)
	}

	// Then execute data.sql to add sample data (optional)
	if _, err := os.Stat("schema.sql"); err == nil {
		log.Println("Adding sample data...")
		data, err := os.ReadFile("schema.sql")
		if err != nil {
			log.Println("Warning: Failed to read schema.sql:", err)
			return DB, nil
		}

		_, err = DB.Exec(string(data))
		if err != nil {
			log.Println("Warning: Failed to execute data.sql:", err)
			log.Println("This might be normal if data already exists")
			return DB, nil
		}
		log.Println("Sample data added successfully")
	} else {
		log.Println("No data.sql found, skipping sample data")
	}
	return DB, nil
}
