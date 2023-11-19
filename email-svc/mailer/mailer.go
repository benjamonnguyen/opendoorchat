// Package mailer provides implementations of the Mailer interface
// to send transactional emails.
package mailer

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/opendoor-chat-services/email-svc/model"
	"github.com/jhillyerd/enmime"
)

type Mailer interface {
	Send(context.Context, enmime.Envelope) (*http.Response, error)
	GetEmail(context.Context, string) (*model.Email, *http.Response, error)
}
