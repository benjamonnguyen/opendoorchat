package repo

import (
	"context"
	"log"

	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserRepo interface {
	GetUser(context.Context, string) (model.User, error)
}

type mongoUserRepo struct {
	usersCollection *mongo.Collection
}

func NewMongoUserRepo(cfg config.Config, cl *mongo.Client) *mongoUserRepo {
	usersCollection := cl.Database(cfg.Mongo.Database).Collection("users")
	if usersCollection == nil {
		log.Fatalln("users collection does not exist")
	}

	return &mongoUserRepo{
		usersCollection: usersCollection,
	}
}

func (repo *mongoUserRepo) GetUser(ctx context.Context, id string) (model.User, error) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.User{}, status.Error(codes.InvalidArgument, "invalid id")
	}
	res := repo.usersCollection.FindOne(ctx, bson.M{"_id": objectId})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return model.User{}, status.Error(codes.NotFound, res.Err().Error())
		}
		return model.User{}, status.Error(codes.Internal, res.Err().Error())
	}
	var user model.User
	if err := res.Decode(&user); err != nil {
		return model.User{}, status.Error(codes.Internal, err.Error())
	}
	return user, nil
}
