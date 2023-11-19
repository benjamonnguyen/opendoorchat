package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Email struct {
	MessageId string `json:"messageId,omitempty"`
}

type EmailThread struct {
	Id primitive.ObjectID `json:"id,omitempty"        bson:"_id"`
	// TODO Participants []*userpb.User     `bson:"participants,omitempty"`
	Emails    []Email            `json:"emails,omitempty"`
	ChatId    primitive.ObjectID `json:"chatId,omitempty"`
	CreatedAt time.Time          `json:"createdAt,omitempty"`
}

type ThreadSearchTerms struct {
	ChatId         string `json:"chatId,omitempty"`
	EmailMessageId string `json:"emailMessageId,omitempty"`
}
