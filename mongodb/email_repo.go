package mongodb

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/emailsvc"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Ensure interface is implemented
var _ emailsvc.EmailRepo = (*mongoEmailRepo)(nil)

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
) (emailsvc.EmailThread, opendoorchat.Error) {
	const op = "mongoEmailRepo.ThreadSearch"
	var orValues []bson.M
	if st.ChatId != "" {
		id, err := primitive.ObjectIDFromHex(st.ChatId)
		if err != nil {
			return emailsvc.EmailThread{}, opendoorchat.NewErr(400, "invalid ChatId", "")
		}
		orValues = append(orValues, bson.M{"chatId": id})
	}
	if st.EmailMessageId != "" {
		orValues = append(orValues, bson.M{"emails.messageId": st.EmailMessageId})
	}
	res := repo.emailThreadsCollection.FindOne(ctx, bson.M{
		"$or": orValues,
	})
	var thread emailsvc.EmailThread
	err := res.Decode(&thread)
	if err == mongo.ErrNoDocuments {
		return emailsvc.EmailThread{}, opendoorchat.NewErr(404, "", "")
	} else if err != nil {
		return emailsvc.EmailThread{}, opendoorchat.FromErr(err, fmt.Sprintf("%s: FindOne", op))
	}
	return thread, nil
}

func (repo *mongoEmailRepo) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email emailsvc.Email,
) opendoorchat.Error {
	const op = "mongoEmailRepo.AddEmail"
	email.SentAt = time.Now()
	res := repo.emailThreadsCollection.FindOneAndUpdate(ctx, bson.M{"_id": threadId}, bson.M{
		"$push": bson.M{
			"emails": email,
		},
	}, options.FindOneAndUpdate().SetProjection(bson.M{"_id": -1}))
	if res.Err() != nil {
		if res.Err() == mongo.ErrNoDocuments {
			return opendoorchat.NewErr(404, "", "")
		}
		return opendoorchat.FromErr(res.Err(), fmt.Sprintf("%s: FindOneAndUpdate", op))
	}
	return nil
}
