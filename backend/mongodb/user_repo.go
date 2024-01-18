package mongodb

import (
	"context"
	"log"
	"strings"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/backend"
	"github.com/benjamonnguyen/opendoorchat/backend/usersvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var _ usersvc.UserRepo = (*mongoUserRepo)(nil)

type mongoUserRepo struct {
	usersCollection *mongo.Collection
}

func NewUserRepo(cfg backend.Config, cl *mongo.Client) *mongoUserRepo {
	usersCollection := cl.Database(cfg.Mongo.Database).Collection("users")
	if usersCollection == nil {
		log.Fatalln("users collection does not exist")
	}

	return &mongoUserRepo{
		usersCollection: usersCollection,
	}
}

func (repo *mongoUserRepo) CreateUser(ctx context.Context, user usersvc.User) app.Error {
	const op = "mongoUserRepo.CreateUser"
	if err := user.Validate(); err != nil {
		return app.FromErr(err, op)
	}
	user.Id = ""
	user.FirstName = fixCasing(user.FirstName)
	user.LastName = fixCasing(user.LastName)
	user.Email = strings.ToLower(user.Email)
	user.CreatedAt = time.Now()
	_, err := repo.usersCollection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return app.NewErr(409, "", "")
		}
		return app.FromErr(err, op)
	}
	return nil
}

func (repo *mongoUserRepo) GetUser(
	ctx context.Context,
	id string,
) (usersvc.User, app.Error) {
	const op = "mongoUserRepo.GetUser"
	if id == "" {
		return usersvc.User{}, app.NewErr(400, "required id is blank", "")
	}
	objectId, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return usersvc.User{}, app.NewErr(400, "invalid id", "")
	}
	res := repo.usersCollection.FindOne(ctx, bson.M{"_id": objectId})
	var user usersvc.User
	if err := res.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return usersvc.User{}, app.NewErr(404, "", "")
		}
		return usersvc.User{}, app.FromErr(err, op)
	}
	return user, nil
}

func (repo *mongoUserRepo) SearchUser(
	ctx context.Context,
	st usersvc.UserSearchTerms,
) (usersvc.User, app.Error) {
	const op = "mongoUserRepo.SearchUser"
	//
	var orValues []bson.M
	if st.Email != "" {
		orValues = append(orValues, bson.M{"email": strings.ToLower(st.Email)})
	}

	//
	res := repo.usersCollection.FindOne(ctx, bson.M{
		"$or": orValues,
	})
	var user usersvc.User
	if err := res.Decode(&user); err != nil {
		if err == mongo.ErrNoDocuments {
			return usersvc.User{}, app.NewErr(404, "", "")
		}
		return usersvc.User{}, app.FromErr(err, op)
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
