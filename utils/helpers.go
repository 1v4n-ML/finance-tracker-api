package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/1v4n-ML/finance-tracker-api/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func ParseDateToISO(date string) (time.Time, error) {
	var layout = os.Getenv("DATE_LAYOUT")
	d, err := time.Parse(layout, date)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed parsing date: %v", err)
	}
	return d, nil
}

func NewContextWithTimeout(baseCtx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if baseCtx == nil {
		baseCtx = context.Background()
	}
	return context.WithTimeout(baseCtx, timeout)
}

func MapOperator(op string) (string, error) {
	switch op {
	case "eq":
		return "$eq", nil
	case "ne":
		return "$ne", nil
	case "gt":
		return "$gt", nil
	case "gte":
		return "$gte", nil
	case "lt":
		return "$lt", nil
	case "lte":
		return "$lte", nil
	case "in":
		return "$in", nil
	case "nin":
		return "$nin", nil
	default:
		return "", fmt.Errorf("unsupported operator: %s", op)
	}
}

func ParseFilterValue(field string, value interface{}, operator string) (interface{}, error) {
	// Handle fields expecting ObjectID
	if field == "category_id" || field == "account_id" || field == "_id" {
		// Handle 'in'/'nin' operators which expect an array
		if operator == "in" || operator == "nin" {
			valSlice, ok := value.([]interface{})
			if !ok {
				return nil, fmt.Errorf("value for '%s' with operator '%s' must be an array of strings", field, operator)
			}
			objIDs := make([]primitive.ObjectID, 0, len(valSlice))
			for _, v := range valSlice {
				strVal, ok := v.(string)
				if !ok {
					return nil, fmt.Errorf("value element for '%s' must be a string ObjectID", field)
				}
				objID, err := primitive.ObjectIDFromHex(strVal)
				if err != nil {
					return nil, fmt.Errorf("invalid ObjectID string '%s' for field '%s': %w", strVal, field, err)
				}
				objIDs = append(objIDs, objID)
			}
			return objIDs, nil
		} else {
			// Handle single ObjectID
			strVal, ok := value.(string)
			if !ok {
				return nil, fmt.Errorf("value for '%s' must be a string ObjectID", field)
			}
			objID, err := primitive.ObjectIDFromHex(strVal)
			if err != nil {
				return nil, fmt.Errorf("invalid ObjectID string '%s' for field '%s': %w", strVal, field, err)
			}
			return objID, nil
		}
	}

	// Handle date field
	if field == "date" {
		strVal, ok := value.(string)
		if !ok {
			return nil, fmt.Errorf("value for 'date' field must be an ISO 8601 date string")
		}
		// Try parsing with different layouts if necessary
		t, err := time.Parse(time.RFC3339, strVal)
		if err != nil {
			// Add more layouts if needed, e.g., "2006-01-02"
			t, err = time.Parse("2006-01-02", strVal)
			if err != nil {
				return nil, fmt.Errorf("invalid date format for 'date' field (expected RFC3339 or YYYY-MM-DD): %w", err)
			}
			// If only date is provided, adjust based on operator for inclusive ranges
			if operator == "gte" { // Start of the day
				t = time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
			} else if operator == "lte" { // End of the day
				t = time.Date(t.Year(), t.Month(), t.Day(), 23, 59, 59, 999999999, t.Location())
			}
		}
		return t, nil
	}

	// Handle numeric fields (like amount)
	if field == "amount" {
		switch v := value.(type) {
		case float64:
			return v, nil
		case float32:
			return float64(v), nil
		case int:
			return float64(v), nil
		case int64:
			return float64(v), nil
		case string: // Allow string representation of numbers
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				return nil, fmt.Errorf("invalid numeric string '%s' for field '%s': %w", v, field, err)
			}
			return f, nil
		default:
			return nil, fmt.Errorf("unsupported type for numeric field '%s'", field)
		}
	}

	// Handle string fields (like type, description)
	// Allow 'in'/'nin' for string fields too
	if operator == "in" || operator == "nin" {
		valSlice, ok := value.([]interface{})
		if !ok {
			return nil, fmt.Errorf("value for '%s' with operator '%s' must be an array of strings", field, operator)
		}
		strVals := make([]string, 0, len(valSlice))
		for _, v := range valSlice {
			strVal, ok := v.(string)
			if !ok {
				return nil, fmt.Errorf("value element for '%s' must be a string", field)
			}
			strVals = append(strVals, strVal)
		}
		return strVals, nil
	}

	// Default: assume it's a simple type like string, bool etc.
	return value, nil
}

