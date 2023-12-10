package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/commons/db"
	"github.com/benjamonnguyen/opendoor-chat/commons/mq"
	"github.com/benjamonnguyen/opendoor-chat/commons/service"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/consumer"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/mailer"
	emailrepo "github.com/benjamonnguyen/opendoor-chat/email-svc/repo"
	emailservice "github.com/benjamonnguyen/opendoor-chat/email-svc/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	start := time.Now()
	cfg := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdownManager := service.GracefulShutdownManager{}

	// repositories
	dbClient := initDbClient(ctx, cfg, shutdownManager)
	emailRepo := emailrepo.NewMongoEmailRepo(cfg, dbClient)

	// services
	emailService := emailservice.NewEmailService(emailRepo)

	// dependencies
	m := mailer.NewMailerSendMailer(cfg.MailerSendApiKey)

	startEmailSvcConsumers(ctx, cfg, shutdownManager, emailService, m)
	listenAndServeRoutes(ctx, cfg, shutdownManager)

	log.Info().Msgf("started after %s", time.Since(start).Truncate(time.Second))

	shutdownManager.ShutdownOnInterrupt(20 * time.Second)
}

func loadConfig() config.Config {
	var cfg config.Config
	config.LoadConfig("config", "yaml", ".", &cfg)
	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	return cfg
}

func initDbClient(
	ctx context.Context,
	cfg config.Config,
	shutdownManager service.GracefulShutdownManager,
) *mongo.Client {
	connCtx, connCanc := context.WithTimeout(ctx, 10*time.Second)
	defer connCanc()
	dbClient := db.ConnectMongoClient(connCtx, cfg.Mongo)
	shutdownManager.AddHandler(func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			log.Error().Err(err).Msg("failed dbClient.Disconnect")
		}
	})
	return dbClient
}

func listenAndServeRoutes(
	ctx context.Context,
	cfg config.Config,
	shutdownManager service.GracefulShutdownManager,
) {
	srv := buildServer(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port))
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("failed srv.ListenAndServe")
		}
	}()
	shutdownManager.AddHandler(func() {
		if err := srv.Shutdown(ctx); err != nil {
			log.Error().Err(err).Msg("failed srv.Shutdown")
		}
	})
}

func startEmailSvcConsumers(
	ctx context.Context,
	cfg config.Config,
	shutdownManager service.GracefulShutdownManager,
	emailService emailservice.EmailService,
	m mailer.Mailer,
) {
	consumerCl := mq.NewSplitConsumerClient(
		ctx,
		cfg.Kafka,
		fmt.Sprintf("%s-%s", cfg.Kafka.User, "email-svc"),
	)
	consumer.AddInboundEmailsConsumer(ctx, cfg, emailService, m, consumerCl)
	shutdownManager.AddHandler(func() {
		consumerCl.Shutdown()
	})
	go consumerCl.Poll(ctx)
}
