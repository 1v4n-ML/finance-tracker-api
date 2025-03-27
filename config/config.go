package config

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/1v4n-ML/finance-tracker-api/utils"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --- Timeout Configuration ---
const (
	// Environment variable names for timeouts (in milliseconds)
	dbTimeoutEnvVar     = "TIMEOUT_MS_DATABASE"
	reportTimeoutEnvVar = "TIMEOUT_MS_REQUEST"

	// Default timeout durations
	defaultDbTimeout      = 5 * time.Second  // Default for standard DB operations
	defaultRequestTimeout = 30 * time.Second // Default for reports/aggregations
	defaultServerPort     = "8080"           // Default server port
	connectTimeout        = 10 * time.Second // Timeout for initial DB connection
)

// Config holds all configuration for the application
type Config struct {
	MongoDB struct {
		URI      string
		Database string
		// Collection string // Note: Collection name might be better handled per-repository/service
	}
	Server struct {
		Port string
	}
	Timeouts struct { // New struct for timeouts
		Database time.Duration
		Request  time.Duration
	}
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	config := &Config{}

	// --- MongoDB Configuration (Required) ---
	config.MongoDB.URI = os.Getenv("DB_URI")
	if config.MongoDB.URI == "" {
		log.Fatal("FATAL: Environment variable DB_URI is not set.")
	}
	config.MongoDB.Database = os.Getenv("DB_DATABASE")
	if config.MongoDB.Database == "" {
		log.Fatal("FATAL: Environment variable DB_DATABASE is not set.")
	}

	// --- Server Configuration ---
	config.Server.Port = utils.GetEnvOrDefault("SERVER_PORT", defaultServerPort)

	// --- Timeout Configuration ---
	config.Timeouts.Database = utils.ParseTimeout(dbTimeoutEnvVar, defaultDbTimeout)
	config.Timeouts.Request = utils.ParseTimeout(reportTimeoutEnvVar, defaultRequestTimeout)

	log.Printf("Loaded configuration: Port=%s, DB=%s", config.Server.Port, config.MongoDB.Database)
	log.Printf("Loaded timeouts: DB=%v, Request=%v", config.Timeouts.Database, config.Timeouts.Request)

	return config
}

// ConnectDatabase establishes a connection to MongoDB
func ConnectDatabase(config *Config) *mongo.Database {
	// This timeout is specifically for the initial connection attempt
	ctx, cancel := context.WithTimeout(context.Background(), connectTimeout)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		// Use Fatalf for formatted error messages
		log.Fatalf("FATAL: Failed to connect to MongoDB at %s: %v", config.MongoDB.URI, err)
	}

	// Ping the primary to verify connection is established.
	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatalf("FATAL: Failed to ping MongoDB: %v", err)
	}

	log.Println("Successfully connected to MongoDB!")
	return client.Database(config.MongoDB.Database)
}
