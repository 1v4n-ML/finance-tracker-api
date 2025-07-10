package main

import (
	"context"
	"log"

	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/routes"
	"github.com/1v4n-ML/finance-tracker-api/utils"
	"github.com/joho/godotenv"
	"gopkg.in/robfig/cron.v2"
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

	//Setup scheduler
	c := cron.New()
	_, err := c.AddFunc("*/3 * * * *", func() {
		utils.RecalculateAllBalancesService(db, context.Background())
	})
	if err != nil {
		log.Fatalf("catastrophic failure when starting CRON task")
	}
	c.Start()
	log.Println("starting scheduler to run at every 3 min")

	// Start server
	// Listen on all interfaces (0.0.0.0) on the specified port
	listenAddr := "0.0.0.0:" + AppConfig.Server.Port
	log.Printf("Starting server on %s", listenAddr)
	err = router.Run(listenAddr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
