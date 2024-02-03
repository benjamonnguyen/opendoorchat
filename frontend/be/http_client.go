package be

import (
	"net/http"

	app "github.com/benjamonnguyen/opendoorchat"
)

// TODO all users and auth stuff is handled by auth client

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

type Client struct {
	cl      *http.Client
	baseUrl string
}

func NewClient(cl *http.Client, baseUrl string) *Client {
	return &Client{
		cl:      cl,
		baseUrl: baseUrl,
	}
}

func (cl *Client) CreateChat(chat Chat, accessToken string) app.Error {

}

func addAccessTokenHeader(req *http.Request) error {
	// token, err := req.Cookie(AccessTokenCookieKey)
	// if err != nil {
	// 	return err
	// }

	// req.Header.Add(AccessTokenHeaderKey, token.Value)
	return nil
}
