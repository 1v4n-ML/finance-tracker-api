package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Transaction represents a financial transaction
type Transaction struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Amount      float64            `json:"amount" bson:"amount" binding:"required,min=0"`
	Date        time.Time          `json:"date" bson:"date" binding:"required"`
	Description string             `json:"description" bson:"description"`
	CategoryID  primitive.ObjectID `json:"category_id,omitempty" bson:"category_id,omitempty"`
	Type        string             `json:"type" bson:"type" binding:"required,oneof=income expense"` // income or expense
	Account     primitive.ObjectID `json:"account_id,omitempty" bson:"account_id,omitempty"`         // pix, credit card, etc.
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

// Category represents a transaction category
type Category struct {
	ID          primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name" binding:"required"`
	Description string             `json:"description,omitempty" bson:"description,omitempty"`
	Color       string             `json:"color,omitempty" bson:"color,omitempty"`                   // For UI representation
	Type        string             `json:"type" bson:"type" binding:"required,oneof=income expense"` // income or expense
	CreatedAt   time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at" bson:"updated_at"`
}

type Account struct {
	ID         primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Name       string             `json:"name" bson:"name" binding:"required"`
	Type       string             `json:"type" bson:"type" binding:"required,oneof=wallet bank credit_card"`
	Balance    float64            `json:"balance" bson:"balance"`
	Color      string             `json:"color,omitempty" bson:"color,omitempty"` // For UI representation
	ClosureDay int                `json:"closure_day" bson:"closure_day" binding:"omitempty,required_if=Type credit_card,gte=1,lte=31"`
	PayDay     int                `json:"payday" bson:"payday" binding:"omitempty,required_with=ClosureDay,gte=1,lte=31"`
	CreatedAt  time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at" bson:"updated_at"`
}

type Filter struct {
	Field    string      `json:"field" binding:"required"`
	Operator string      `json:"operator" binding:"required,oneof=eq ne gt gte lt lte in nin"`
	Value    interface{} `json:"value" binding:"required"` // Use interface{} to accept various types (string, number, array, date string)
}

type Metric struct {
	Name      string `json:"name" binding:"required"`
	Operation string `json:"operation" binding:"required,oneof=sum count avg"`
	Field     string `json:"field,omitempty"` // Optional, but required for sum/avg
}

type AggregationRequest struct {
	Filters []Filter       `json:"filters"`
	GroupBy []string       `json:"groupBy" binding:"required,min=1"`
	Metrics []Metric       `json:"metrics" binding:"required,min=1"`
	SortBy  map[string]int `json:"sortBy"` // Key: field name, Value: 1 (asc) or -1 (desc)
	Limit   *int64         `json:"limit"`  // Use pointer for optional field
	Offset  *int64         `json:"offset"` // Use pointer for optional field
}
