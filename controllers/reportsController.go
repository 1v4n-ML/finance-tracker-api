package controllers

import (
	"log"

	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/models"
	"github.com/1v4n-ML/finance-tracker-api/services"
	"github.com/1v4n-ML/finance-tracker-api/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ReportsController struct {
	db  *mongo.Database
	col *mongo.Collection
	cfg *config.Config
}

func NewReportsController(db *mongo.Database, cfg *config.Config) *ReportsController {
	return &ReportsController{
		db:  db,
		col: db.Collection("reports"),
		cfg: cfg,
	}
}

func (rc *ReportsController) AggregateTransactions(c *gin.Context) {
	var req models.AggregationRequest

	// Bind JSON request body
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	// Build the aggregation pipeline
	pipeline, err := services.BuildAggregationPipeline(req)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to build aggregation pipeline: " + err.Error()})
		return
	}

	// Log the generated pipeline (optional, for debugging)
	log.Printf("Executing aggregation pipeline: %+v\n", pipeline)

	// Execute the aggregation query
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), rc.cfg.Timeouts.Request)
	defer cancel()

	// Assuming transactionCollection is initialized
	cursor, err := rc.col.Aggregate(ctx, pipeline, options.Aggregate())
	if err != nil {
		log.Printf("MongoDB aggregation error: %v", err)
		c.JSON(500, gin.H{"error": "Failed to execute aggregation query"})
		return
	}
	defer cursor.Close(ctx)

	// Decode results into a slice of maps (flexible for dynamic results)
	var results []bson.M // or []map[string]interface{}
	if err = cursor.All(ctx, &results); err != nil {
		log.Printf("MongoDB cursor decoding error: %v", err)
		c.JSON(500, gin.H{"error": "Failed to decode aggregation results"})
		return
	}

	// Handle case where results might be nil from MongoDB if no documents matched
	if results == nil {
		results = []bson.M{}
	}

	c.JSON(200, results)
}
