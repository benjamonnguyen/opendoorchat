package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/httputil"
)

type User struct {
	Enabled         bool                       `json:"enabled,omitempty"`
	Attributes      map[string]string          `json:"attributes,omitempty"`
	Credentials     []CredentialRepresentation `json:"credentials,omitempty"`
	Email           string                     `json:"email,omitempty"` // Email is always lowercased by Keycloak
	EmailVerified   bool                       `json:"emailVerified,omitempty"`
	FirstName       string                     `json:"firstName,omitempty"`
	LastName        string                     `json:"lastName,omitempty"`
	RequiredActions []string                   `json:"requiredActions,omitempty"`
}

var _ app.User = (*User)(nil)

type CredentialRepresentation struct {
	Type      string `json:"type,omitempty"` // "password"
	Value     string `json:"value,omitempty"`
	Temporary bool   `json:"temporary,omitempty"`
}

type UserInfo struct {
	// 	"sub": "b10c21d4-5c4c-4d19-b180-f29abc52124f",
	// 	"email_verified": false,
	// 	"name": "Jesse Pinkman",
	// 	"preferred_username": "captaincook@b.b",
	// 	"given_name": "Jesse",
	// 	"family_name": "Pinkman",
	// 	"email": "captaincook@b.b"
	Email         string `json:"email,omitempty"`
	EmailVerified bool   `json:"email_verified,omitempty"`
	FirstName     string `json:"given_name,omitempty"`
	LastName      string `json:"family_name,omitempty"`
}

var _ app.User = (*UserInfo)(nil)

type keycloakUserCl struct {
	cl           *http.Client
	cfg          Config
	serviceToken string
}

func NewUserRepo(cl *http.Client, cfg Config) *keycloakUserCl {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	tkn, _ := requestServiceToken(ctx, cl, cfg)
	return &keycloakUserCl{
		cl:           cl,
		cfg:          cfg,
		serviceToken: tkn,
	}
}

var _ app.UserRepo = (*keycloakUserCl)(nil)

func (r *keycloakUserCl) CreateUser(ctx context.Context, usr app.User) app.Error {
	const (
		op   = "keycloakUserCl.CreateUser"
		path = "/admin/realms/opendoor-chat/users"
	)

	// validate
	if err := usr.Validate(); err != nil {
		return app.FromErr(err, op)
	}

	// build request
	buf := &bytes.Buffer{}
	json.NewEncoder(buf).Encode(usr)
	req, _ := http.NewRequestWithContext(ctx, http.MethodPost, r.cfg.BaseUrl+path, buf)
	req.Header.Add("Content-Type", "application/json")

	// register user with retry if token is expired
	_, err := httputil.DoWithRetries[interface{}](func() (any, app.Error) {
		req.Header.Add("Authorization", "Bearer "+r.serviceToken)
		resp, err := r.cl.Do(req)
		if err != nil {
			return nil, app.FromErr(err, op)
		}
		if resp.StatusCode != 201 {
			if resp.StatusCode == 401 {
				tkn, err := requestServiceToken(req.Context(), r.cl, r.cfg)
				if err != nil {
					return nil, app.FromErr(err, op)
				}
				r.serviceToken = tkn
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

func (r *keycloakUserCl) GetUser(ctx context.Context, id string) (app.User, app.Error) {
	const (
		op   = "keycloakUserCl.GetUser"
		path = "/admin/realms/opendoor-chat/users/"
	)

	// validate
	if id == "" {
		return nil, app.NewErr(400, "required id is blank", op)
	}

	// build request
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, r.cfg.BaseUrl+path+id, nil)

	// get user
	resp, err := r.cl.Do(req)
	if err != nil {
		return nil, app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return nil, app.NewErr(resp.StatusCode, resp.Status, op)
	}

	//
	var usr User
	json.NewDecoder(resp.Body).Decode(&usr)
	return usr, nil
}

func (r *keycloakUserCl) SearchUserByEmail(
	ctx context.Context,
	email string,
) ([]app.User, app.Error) {
	const (
		op   = "keycloakUserCl.SearchUsers"
		path = "/admin/realms/opendoor-chat/users"
	)

	// validate
	if email == "" {
		return nil, app.NewErr(400, "required email is missing", op)
	}

	// build url
	queries := url.Values{}
	queries.Add("realm", "opendoor-chat")
	queries.Add("exact", "true")
	queries.Add("max", "1")
	queries.Add("email", email)
	u, _ := url.JoinPath(r.cfg.BaseUrl, path, queries.Encode())

	// build request
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)

	// search
	resp, err := r.cl.Do(req)
	if err != nil {
		return nil, app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return nil, app.NewErr(resp.StatusCode, resp.Status, op)
	}

	//
	var usrs []User
	json.NewDecoder(resp.Body).Decode(&usrs)
	res := make([]app.User, len(usrs))
	for i, u := range usrs {
		res[i] = app.User(u)
	}
	return res, nil
}

func (r *keycloakUserCl) Me(ctx context.Context, accessToken string) (app.User, app.Error) {
	const (
		op   = "keycloakUserCl.Me"
		path = "/realms/opendoor-chat/protocol/openid-connect/userinfo"
	)

	// build request
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, r.cfg.BaseUrl+path, nil)
	req.Header.Add("Authorization", "Bearer "+accessToken)

	// handle response
	resp, err := r.cl.Do(req)
	if err != nil {
		return nil, app.FromErr(err, op)
	}
	if resp.StatusCode != 200 {
		return nil, app.NewErr(resp.StatusCode, resp.Status, op)
	}

	// decode
	var usr UserInfo
	json.NewDecoder(resp.Body).Decode(&usr)
	return usr, nil
}

// UserInfo methods
func (u UserInfo) GetLastName() string {
	return u.LastName
}

func (u UserInfo) GetFirstName() string {
	return u.FirstName
}

func (u UserInfo) GetAttributes() map[string]string {
	return nil
}
func (u UserInfo) GetEmail() string {
	return u.Email
}
func (u UserInfo) IsVerified() bool {
	return u.EmailVerified
}

func (u UserInfo) Validate() error {
	return nil
}

// User methods
func (u User) GetLastName() string {
	return u.LastName
}

func (u User) GetFirstName() string {
	return u.FirstName
}

func (u User) Validate() error {
	if u.Email == "" {
		return errors.New("required Email is missing")
	}
	if len(u.Credentials) == 0 {
		return errors.New("required Password is missing")
	}
	if u.FirstName == "" {
		return errors.New("required FirstName is missing")
	}
	if u.LastName == "" {
		return errors.New("required LastName is missing")
	}

	return nil
}

func (u User) GetAttributes() map[string]string {
	return u.Attributes
}
func (u User) GetEmail() string {
	return u.Email
}
func (u User) IsVerified() bool {
	return u.EmailVerified
}
