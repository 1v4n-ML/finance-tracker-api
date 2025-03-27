package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/1v4n-ML/finance-tracker-api/config"
	"github.com/1v4n-ML/finance-tracker-api/models"
	"github.com/1v4n-ML/finance-tracker-api/utils"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AccountController struct {
	db  *mongo.Database
	col *mongo.Collection
	cfg *config.Config
}

func NewAccountController(db *mongo.Database, cfg *config.Config) *AccountController {
	return &AccountController{
		db:  db,
		col: db.Collection("accounts"),
		cfg: cfg,
	}
}

func (ac *AccountController) GetAllAccounts(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), ac.cfg.Timeouts.Request)
	defer cancel()

	cursor, err := ac.col.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	defer func(cursor *mongo.Cursor, ctz context.Context) {
		err := cursor.Close(ctx)
		if err != nil {
		}
	}(cursor, ctx)

	var accounts []models.Account
	if err = cursor.All(ctx, &accounts); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, &accounts)
}

func (ac *AccountController) GetAccountById(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), ac.cfg.Timeouts.Request)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var account models.Account

	err = ac.col.FindOne(ctx, bson.M{"_id": id}).Decode(&account)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, account)
}

func (ac *AccountController) CreateAccount(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), ac.cfg.Timeouts.Request)
	defer cancel()

	var account models.Account

	if err := c.ShouldBindJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.ID = primitive.NewObjectID()
	account.CreatedAt = time.Now()
	account.Balance = 0

	result, err := ac.col.InsertOne(ctx, account)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

func (ac *AccountController) UpdateAccount(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), ac.cfg.Timeouts.Request)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var account models.Account
	if err := c.ShouldBindBodyWithJSON(&account); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	account.UpdatedAt = time.Now()
	_, err = ac.col.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": account},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account updated"})
}

func (ac *AccountController) DeleteAccount(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), ac.cfg.Timeouts.Request)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	_, err = ac.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "account deleted"})
}
