package mongodb

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/services/usersvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type mongoUserRepo struct {
	usersCollection *mongo.Collection
}

func NewUserRepo(cfg opendoorchat.Config, cl *mongo.Client) *mongoUserRepo {
	usersCollection := cl.Database(cfg.Mongo.Database).Collection("users")
	if usersCollection == nil {
		log.Fatalln("users collection does not exist")
	}

	return &mongoUserRepo{
		usersCollection: usersCollection,
	}
}

func (repo *mongoUserRepo) CreateUser(ctx context.Context, user usersvc.User) httputil.HttpError {
	if err := user.Validate(); err != nil {
		return httputil.HttpErrorFromErr(err)
	}
	user.Id = ""
	user.FirstName = fixCasing(user.FirstName)
	user.LastName = fixCasing(user.LastName)
	user.Email = strings.ToLower(user.Email)
	user.CreatedAt = time.Now()
	_, err := repo.usersCollection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return httputil.NewHttpError(409, "", err.Error())
		}
		return httputil.HttpErrorFromErr(err)
	}
	return nil
}

func (repo *mongoUserRepo) GetUser(
	ctx context.Context,
	id string,
) (usersvc.User, httputil.HttpError) {
	if id == "" {
		return usersvc.User{}, httputil.NewHttpError(
			http.StatusBadRequest,
			"required id is blank",
			"",
		)
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return usersvc.User{}, httputil.NewHttpError(
			http.StatusBadRequest,
			"invalid id",
			err.Error(),
		)
	}
	res := repo.usersCollection.FindOne(ctx, bson.M{"_id": objectId})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return usersvc.User{}, httputil.NewHttpError(http.StatusNotFound, "", res.Err().Error())
		}
		return usersvc.User{}, httputil.HttpErrorFromErr(res.Err())
	}
	var user usersvc.User
	if err := res.Decode(&user); err != nil {
		return usersvc.User{}, httputil.HttpErrorFromErr(res.Err())
	}
	return user, nil
}

func (repo *mongoUserRepo) SearchUser(
	ctx context.Context,
	st usersvc.UserSearchTerms,
) (usersvc.User, httputil.HttpError) {
	//
	var orValues []bson.M
	if st.Email != "" {
		orValues = append(orValues, bson.M{"email": strings.ToLower(st.Email)})
	}

	//
	res := repo.usersCollection.FindOne(ctx, bson.M{
		"$or": orValues,
	})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return usersvc.User{}, httputil.NewHttpError(
				http.StatusNotFound,
				"",
				res.Err().Error(),
			)
		}
		return usersvc.User{}, httputil.HttpErrorFromErr(res.Err())
	}

	//
	var user usersvc.User
	if err := res.Decode(&user); err != nil {
		return usersvc.User{}, httputil.HttpErrorFromErr(err)
	}
	return user, nil
}

func fixCasing(name string) string {
	var res string
	if len(name) > 0 {
		res += strings.ToUpper(string(name[0]))
	}
	if len(name) > 1 {
		res += strings.ToLower(string(name[1:]))
	}
	return res
}
