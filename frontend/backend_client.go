package frontend

import (
	"context"
	"net/http"
)

type BackendClient interface {
	Authenticate(ctx context.Context, email, password string) (*http.Response, error)
	CreateUser(context.Context, User) (*http.Response, error)
	// CreateChat(context.Context, Chat) (*http.Response, error)
	// AddChatMessage(context.Context, ChatMessage) (*http.Response, error)
}

type User struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Password  string `json:"password"`
}

type Chat struct {
	OwnerId        string
	ParticipantIds []string
	Subject        string
}

type ChatMessage struct {
	Text   string
	SentAt int64
}

const (
	AccessTokenCookieKey = "OPENDOOR_CHAT_TOKEN"
	AccessTokenHeaderKey = "X-Access-Token"
)
