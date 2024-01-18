package mailersend

import (
	"context"
	"fmt"
	"math"
	"net/http"
	"net/mail"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/backend/emailsvc"
	"github.com/jhillyerd/enmime"
	"github.com/mailersend/mailersend-go"
	"github.com/rs/zerolog/log"
)

var _ emailsvc.Mailer = (*mailerSendMailer)(nil)

type mailerSendMailer struct {
	client *mailersend.Mailersend
}

// NewMailer constructs a MailerSend API adapter for the Mailer interface
func NewMailer(apiKey string) mailerSendMailer {
	return mailerSendMailer{
		client: mailersend.NewMailersend(apiKey),
	}
}

func (mailer mailerSendMailer) Send(
	ctx context.Context,
	payload enmime.Envelope,
) (*http.Response, app.Error) {
	const op = "mailerSendMailer.Send"
	from, err := mail.ParseAddress(payload.GetHeader("From"))
	if err != nil {
		return nil, app.FromErr(err, fmt.Sprintf("%s: ParseAddress", op))
	}

	var rcpts []mailersend.Recipient
	toAddrs, err := mail.ParseAddressList(payload.GetHeader("To"))
	if err != nil {
		return nil, app.FromErr(err, fmt.Sprintf("%s: ParseAddressList", op))
	}
	for _, addr := range toAddrs {
		rcpts = append(rcpts, mailersend.Recipient{Name: addr.Name, Email: addr.Address})
	}

	msg := mailer.client.Email.NewMessage()
	msg.SetFrom(mailersend.Recipient{Name: from.Name, Email: from.Address})
	msg.SetRecipients(rcpts)
	msg.SetSubject(payload.GetHeader("Subject"))
	msg.SetHTML(payload.HTML)
	msg.SetText(payload.Text)
	inReplyTo := payload.GetHeader("In-Reply-To")
	if len(inReplyTo) > 2 {
		msg.SetInReplyTo(inReplyTo[1 : len(inReplyTo)-1])
	}
	// TODO msg.SetTags(tags)
	// TODO msg.Attachments()
	// TODO msg.TemplateID()
	log.Debug().
		Str("mailer", "mailerSend").
		Interface("msg", msg).
		Msg("sending msg")

	resp, err := mailer.client.Email.Send(ctx, msg)
	if err != nil {
		e := app.FromErr(err, fmt.Sprintf("%s: mailerSend.EmailService.Send", op))
		log.Error().Err(e).Send()
		return nil, e
	}

	return resp.Response, nil
}

func (mailer mailerSendMailer) GetEmail(
	ctx context.Context,
	mailerMsgId string,
) (emailsvc.Email, app.Error) {
	const op = "mailerSendMailer.GetEmail"
	for i := 0; i < 3; i++ {
		root, resp, err := mailer.client.Message.Get(ctx, mailerMsgId)
		if err != nil {
			log.Error().Err(err).Str("mailerMsgId", mailerMsgId).Msg("failed Message.Get")
			return emailsvc.Email{}, app.FromErr(
				err,
				fmt.Sprintf("%s: mailerSend.MessageService.Get", op),
			)
		}
		if resp.StatusCode != 200 {
			return emailsvc.Email{}, app.NewErr(resp.StatusCode, resp.Status, "")
		}
		if len(root.Data.Emails) == 0 {
			backoff := time.Duration(math.Pow(6.0, float64(i))) * time.Second
			log.Debug().Int("retry", i).Dur("backoff", backoff).Msg("GetEmail: not found")
			time.Sleep(backoff)
			continue
		}
		return emailsvc.Email{
			MessageId: fmt.Sprintf("<%s@mailersend.net>", root.Data.Emails[0].ID),
		}, nil
	}
	return emailsvc.Email{}, app.NewErr(404, "", "")
}
