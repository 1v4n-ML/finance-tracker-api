package main

import (
	"log"

	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/routes"
	"github.com/joho/godotenv"
)

var AppConfig *config.Config // package-level variable

func main() {
	//Load .env variables
	godotenv.Load("/home/minhoca-manca/Go_dev/finance-tracker-api/.env")

	// Initialize configuration
	AppConfig = config.LoadConfig()

	// Setup database connection
	db := config.ConnectDatabase(AppConfig)

	// Setup router with routes
	router := routes.SetupRouter(db, AppConfig)

	// Start server
	// Listen on all interfaces (0.0.0.0) on the specified port
	listenAddr := "0.0.0.0:" + AppConfig.Server.Port
	log.Printf("Starting server on %s", listenAddr)
	err := router.Run(listenAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
