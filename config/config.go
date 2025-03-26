package config

import (
	"context"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Config holds all configuration for the application
type Config struct {
	MongoDB struct {
		URI        string
		Database   string
		Collection string
	}
	Server struct {
		Port string
	}
	JWT struct {
		Secret string
	}
}

// LoadConfig loads configuration from environment variables or defaults
func LoadConfig() *Config {
	config := &Config{}

	config.MongoDB.URI = os.Getenv("DB_URI")
	config.MongoDB.Database = os.Getenv("DB_DATABASE")
	config.Server.Port = os.Getenv("SERVER_PORT")
	config.JWT.Secret = "your-secret-key" // Use environment variable in production

	return config
}

// ConnectDatabase establishes a connection to MongoDB
func ConnectDatabase(config *Config) *mongo.Database {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(config.MongoDB.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Connected to MongoDB!")
	return client.Database(config.MongoDB.Database)
}
