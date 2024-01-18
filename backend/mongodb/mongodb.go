package mongodb

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/benjamonnguyen/opendoorchat/backend"
)

func ConnectMongoClient(ctx context.Context, cfg backend.MongoConfig) *mongo.Client {
	opts := options.Client().ApplyURI(cfg.URI).
		SetServerAPIOptions(options.ServerAPI(options.ServerAPIVersion1))
	c, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatalln("failed mongo.Connect:", err)
	}
	return c
}
