// https://www.keycloak.org/docs/latest/securing_apps/
package keycloak

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
)

type AuthClient struct {
	cl           *http.Client
	cfg          Config
	serviceToken string
}

func NewAuthClient(cl *http.Client, cfg Config) *AuthClient {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tkn, _ := requestServiceToken(ctx, cl, cfg)
	//
	return &AuthClient{
		cl:           cl,
		cfg:          cfg,
		serviceToken: tkn,
	}
}

func (cl *AuthClient) LogOut(ctx context.Context, refreshToken string) app.Error {
	const (
		op   = "AuthClient.LogOut"
		path = "/realms/opendoor-chat/protocol/openid-connect/logout"
	)

	// build body
	data := url.Values{}
	data.Add("refresh_token", refreshToken)
	data.Add("client_id", cl.cfg.ClientId)
	data.Add("client_secret", cl.cfg.ClientSecret)

	// build request
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.cfg.BaseUrl+path,
		strings.NewReader(data.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// logout
	resp, err := cl.cl.Do(req)
	if err != nil {
		return app.FromErr(err, op)
	}
	if resp.StatusCode != 204 {
		return app.NewErr(resp.StatusCode, resp.Status, op)
	}
	return nil
}

func (cl *AuthClient) Introspect(
	ctx context.Context,
	accessToken string,
) (Introspection, app.Error) {
	const (
		op   = "AuthClient.Introspect"
		path = "/realms/opendoor-chat/protocol/openid-connect/token/introspect"
	)

	// build data
	data := url.Values{}
	data.Add("token_type_hint", "access_token")
	data.Add("token", accessToken)
	data.Add("client_id", cl.cfg.ClientId)
	data.Add("client_secret", cl.cfg.ClientSecret)

	// build request
	req, _ := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.cfg.BaseUrl+path,
		strings.NewReader(data.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-url-form-urlencoded")

	// introspect
	resp, err := cl.cl.Do(req)
	if err != nil {
		return Introspection{}, app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return Introspection{}, app.NewErr(resp.StatusCode, resp.Status, op)
	}

	// decode
	var res Introspection
	json.NewDecoder(resp.Body).Decode(&res)
	return res, nil
}

type Introspection struct {
	// 	"scope": "email profile",
	//   "sid": "3c9871b7-58f0-4886-9213-11ed238b0209",
	//   "active": true
	UserInfo
	Scope  string `json:"scope,omitempty"`
	Active bool   `json:"active,omitempty"`
}

// RequestAccessToken returns accessToken and optional refreshToken or else Error
func (cl *AuthClient) RequestAccessToken(
	ctx context.Context, refreshToken, username, password string,
) (string, string, app.Error) {
	const op = "AuthClient.RequestAccessToken"
	// build payload
	data := url.Values{}
	data.Add("client_id", cl.cfg.ClientId)
	data.Add("client_secret", cl.cfg.ClientSecret)
	if refreshToken != "" {
		data.Add("refresh_token", refreshToken)
		data.Add("grant_type", "refresh_token")
	} else if username != "" && password != "" {
		data.Add("username", username)
		data.Add("password", password)
		data.Add("grant_type", "password")
	} else {
		return "", "", app.NewErr(400, "missing username/password or refresh_token", op)
	}

	//
	return requestAccessToken(ctx, cl.cl, cl.cfg, data)
}

func requestAccessToken(
	ctx context.Context,
	cl *http.Client,
	cfg Config,
	data url.Values,
) (string, string, app.Error) {
	const (
		op   = "AuthClient.requestAccessToken"
		path = "/realms/opendoor-chat/protocol/openid-connect/token"
	)
	// devlog.Printf("%s: data: %s", op, data.Encode())
	// build request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cfg.BaseUrl+path,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// get tokens
	resp, err := cl.Do(req)
	if err != nil {
		return "", "", app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return "", "", app.NewErr(resp.StatusCode, resp.Status, op)
	}
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&body)
	return body.AccessToken, body.RefreshToken, nil
}

func requestServiceToken(
	ctx context.Context,
	cl *http.Client,
	cfg Config,
) (string, app.Error) {
	data := url.Values{}
	data.Add("client_id", cfg.ClientId)
	data.Add("client_secret", cfg.ClientSecret)
	data.Add("grant_type", "client_credentials")
	token, _, err := requestAccessToken(ctx, cl, cfg, data)
	return token, err
}
