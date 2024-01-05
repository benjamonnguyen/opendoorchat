package emailsvc

import (
	"context"
	"net/http"

	"github.com/benjamonnguyen/opendoorchat"
	"github.com/jhillyerd/enmime"
)

// Mailer provdes API to send and manage transactional emails.
type Mailer interface {
	Send(context.Context, enmime.Envelope) (*http.Response, opendoorchat.Error)
	GetEmail(context.Context, string) (Email, opendoorchat.Error)
}
