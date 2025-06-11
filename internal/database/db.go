// internal/database/db.go
package database

import (
	"database/sql"
	"fmt"
	"time"
	"userms/internal/utils"

	_ "github.com/lib/pq"
)

func Connect(databaseURL string) (*sql.DB, error) {
	utils.LogDB("Connecting to PostgreSQL database...")
	utils.LogDB("Database URL: %s", utils.MaskPassword(databaseURL))

	// Retry connection with backoff
	var db *sql.DB
	var err error
	for i := 0; i < 30; i++ {
		db, err = sql.Open("postgres", databaseURL)
		if err != nil {
			utils.LogError("Failed to open database connection: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		// Test the connection
		if err = db.Ping(); err != nil {
			utils.LogWarning("Database not ready, retrying in 2 seconds... (attempt %d/30)", i+1)
			time.Sleep(2 * time.Second)
			continue
		}

		utils.LogSuccess("Connected to PostgreSQL database")
		break
	}

	if err != nil {
		return nil, fmt.Errorf("failed to connect to database after 30 attempts: %v", err)
	}

	// Set connection pool settings
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test if our tables exist
	var count int
	err = db.QueryRow("SELECT COUNT(*) FROM information_schema.tables WHERE table_name = 'users'").Scan(&count)
	if err != nil {
		utils.LogError("Failed to check if tables exist: %v", err)
		return nil, err
	}

	if count == 0 {
		utils.LogWarning("Tables don't exist yet - they should be created by init.sql")
	} else {
		utils.LogDB("Database tables verified")
	}

	return db, nil
}
