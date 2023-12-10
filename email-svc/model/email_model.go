package model

import (
	"time"

	usermodel "github.com/benjamonnguyen/opendoor-chat/user-svc/model"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Email struct {
	MessageId string `json:"messageId,omitempty"`
}

type EmailThread struct {
	Id           primitive.ObjectID `json:"id,omitempty"        bson:"_id"`
	Participants []usermodel.User   `                           bson:"participants,omitempty"`
	Emails       []Email            `json:"emails,omitempty"`
	ChatId       primitive.ObjectID `json:"chatId,omitempty"`
	CreatedAt    time.Time          `json:"createdAt,omitempty"`
}

type ThreadSearchTerms struct {
	ChatId         string `json:"chatId,omitempty"`
	EmailMessageId string `json:"emailMessageId,omitempty"`
}
