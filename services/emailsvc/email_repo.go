package emailsvc

import (
	"context"
	"time"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/benjamonnguyen/opendoorchat/services/usersvc"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailRepo interface {
	ThreadSearch(context.Context, ThreadSearchTerms) (EmailThread, httputil.HttpError)
	AddEmail(context.Context, primitive.ObjectID, Email) httputil.HttpError
}

type Email struct {
	MessageId string    `json:"messageId,omitempty" bson:"messageId"`
	SentAt    time.Time `json:"sentAt,omitempty"    bson:"sentAt"`
}

type EmailThread struct {
	Id           primitive.ObjectID `json:"id,omitempty"        bson:"_id"`
	Participants []usersvc.User     `                           bson:"participants"`
	Emails       []Email            `json:"emails,omitempty"    bson:"emails"`
	ChatId       primitive.ObjectID `json:"chatId,omitempty"    bson:"chatId"`
	CreatedAt    time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
}

type ThreadSearchTerms struct {
	ChatId         string `json:"chatId,omitempty"`
	EmailMessageId string `json:"emailMessageId,omitempty"`
}
