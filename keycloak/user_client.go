package keycloak

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// User methods
func (u User) Name() string {
	return fmt.Sprintf("%s %s", u.FirstName, u.LastName)
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

// "userinfo_endpoint":"http://localhost:9090/realms/opendoor-chat/protocol/openid-connect/userinfo",
