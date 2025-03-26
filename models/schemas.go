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
