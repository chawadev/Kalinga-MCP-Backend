package database

import (
	"context"
	"log"
	"time"

	"github.com/chawadev/kalinga-backend/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client
var Database *mongo.Database
var cfg *config.Config

func Connect() error {
	cfg = config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var err error
	Client, err = mongo.Connect(ctx, options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return err
	}

	// Ping the database to verify connection
	if err := Client.Ping(ctx, nil); err != nil {
		return err
	}

	Database = Client.Database(cfg.DatabaseName)
	log.Println("Connected to MongoDB")
	return nil
}

func Disconnect() error {
	if Client != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return Client.Disconnect(ctx)
	}
	return nil
}
