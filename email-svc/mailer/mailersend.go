package mailer

import (
	"context"
	"fmt"
	"net/http"
	"net/mail"

	"github.com/benjamonnguyen/opendoor-chat-services/email-svc/model"
	"github.com/jhillyerd/enmime"
	"github.com/mailersend/mailersend-go"
	"github.com/rs/zerolog/log"
)

type mailerSendMailer struct {
	client *mailersend.Mailersend
}

// NewMailerSendMailer constructs a MailerSend API adapter for the Mailer interface
func NewMailerSendMailer(apiKey string) mailerSendMailer {
	return mailerSendMailer{
		client: mailersend.NewMailersend(apiKey),
	}
}

func (mailer mailerSendMailer) Send(
	ctx context.Context,
	payload enmime.Envelope,
) (*http.Response, error) {
	from, err := mail.ParseAddress(payload.GetHeader("From"))
	if err != nil {
		return nil, err
	}

	var rcpts []mailersend.Recipient
	toAddrs, err := mail.ParseAddressList(payload.GetHeader("To"))
	if err != nil {
		return nil, err
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
		log.Error().
			Str("mailer", "mailerSend").
			Err(err).
			Msg("failed sending msg")
		return nil, err
	}

	return resp.Response, nil
}

func (mailer mailerSendMailer) GetEmail(
	ctx context.Context,
	mailerMsgId string,
) (*model.Email, *http.Response, error) {
	root, resp, err := mailer.client.Message.Get(ctx, mailerMsgId)
	if err != nil {
		log.Error().Err(err).Str("mailerMsgId", mailerMsgId).Msg("failed Message.Get")
		return nil, nil, err
	}
	if resp.StatusCode != 200 || len(root.Data.Emails) == 0 {
		return nil, resp.Response, nil
	}
	email := &model.Email{
		MessageId: fmt.Sprintf("<%s@mailersend.net>", root.Data.Emails[0].ID),
	}
	return email, resp.Response, nil
}
