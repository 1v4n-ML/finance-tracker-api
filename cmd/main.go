package main

import (
	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/routes"
	"github.com/joho/godotenv"
)

func main() {
	//Load .env variables
	godotenv.Load("/home/minhoca-manca/Go_dev/finance-tracker-api/.env")

	// Initialize configuration
	cfg := config.LoadConfig()

	// Setup database connection
	db := config.ConnectDatabase(cfg)

	// Setup router with routes
	router := routes.SetupRouter(db)

	// Start server
	router.Run(":" + cfg.Server.Port)
}
