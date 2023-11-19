package db

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type MongoConfig struct {
	URI      string
	Database string
}

func ConnectMongoClient(ctx context.Context, cfg MongoConfig) *mongo.Client {
	opts := options.Client().ApplyURI(cfg.URI).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalln("failed mongo.Connect:", err)
	}
	return c
}
