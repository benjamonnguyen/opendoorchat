package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/benjamonnguyen/opendoorchat/httputil"
)

type AuthClient struct {
	cl           *http.Client
	cfg          frontend.KeycloakCfg
	serviceToken string
}

func NewAuthClient(cl *http.Client, cfg frontend.KeycloakCfg) *AuthClient {
	self := &AuthClient{
		cl:  cl,
		cfg: cfg,
	}
	//
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	self.repopulateServiceToken(ctx)
	//
	return self
}

type UserRepresentation struct {
	Enabled         bool                       `json:"enabled,omitempty"`
	Attributes      map[string]string          `json:"attributes,omitempty"`
	Credentials     []CredentialRepresentation `json:"credentials,omitempty"`
	Email           string                     `json:"email,omitempty"`
	EmailVerified   bool                       `json:"emailVerified,omitempty"`
	FirstName       string                     `json:"firstName,omitempty"`
	LastName        string                     `json:"lastName,omitempty"`
	RequiredActions []string                   `json:"requiredActions,omitempty"`
}

type CredentialRepresentation struct {
	Type      string `json:"type,omitempty"` // password
	Value     string `json:"value,omitempty"`
	Temporary bool   `json:"temporary,omitempty"`
}

func (cl *AuthClient) RegisterUser(ctx context.Context, usr UserRepresentation) app.Error {
	const (
		op   = "authClient.RegisterUser"
		path = "/admin/realms/opendoor-chat/users"
	)

	// build request
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(usr)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, cl.cfg.BaseUrl+path, buf)
	req.Header.Add("Content-Type", "application/json")

	// register user
	_, err := httputil.DoWithRetries[interface{}](func() (any, app.Error) {
		req.Header.Add("Authorization", "Bearer "+cl.serviceToken)
		resp, err := cl.cl.Do(req)
		if err != nil {
			return nil, app.FromErr(err, op)
		}
		if resp.StatusCode != 201 {
			if resp.StatusCode == 401 {
				cl.repopulateServiceToken(ctx)
			}
			return nil, app.NewErr(resp.StatusCode, resp.Status, op)
		}
		return nil, nil
	},
		2,
		func(code int) bool { return code == 401 },
		httputil.ExponentialBackoffConfigs{},
	)
	return err
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
	return cl.requestAccessToken(ctx, data)
}

func (cl *AuthClient) requestAccessToken(
	ctx context.Context,
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
		cl.cfg.BaseUrl+path,
		strings.NewReader(data.Encode()),
	)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// get tokens
	resp, err := cl.cl.Do(req)
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

func (cl *AuthClient) repopulateServiceToken(ctx context.Context) app.Error {
	data := url.Values{}
	data.Add("client_id", cl.cfg.ClientId)
	data.Add("client_secret", cl.cfg.ClientSecret)
	data.Add("grant_type", "client_credentials")
	token, _, err := cl.requestAccessToken(ctx, data)
	cl.serviceToken = token
	return err
}

// "userinfo_endpoint":"http://localhost:9090/realms/opendoor-chat/protocol/openid-connect/userinfo",
// "introspection_endpoint":"http://localhost:9090/realms/opendoor-chat/protocol/openid-connect/token/introspect"
