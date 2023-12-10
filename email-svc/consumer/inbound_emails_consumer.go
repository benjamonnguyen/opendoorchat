package consumer

import (
	"bytes"
	"context"
	"fmt"
	"net/mail"
	"time"

	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/commons/mq"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/mailer"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/model"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/service"
	"github.com/jhillyerd/enmime"
	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
)

const (
	inboundEmailsConsumer = "inbound-emails"
)

func AddInboundEmailsConsumer(
	ctx context.Context,
	cfg config.Config,
	emailSvc service.EmailService,
	m mailer.Mailer,
	cl mq.KafkaConsumerClient,
) {
	if err := cl.SetRecordHandler(cfg.Kafka.Topics[inboundEmailsConsumer], func(rec *kgo.Record) {
		forwardEmail(ctx, cfg, emailSvc, m, rec)
	}); err != nil {
		log.Fatal().Err(err).Msg("failed AddInboundEmailsConsumer")
	}
	log.Info().Msg("added inboundEmails consumer")
}

// forwardEmail forwards the inbound email to participants of the EmailThread
// matching the "In-Reply-To" header
func forwardEmail(
	ctx context.Context,
	cfg config.Config,
	emailSvc service.EmailService,
	m mailer.Mailer,
	rec *kgo.Record,
) {
	start := time.Now()
	log.Debug().
		Str("record", string(rec.Value)).
		Str("consumer", inboundEmailsConsumer).
		Msg("got inbound email")

	//
	inbound, err := enmime.ReadEnvelope(bytes.NewReader(rec.Value))
	if err != nil {
		log.Error().Err(err).Msg("failed ReadEnvelope")
		return
	}

	// get thread
	msgId := inbound.GetHeader("In-Reply-To")
	threadCtx, threadCanc := context.WithTimeout(ctx, cfg.ReadTimeout)
	defer threadCanc()
	st := model.ThreadSearchTerms{
		EmailMessageId: msgId,
	}
	thread, httperr := emailSvc.ThreadSearch(threadCtx, st)
	if httperr != nil {
		log.Error().Err(err).Interface("searchTerms", st).Msg("failed ThreadSearch")
		// TODO DLQ using status.Code(err)
		return
	}

	// contruct outbound
	outbound := inbound.Clone()
	outbound.DeleteHeader("To")
	senderAddr, err := mail.ParseAddress(inbound.GetHeader("From"))
	if err != nil {
		log.Error().Err(err).Msg("failed ParseAddress for From header")
		return
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
		log.Error().Err(err).Msg("failed forwarding email")
		return
	}
	sent, _ := inbound.Date()
	log.Debug().
		Int("code", mailerResp.StatusCode).
		Dur("timeSinceSent", time.Since(sent)).
		Dur("timeSinceConsumed", time.Since(start)).
		Msg("got Mailer.Send() response")

	// add new messageId to thread
	if mailerResp.StatusCode == 202 {
		email, httperr := m.GetEmail(ctx, mailerResp.Header.Get("X-Message-Id"))
		if httperr != nil {
			log.Error().Err(httperr).Msg("failed Mailer.GetEmail")
			return
		}
		addCtx, addCanc := context.WithTimeout(ctx, cfg.RequestTimeout)
		defer addCanc()
		err = emailSvc.AddEmail(addCtx, thread.Id, email)
		if err != nil {
			log.Error().Err(err).Msg("failed AddEmail")
			return
		}
	}
	// TODO DLQ
}
