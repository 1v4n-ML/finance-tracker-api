package main

import (
	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/routes"
	"log"
)

func main() {
	// Initialize configuration
	cfg := config.LoadConfig()

	// Setup database connection
	db := config.ConnectDatabase(cfg)

	// Setup router with routes
	router := routes.SetupRouter(db)

	// Start server
	log.Printf("Starting server on port %s", cfg.Server.Port)
	router.Run(":" + cfg.Server.Port)
}
