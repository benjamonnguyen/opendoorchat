package model

import (
	"time"

	usermodel "github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Email struct {
	MessageId string    `json:"messageId,omitempty" bson:"messageId"`
	SentAt    time.Time `json:"sentAt,omitempty"    bson:"sentAt"`
}

type EmailThread struct {
	Id           primitive.ObjectID `json:"id,omitempty"        bson:"_id"`
	Participants []usermodel.User   `                           bson:"participants"`
	Emails       []Email            `json:"emails,omitempty"    bson:"emails"`
	ChatId       primitive.ObjectID `json:"chatId,omitempty"    bson:"chatId"`
	CreatedAt    time.Time          `json:"createdAt,omitempty" bson:"createdAt"`
}

type ThreadSearchTerms struct {
	ChatId         string `json:"chatId,omitempty"`
	EmailMessageId string `json:"emailMessageId,omitempty"`
}
