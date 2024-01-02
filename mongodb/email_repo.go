package mongodb

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/services/emailsvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type mongoEmailRepo struct {
	emailThreadsCollection *mongo.Collection
}

func NewEmailRepo(cfg opendoorchat.Config, cl *mongo.Client) *mongoEmailRepo {
	emailThreadsCollection := cl.Database(cfg.Mongo.Database).Collection("emailThreads")
	if emailThreadsCollection == nil {
		log.Fatalln("emailThreads collection does not exist")
	}

	return &mongoEmailRepo{
		emailThreadsCollection: emailThreadsCollection,
	}
}

func (repo *mongoEmailRepo) ThreadSearch(
	ctx context.Context,
	st emailsvc.ThreadSearchTerms,
) (emailsvc.EmailThread, httputil.HttpError) {
	var orValues []bson.M
	if st.ChatId != "" {
		id, err := primitive.ObjectIDFromHex(st.ChatId)
		if err != nil {
			return emailsvc.EmailThread{}, httputil.NewHttpError(
				http.StatusBadRequest,
				"invalid ChatId",
				"",
			)
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
			return emailsvc.EmailThread{}, httputil.NewHttpError(
				http.StatusNotFound,
				"",
				res.Err().Error(),
			)
		}
		return emailsvc.EmailThread{}, httputil.HttpErrorFromErr(res.Err())
	}
	var thread emailsvc.EmailThread
	if err := res.Decode(&thread); err != nil {
		return emailsvc.EmailThread{}, httputil.HttpErrorFromErr(err)
	}
	return thread, nil
}

func (repo *mongoEmailRepo) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email emailsvc.Email,
) httputil.HttpError {
	email.SentAt = time.Now()
	res := repo.emailThreadsCollection.FindOneAndUpdate(ctx, bson.M{"_id": threadId}, bson.M{
		"$push": bson.M{
			"emails": email,
		},
	}, options.FindOneAndUpdate().SetProjection(bson.M{"_id": -1}))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return httputil.NewHttpError(
				http.StatusNotFound,
				"",
				res.Err().Error(),
			)
		}
		return httputil.HttpErrorFromErr(res.Err())
	}
	return nil
}
