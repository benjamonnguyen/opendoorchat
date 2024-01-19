package emailsvc

import (
	"context"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailRepo interface {
	ThreadSearch(context.Context, ThreadSearchTerms) (EmailThread, app.Error)
	AddEmail(context.Context, primitive.ObjectID, Email) app.Error
}

type Email struct {
	MessageId string    `json:"messageId,omitempty" bson:"messageId"`
	SentAt    time.Time `json:"sentAt,omitempty"    bson:"sentAt"`
}

type EmailThread struct {
	Id           primitive.ObjectID `json:"id,omitempty"        bson:"_id"`
	Participants []app.User         `                           bson:"participants"`
	Emails       []Email            `json:"emails,omitempty"    bson:"emails"`
	ChatId       primitive.ObjectID `json:"chatId,omitempty"    bson:"chatId"`
	CreatedAt    time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
}

type ThreadSearchTerms struct {
	ChatId         string `json:"chatId,omitempty"`
	EmailMessageId string `json:"emailMessageId,omitempty"`
}
