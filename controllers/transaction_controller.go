package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/1v4n-ML/finance-tracker-api/models"
	"github.com/1v4n-ML/finance-tracker-api/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type TransactionController struct {
	db  *mongo.Database
	col *mongo.Collection
}

func NewTransactionController(db *mongo.Database) *TransactionController {
	return &TransactionController{db: db, col: db.Collection("transactions")}
}

// GetAll returns all transactions
func (tc *TransactionController) GetAll(c *gin.Context) {
	ctx := context.Background()

	startDate, err := utils.ParseDateToISO(c.Query("start_date"))
	if err != nil {
		//dunno what to do
	}
	endDate, err := utils.ParseDateToISO(c.Query("end_date"))
	if err != nil {
		//dunno what to do
	}

	var cursor *mongo.Cursor
	if startDate.IsZero() && endDate.IsZero() {
		cursor, err = tc.col.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query fucked up smh"})
			return
		}
	} else {
		cursor, err = tc.col.Find(ctx, bson.M{"date": bson.M{"$gte": startDate, "$lte": endDate}})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "query fucked up smh"})
			return
		}
	}

	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var transactions []models.Transaction
	if err = cursor.All(ctx, &transactions); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, transactions)
}

// GetByID returns a single transaction by ID
func (tc *TransactionController) GetByID(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var transaction models.Transaction
	ctx := context.Background()

	err = tc.col.FindOne(ctx, bson.M{"_id": id}).Decode(&transaction)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Transaction not found"})
		return
	}

	c.JSON(http.StatusOK, transaction)
}

// Create adds a new transaction
func (tc *TransactionController) Create(c *gin.Context) {
	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()

	transaction.ID = primitive.NewObjectID()
	transaction.CreatedAt = time.Now()
	result, err := tc.col.InsertOne(ctx, transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

// Update modifies an existing transaction
func (tc *TransactionController) Update(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var transaction models.Transaction
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx := context.Background()
	transaction.UpdatedAt = time.Now()
	_, err = tc.col.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": transaction},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction updated"})
}

// Delete removes a transaction
func (tc *TransactionController) Delete(c *gin.Context) {
	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	ctx := context.Background()

	_, err = tc.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Transaction deleted"})
}
