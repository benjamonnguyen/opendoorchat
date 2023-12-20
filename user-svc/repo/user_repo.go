package repo

import (
	"context"
	"log"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type UserRepo interface {
	GetUser(ctx context.Context, id string) (model.User, httputil.HttpError)
	SearchUser(context.Context, model.UserSearchTerms) (model.User, httputil.HttpError)
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

func (repo *mongoUserRepo) GetUser(
	ctx context.Context,
	id string,
) (model.User, httputil.HttpError) {
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return model.User{}, httputil.NewHttpError(http.StatusBadRequest, "invalid id", err.Error())
	}
	res := repo.usersCollection.FindOne(ctx, bson.M{"_id": objectId})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return model.User{}, httputil.NewHttpError(http.StatusNotFound, "", res.Err().Error())
		}
		return model.User{}, httputil.HttpErrorFromErr(res.Err())
	}
	var user model.User
	if err := res.Decode(&user); err != nil {
		return model.User{}, httputil.HttpErrorFromErr(res.Err())
	}
	return user, nil
}

func (repo *mongoUserRepo) SearchUser(
	ctx context.Context,
	st model.UserSearchTerms,
) (model.User, httputil.HttpError) {
	//
	var orValues []bson.M
	if st.Email != "" {
		orValues = append(orValues, bson.M{"email": st.Email})
	}

	//
	res := repo.usersCollection.FindOne(ctx, bson.M{
		"$or": orValues,
	})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return model.User{}, httputil.NewHttpError(
				http.StatusNotFound,
				"",
				res.Err().Error(),
			)
		}
		return model.User{}, httputil.HttpErrorFromErr(res.Err())
	}

	//
	var user model.User
	if err := res.Decode(&user); err != nil {
		return model.User{}, httputil.HttpErrorFromErr(err)
	}
	return user, nil
}
