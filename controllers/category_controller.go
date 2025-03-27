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

type CategoryController struct {
	db  *mongo.Database
	col *mongo.Collection
	cfg *config.Config
}

func NewCategoryController(db *mongo.Database, cfg *config.Config) *CategoryController {
	return &CategoryController{
		db:  db,
		col: db.Collection("categories"),
		cfg: cfg,
	}
}

// GetAllCategories returns all categories
func (cc *CategoryController) GetAllCategories(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), cc.cfg.Timeouts.Request)
	defer cancel()

	cursor, err := cc.col.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func(cursor *mongo.Cursor, ctx context.Context) {
		err := cursor.Close(ctx)
		if err != nil {

		}
	}(cursor, ctx)

	var categories []models.Category
	if err = cursor.All(ctx, &categories); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// CreateCategory adds a new category
func (cc *CategoryController) CreateCategory(c *gin.Context) {
	var category models.Category
	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), cc.cfg.Timeouts.Request)
	defer cancel()

	category.ID = primitive.NewObjectID()
	category.CreatedAt = time.Now()
	result, err := cc.col.InsertOne(ctx, category)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": result.InsertedID})
}

func (cc *CategoryController) UpdateCategory(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), cc.cfg.Timeouts.Request)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	var category models.Category
	if err := c.ShouldBindBodyWithJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	category.UpdatedAt = time.Now()
	_, err = cc.col.UpdateOne(
		ctx,
		bson.M{"_id": id},
		bson.M{"$set": category},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category updated"})
}

func (cc *CategoryController) DeleteCategory(c *gin.Context) {
	ctx, cancel := utils.NewContextWithTimeout(c.Request.Context(), cc.cfg.Timeouts.Request)
	defer cancel()

	id, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}

	_, err = cc.col.DeleteOne(ctx, bson.M{"_id": id})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "category deleted"})
}
