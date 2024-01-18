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

	// register user
	_, err := httputil.DoWithRetries[interface{}](func() (any, app.Error) {
		resp, err := cl.cl.Do(req)
		if err != nil {
			return nil, app.FromErr(err, op)
		}
		if resp.StatusCode != 201 {
			if resp.StatusCode == 401 {
				cl.repopulateServiceToken(ctx)
			}
			return nil, app.NewErr(resp.StatusCode, resp.Status, "")
		}
		return nil, nil
	},
		2,
		func(code int) bool { return code == 401 },
		httputil.ExponentialBackoffConfigs{},
	)
	return err
}

// RequestAccessToken returns accessToken and optional refreshToken or else Error
func (cl *AuthClient) RequestAccessToken(
	ctx context.Context,
	clientId, clientSecret, refreshToken, username, password string,
) (string, string, app.Error) {
	const (
		op   = "authClient.RequestAccessToken"
		path = "/realms/opendoor-chat/protocol/openid-connect/token"
	)
	// build body
	data := url.Values{}
	data.Add("client_id", clientId)
	data.Add("client_secret", clientSecret)
	if refreshToken != "" {
		data.Add("refresh_token", refreshToken)
		data.Add("grant_type", "refresh_token")
	} else if username != "" {
		data.Add("username", username)
		data.Add("password", password)
		data.Add("grant_type", "password")
	} else {
		data.Add("grant_type", "client_credentials")
	}

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

	// get tokens
	resp, err := cl.cl.Do(req)
	if err != nil {
		return "", "", app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return "", "", app.NewErr(resp.StatusCode, resp.Status, "")
	}
	var body struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}
	json.NewDecoder(resp.Body).Decode(&body)
	return body.AccessToken, body.RefreshToken, nil
}

func (cl *AuthClient) repopulateServiceToken(ctx context.Context) app.Error {
	token, _, err := cl.RequestAccessToken(ctx, cl.cfg.ClientId, cl.cfg.ClientSecret, "", "", "")
	cl.serviceToken = token
	return err
}
