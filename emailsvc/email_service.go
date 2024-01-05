package emailsvc

import (
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/benjamonnguyen/opendoorchat"
	"github.com/jhillyerd/enmime"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EmailService interface {
	ThreadSearch(
		ctx context.Context,
		st ThreadSearchTerms,
	) (EmailThread, opendoorchat.Error)
	AddEmail(
		ctx context.Context,
		threadId primitive.ObjectID,
		email Email,
	) opendoorchat.Error
	ForwardInboundEmail(
		ctx context.Context,
		cfg opendoorchat.Config,
		m Mailer,
		inbound *enmime.Envelope,
	) opendoorchat.Error
}

var _ EmailService = (*emailService)(nil)

type emailService struct {
	repo EmailRepo
}

func NewEmailService(repo EmailRepo) *emailService {
	return &emailService{
		repo: repo,
	}
}

func (s *emailService) ThreadSearch(
	ctx context.Context,
	st ThreadSearchTerms,
) (EmailThread, opendoorchat.Error) {
	if st == (ThreadSearchTerms{}) {
		return EmailThread{}, opendoorchat.NewErr(400, "missing ThreadSearchTerms", "")
	}

	thread, err := s.repo.ThreadSearch(ctx, st)
	if err != nil {
		return EmailThread{}, err
	}

	return thread, nil
}

func (s *emailService) AddEmail(
	ctx context.Context,
	threadId primitive.ObjectID,
	email Email,
) opendoorchat.Error {
	if threadId == primitive.NilObjectID {
		return opendoorchat.NewErr(400, "missing threadId", "")
	}
	if email == (Email{}) {
		return opendoorchat.NewErr(400, "missing email", "")
	}

	if err := s.repo.AddEmail(ctx, threadId, email); err != nil {
		return err
	}
	return nil
}

// ForwardInboundEmail forwards inbound email to participants of the EmailThread
// matching the "In-Reply-To" header
func (s *emailService) ForwardInboundEmail(
	ctx context.Context,
	cfg opendoorchat.Config,
	m Mailer,
	inbound *enmime.Envelope,
) opendoorchat.Error {
	start := time.Now()
	const op = "ForwardInboundEmail"
	// devlog.Printf("got inbound email %#v\n", inbound)

	if inbound == nil {
		return opendoorchat.NewErr(400, "inbound is nil", "")
	}
	// get thread
	msgId := inbound.GetHeader("In-Reply-To")
	threadCtx, threadCanc := context.WithTimeout(ctx, cfg.ReadTimeout)
	defer threadCanc()
	st := ThreadSearchTerms{
		EmailMessageId: msgId,
	}
	thread, err := s.ThreadSearch(threadCtx, st)
	if err != nil {
		err = opendoorchat.FromErr(err, op)
		return err
	}

	// contruct outbound
	outbound := inbound.Clone()
	outbound.DeleteHeader("To")
	senderAddr, e := mail.ParseAddress(outbound.GetHeader("From"))
	if e != nil {
		err := opendoorchat.FromErr(e, fmt.Sprintf("%s: ParseAddress", op))
		log.Error().Err(err).Str("Header[From]", outbound.GetHeader("From")).Send()
		return err
	}
	for _, p := range thread.Participants {
		p := p
		if p.Email == senderAddr.Address {
			outbound.SetHeader("From",
				[]string{fmt.Sprintf("%s <%s@%s>", p.Name(), "mailer", cfg.Domain)})
		} else {
			outbound.AddHeader("To",
				fmt.Sprintf("%s <%s>", p.Name(), p.Email))
		}
	}

	// send email
	sendCtx, sendCanc := context.WithTimeout(ctx, cfg.RequestTimeout)
	defer sendCanc()
	mailerResp, err := m.Send(sendCtx, *outbound)
	if err != nil {
		err = opendoorchat.FromErr(err, op)
		log.Error().Err(err).Send()
		return err
	}
	sent, _ := inbound.Date()
	log.Debug().
		Int("code", mailerResp.StatusCode).
		Dur("timeSinceSent", time.Since(sent)).
		Dur("timeSinceConsumed", time.Since(start)).
		Msg("got Mailer.Send() response")

	// add new messageId to thread
	if mailerResp.StatusCode == 202 {
		email, err := m.GetEmail(ctx, mailerResp.Header.Get("X-Message-Id"))
		if err != nil {
			err = opendoorchat.FromErr(err, op)
			log.Error().Err(err).Send()
			return err
		}
		addCtx, addCanc := context.WithTimeout(ctx, cfg.RequestTimeout)
		defer addCanc()
		log.Debug().Str("emailMessageId", email.MessageId).Msg("AddEmail")
		err = s.AddEmail(addCtx, thread.Id, email)
		if err != nil {
			err = opendoorchat.FromErr(err, op)
			log.Error().Err(err).Send()
			return err
		}
		return nil
	}
	return opendoorchat.NewErr(mailerResp.StatusCode, mailerResp.Status, "")
}