// Helper function to parse a timeout duration from environment variable (in milliseconds)
func ParseTimeout(envVar string, defaultValue time.Duration) time.Duration {
	timeoutStr := os.Getenv(envVar) // Get value, might be empty
	timeout := defaultValue         // Start with the default

	if timeoutStr != "" {
		parsedMs, err := strconv.Atoi(timeoutStr)
		if err != nil {
			// Log a warning but use the default
			log.Printf("WARNING: Invalid format for %s ('%s'). Using default %v. Error: %v",
				envVar, timeoutStr, defaultValue, err)
		} else if parsedMs <= 0 {
			// Log a warning if non-positive value is provided
			log.Printf("WARNING: Non-positive value for %s ('%d'). Using default %v.",
				envVar, parsedMs, defaultValue)
		} else {
			// Successfully parsed a positive value
			timeout = time.Duration(parsedMs) * time.Millisecond
		}
	} else {
		// Optional: Log if you want to know when defaults are used
		// log.Printf("INFO: Environment variable %s not set. Using default %v.", envVar, defaultValue)
	}
	return timeout
}

// Helper function to get environment variable or return default
func GetEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func UpdateAccountBalanceOnTransaction(ctx context.Context, db *mongo.Database, transaction models.Transaction, changeFactor float64) error {
	if transaction.Account.IsZero() {
		return errors.New("transação não possui uma conta associada")
	}

	// Define o valor da mudança: positivo para income, negativo para expense
	changeAmount := transaction.Amount
	if transaction.Type == "expense" {
		changeAmount = -transaction.Amount
	}

	// Aplica o fator de mudança (para tratar deleções)
	finalChange := changeAmount * changeFactor

	accountsCollection := db.Collection("accounts")
	_, err := accountsCollection.UpdateOne(
		ctx,
		bson.M{"_id": transaction.Account},
		bson.M{"$inc": bson.M{"balance": finalChange}},
	)

	if err != nil {
		return fmt.Errorf("falha ao atualizar saldo da conta %s: %w", transaction.Account.Hex(), err)
	}

	return nil
}

func RecalculateAllBalancesService(db *mongo.Database, ctx context.Context) {
	log.Println("Iniciando tarefa agendada: Recalculando todos os saldos...")

	accountsCol := db.Collection("accounts")
	transactionsCol := db.Collection("transactions")

	// 1. Zera o saldo de todas as contas
	_, err := accountsCol.UpdateMany(ctx, bson.M{}, bson.M{"$set": bson.M{"balance": 0}})
	if err != nil {
		log.Printf("ERRO no scheduler: Falha ao zerar saldos: %v", err)
		return
	}

	// 2. Busca todas as transações
	cursor, err := transactionsCol.Find(ctx, bson.M{})
	if err != nil {
		log.Printf("ERRO no scheduler: Falha ao buscar transações: %v", err)
		return
	}
	defer cursor.Close(ctx)

	// 3. Itera sobre cada transação e atualiza o saldo da conta correspondente
	for cursor.Next(ctx) {
		var trans models.Transaction
		if err := cursor.Decode(&trans); err != nil {
			log.Printf("ERRO no scheduler: Falha ao decodificar transação durante recalculo: %v", err)
			continue
		}

		// Reutiliza a função que já temos!
		if err := UpdateAccountBalanceOnTransaction(ctx, db, trans, 1.0); err != nil {
			log.Printf("ERRO no scheduler: Falha ao atualizar saldo para transação %s: %v", trans.ID.Hex(), err)
		}
	}

	if err := cursor.Err(); err != nil {
		log.Printf("ERRO no scheduler: Erro no cursor de transações: %v", err)
	}

	log.Println("Tarefa finalizada: Saldos recalculados com sucesso.")
}
