package consumer

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"fmt"
// 	"time"

// 	"github.com/benjamonnguyen/opendoor-commons/mq"
// 	"github.com/benjamonnguyen/opendoor-email-svc/internal/emailcfg"
// 	"github.com/benjamonnguyen/opendoor-email-svc/internal/mailer"
// 	"github.com/benjamonnguyen/opendoor-email-svc/internal/service"
// 	"github.com/benjamonnguyen/opendoor-protos/pb/chatpb"
// 	"github.com/benjamonnguyen/opendoor-protos/pb/emailpb"
// 	"github.com/benjamonnguyen/opendoor-protos/pb/userpb"
// 	"github.com/jhillyerd/enmime"
// 	"github.com/rs/zerolog/log"
// 	"github.com/twmb/franz-go/pkg/kgo"
// )

// const (
// 	chatMessagesConsumer = "chat-messages"
// )

// func AddChatMessagesConsumer(
// 	ctx context.Context,
// 	cfg emailcfg.Config,
// 	emailSvc service.IEmailService,
// 	m mailer.Mailer,
// 	cl mq.KafkaConsumerClient,
// ) {
// 	if err := cl.SetRecordHandler(cfg.Kafka.Topics[chatMessagesConsumer], func(rec *kgo.Record) {
// 		sendEmail(ctx, cfg, emailSvc, m, rec)
// 	}); err != nil {
// 		log.Fatal().Err(err).Msg("failed AddChatMessagesConsumer")
// 	}
// 	log.Info().Msg("added chatMessages consumer")
// }

// func sendEmail(
// 	ctx context.Context,
// 	cfg emailcfg.Config,
// 	emailSvc service.IEmailService,
// 	m mailer.Mailer,
// 	rec *kgo.Record,
// ) {
// 	start := time.Now()
// 	log.Debug().
// 		Str("record", string(rec.Value)).
// 		Str("consumer", chatMessagesConsumer).
// 		Msg("got chat message")
// 	var payload *chatpb.MessagePayload
// 	err := json.NewDecoder(bytes.NewReader(rec.Value)).Decode(payload)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed decoding record")
// 		return
// 	}

// 	// get thread
// 	threadCtx, threadCanc := context.WithTimeout(ctx, cfg.ReadTimeout)
// 	defer threadCanc()
// 	req := &emailpb.ThreadSearchReq{
// 		ChatId: &payload.ChatId,
// 	}
// 	threadResp, err := emailSvc.ThreadSearch(threadCtx, req)
// 	if err != nil {
// 		log.Error().Err(err).Interface("req", req).Msg("failed ThreadSearch")
// 		return
// 	}

// 	// get sender/rcpt
// 	sender, rcpts, err := getSenderAndRcpts(
// 		payload.Message.From,
// 		threadResp.EmailThread,
// 	)
// 	if err != nil {
// 		log.Error().Err(err).Msg("failed getSenderAndRcpt")
// 		return
// 	}

// 	// construct email
// 	enmime.Builder().

// 	// mailer.Send

// 	// emailSvc.AddEmail
// }

// func getSenderAndRcpts(
// 	from string,
// 	thread *emailpb.EmailThread,
// ) (sender *userpb.User, rcpts []*userpb.User, err error) {
// 	for i, _ := range thread.Participants {
// 		p := thread.Participants[i]
// 		if p.Id == from {
// 			sender = p
// 		} else {
// 			rcpts = append(rcpts, p)
// 		}
// 	}
// 	if sender == nil {
// 		err = fmt.Errorf("sender %s not found in thread %s", from, thread.Id)
// 		return nil, nil, err
// 	}
// 	return
// }
