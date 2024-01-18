package httpcl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/benjamonnguyen/opendoorchat/frontend/config"
)

type backendClient struct {
	cl  *http.Client
	cfg config.Config
}

var _ frontend.BackendClient = (*backendClient)(nil)

func NewBackendClient(cl *http.Client, cfg config.Config) *backendClient {
	return &backendClient{
		cl:  cl,
		cfg: cfg,
	}
}

func (cl *backendClient) Authenticate(
	ctx context.Context,
	email, password string,
) (*http.Response, error) {
	const op = "backendClient.SignIn"

	// encode args
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).
		Encode(fmt.Sprintf(`{"email": %s, "password": %s}`, email, password))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// build request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.cfg.BackendBaseUrl+"/user/authenticate",
		buf,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := addAccessTokenHeader(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// TODO reCAPTCHA
	// req.Header.Set("Authorization", a.cfg.BackendApiKey)
	resp, err := cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func (cl *backendClient) CreateUser(
	ctx context.Context,
	user frontend.User,
) (*http.Response, error) {
	const op = "backendClient.SignUp"
	// TODO client side validations must be re-done on the server side, as they can always be bypassed.

	// encode args
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).
		Encode(user)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// build request
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		cl.cfg.BackendBaseUrl+"/user",
		buf,
	)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if err := addAccessTokenHeader(req); err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	// get response
	resp, err := cl.cl.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return resp, nil
}

func addAccessTokenHeader(req *http.Request) error {
	token, err := req.Cookie(frontend.AccessTokenCookieKey)
	if err != nil {
		return err
	}

	req.Header.Add(frontend.AccessTokenHeaderKey, token.Value)
	return nil
}
