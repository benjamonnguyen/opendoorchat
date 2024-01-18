package app

import "net/http"

type AuthClient interface {
	RequestAccessToken() (*http.Response, error)
}
