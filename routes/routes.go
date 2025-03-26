// routes/routes.go
package routes

import (
	"github.com/1v4n-ML/finance-tracker-api/controllers"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"
)

// SetupRouter configures the API routes and returns the router
func SetupRouter(db *mongo.Database) *gin.Engine {
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"}, //this should be more restricted for prod
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Create controllers with database dependency
	transactionController := controllers.NewTransactionController(db)
	categoryController := controllers.NewCategoryController(db)
	accountsController := controllers.NewAccountController(db)

	// API routes - no authentication needed
	api := router.Group("/api")
	{
		// Transaction routes
		transactions := api.Group("/transactions")
		{
			transactions.GET("", transactionController.GetAll)
			transactions.GET("/:id", transactionController.GetByID)
			transactions.POST("", transactionController.Create)
			transactions.PUT("/:id", transactionController.Update)
			transactions.DELETE("/:id", transactionController.Delete)
		}

		// Category routes
		categories := api.Group("/categories")
		{
			categories.GET("", categoryController.GetAllCategories)
			categories.POST("", categoryController.CreateCategory)
			categories.PUT("/:id", categoryController.UpdateCategory)
			categories.DELETE("/:id", categoryController.DeleteCategory)
		}

		// Account routes
		accounts := api.Group("/accounts")
		{
			accounts.GET("", accountsController.GetAllAccounts)
			accounts.GET("/:id", accountsController.GetAccountById)
			accounts.POST("", accountsController.CreateAccount)
			accounts.PUT("/:id", accountsController.UpdateAccount)
			accounts.DELETE("/:id", accountsController.DeleteAccount)
		}
	}

	return router
}
