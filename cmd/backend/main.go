package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoorchat/backend"
	"github.com/benjamonnguyen/opendoorchat/backend/emailsvc"
	"github.com/benjamonnguyen/opendoorchat/backend/kafka"
	"github.com/benjamonnguyen/opendoorchat/backend/mailersend"
	"github.com/benjamonnguyen/opendoorchat/backend/mongodb"
	"github.com/jhillyerd/enmime"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	start := time.Now()
	cfg := loadConfig()
	devlog.Init(true, nil)

	// set up graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)
	go func() {
		<-interruptSignal
		cancel()
	}()
	shutdownManager := backend.GracefulShutdownManager{}

	// dependencies
	m := mailersend.NewMailer(cfg.MailerSendApiKey)

	// repositories
	dbClient := initDbClient(ctx, cfg, shutdownManager)
	emailRepo := mongodb.NewEmailRepo(cfg, dbClient)

	// services
	emailService := emailsvc.NewEmailService(emailRepo)

	// controllers
	emailCtrl := emailsvc.NewEmailController(emailService)

	// meat and potatoes
	go startEmailSvcConsumers(ctx, cfg, shutdownManager, emailService, m)
	go listenAndServeRoutes(ctx, cfg, shutdownManager, emailCtrl)

	log.Info().Msgf("started application after %s", time.Since(start).Truncate(time.Second))

	// graceful shutdown
	<-ctx.Done()
	shutdownManager.ShutdownOnInterrupt(20 * time.Second)
}

func loadConfig() backend.Config {
	cfg := backend.LoadConfig("cmd/backend/config.yml")
	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	return cfg
}

func initDbClient(
	ctx context.Context,
	cfg backend.Config,
	shutdownManager backend.GracefulShutdownManager,
) *mongo.Client {
	connCtx, connCanc := context.WithTimeout(ctx, 10*time.Second)
	defer connCanc()
	dbClient := mongodb.ConnectMongoClient(connCtx, cfg.Mongo)
	shutdownManager.AddHandler(func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			log.Error().Err(err).Msg("failed dbClient.Disconnect")
		}
	})
	return dbClient
}

func listenAndServeRoutes(
	ctx context.Context,
	cfg backend.Config,
	shutdownManager backend.GracefulShutdownManager,
	emailCtrl emailsvc.EmailController,
) {
	srv := buildServer(cfg, emailCtrl)
	shutdownManager.AddHandler(func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed srv.Shutdown")
		}
	})
	log.Info().Msgf("starting http server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Error().Err(err).Msg("failed srv.ListenAndServe")
	}
}

func startEmailSvcConsumers(
	ctx context.Context,
	cfg backend.Config,
	shutdownManager backend.GracefulShutdownManager,
	emailService emailsvc.EmailService,
	m emailsvc.Mailer,
) {
	cl := kafka.NewSplitConsumerClient(
		ctx,
		cfg.Kafka,
		fmt.Sprintf("%s-%s", cfg.Kafka.User, "email-svc"),
	)

	// inboundEmailsConsumer
	if err := cl.SetRecordHandler(cfg.Kafka.Topics.InboundEmails, func(rec *kgo.Record) {
		inbound, err := enmime.ReadEnvelope(bytes.NewReader(rec.Value))
		if err != nil {
			log.Error().Err(err).Send()
			return
		}
		if err := emailService.ForwardInboundEmail(ctx, cfg, m, inbound); err != nil {
			// TODO DLQ e.StatusCode()
			log.Error().Err(err).Send()
			return
		}
	}); err != nil {
		log.Fatal().Err(err).Msg("failed AddInboundEmailsConsumer")
	}
	log.Info().Msg("added inboundEmails consumer")
	shutdownManager.AddHandler(func() {
		cl.Shutdown()
	})

	//
	cl.Poll(ctx)
}
