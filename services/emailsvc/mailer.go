package emailsvc

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/gootils/httputil"
	"github.com/jhillyerd/enmime"
)

// Mailer provdes API to send and manage transactional emails.
type Mailer interface {
	Send(context.Context, enmime.Envelope) (*http.Response, error)
	GetEmail(context.Context, string) (Email, httputil.HttpError)
}
