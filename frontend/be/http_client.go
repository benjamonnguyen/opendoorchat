package be

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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

const (
	AccessTokenHeaderKey = "X-Access-Token"
)

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

func (cl *Client) Authenticate(
	ctx context.Context,
	email, password string,
) (string, app.Error) {
	const op = "backendClient.SignIn"

	// encode args
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).
		Encode(fmt.Sprintf(`{"email": %s, "password": %s}`, email, password))
	if err != nil {
		return "", app.FromErr(err, op)
	}

	// build request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.baseUrl+"/user/authenticate",
		buf,
	)
	if err != nil {
		return "", app.FromErr(err, op)
	}
	if err := addAccessTokenHeader(req); err != nil {
		return "", app.FromErr(err, op)
	}

	// TODO reCAPTCHA
	// req.Header.Set("Authorization", a.cfg.BackendApiKey)
	resp, err := cl.cl.Do(req)
	if err != nil {
		return "", app.FromErr(err, op)
	}
	if resp.StatusCode == 200 {
		token, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", app.FromErr(err, op)
		}
		return string(token), nil
	}
	return "", app.NewErr(resp.StatusCode, resp.Status, "")
}

func (cl *Client) CreateUser(
	ctx context.Context,
	user User,
) app.Error {
	const op = "backendClient.SignUp"
	// TODO client side validations must be re-done on the server side, as they can always be bypassed.

	// encode args
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).
		Encode(user)
	if err != nil {
		return app.FromErr(err, op)
	}

	// build request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.baseUrl+"/user",
		buf,
	)
	if err != nil {
		return app.FromErr(err, op)
	}
	if err := addAccessTokenHeader(req); err != nil {
		return app.FromErr(err, op)
	}

	// get response
	resp, err := cl.cl.Do(req)
	if err != nil {
		return app.FromErr(err, op)
	}
	if resp.StatusCode == 201 {
		return nil
	}
	return app.NewErr(resp.StatusCode, resp.Status, "")
}

func addAccessTokenHeader(req *http.Request) error {
	// token, err := req.Cookie(AccessTokenCookieKey)
	// if err != nil {
	// 	return err
	// }

	// req.Header.Add(AccessTokenHeaderKey, token.Value)
	return nil
}
