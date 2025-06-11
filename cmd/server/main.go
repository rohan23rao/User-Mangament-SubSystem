// cmd/server/main.go
package main

import (
	"fmt"
	"log"
	"userms/internal/config"
	"userms/internal/database"
	"userms/internal/server"
	"userms/internal/utils"
)

func main() {
	// Print startup banner
	utils.PrintStartupBanner()

	// Load configuration
	cfg := config.Load()

	// Initialize database
	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Create and start server
	srv := server.New(cfg, db)
	
	utils.LogInfo("Server configuration:")
	utils.LogInfo("  Port: %s", cfg.Port)
	utils.LogInfo("  Kratos Public URL: %s", cfg.KratosPublicURL)
	utils.LogInfo("  Kratos Admin URL: %s", cfg.KratosAdminURL)

	fmt.Printf("\n%s%s🌟 Server ready! Listening on: http://localhost:%s %s\n\n",
		utils.ColorBold, utils.ColorGreen, cfg.Port, utils.ColorReset)
	
	utils.LogSuccess("Server starting on port %s", cfg.Port)
	
	if err := srv.Start(); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}
