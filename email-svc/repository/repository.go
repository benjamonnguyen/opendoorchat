package repository

import (
	"context"
	"log"

	"github.com/benjamonnguyen/opendoor-chat-services/commons/config"
	"github.com/benjamonnguyen/opendoor-chat-services/email-svc/model"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type EmailRepo interface {
	ThreadSearch(context.Context, model.ThreadSearchTerms) (*model.EmailThread, error)
	AddEmail(context.Context, primitive.ObjectID, model.Email) error
}

type mongoEmailRepo struct {
	emailThreadsCollection *mongo.Collection
}

func NewMongoEmailRepo(cfg config.Config, cl *mongo.Client) *mongoEmailRepo {
	repo := &mongoEmailRepo{
		emailThreadsCollection: cl.Database(cfg.Mongo.Database).Collection("emailThreads"),
	}
	if repo.emailThreadsCollection == nil {
		log.Fatalln("emailThreads collection does not exist")
	}

	return repo
}

func (repo *mongoEmailRepo) ThreadSearch(
	ctx context.Context,
	st model.ThreadSearchTerms,
) (*model.EmailThread, error) {
	var orValues []bson.M
	if st.ChatId != "" {
		id, err := primitive.ObjectIDFromHex(st.ChatId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "invalid ChatId")
		}
		orValues = append(orValues, bson.M{"chatId": id})
	}
	if st.EmailMessageId != "" {
		orValues = append(orValues, bson.M{"emails.messageId": st.EmailMessageId})
	}
	res := repo.emailThreadsCollection.FindOne(ctx, bson.M{
		"$or": orValues,
	})
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return nil, status.Error(codes.NotFound, res.Err().Error())
		}
		return nil, status.Error(codes.Internal, res.Err().Error())
	}
	var thread model.EmailThread
	if err := res.Decode(&thread); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &thread, nil
}

func (repo *mongoEmailRepo) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email model.Email,
) error {
	res := repo.emailThreadsCollection.FindOneAndUpdate(ctx, bson.M{"_id": threadId}, bson.M{
		"$push": bson.M{
			"emails": email,
		},
	}, options.FindOneAndUpdate().SetProjection(bson.M{"_id": -1}))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return status.Error(codes.NotFound, res.Err().Error())
		}
		return status.Error(codes.Internal, res.Err().Error())
	}
	return nil
}
