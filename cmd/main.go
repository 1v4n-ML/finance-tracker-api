package main

import (
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
	router.Run(":" + AppConfig.Server.Port)
}
