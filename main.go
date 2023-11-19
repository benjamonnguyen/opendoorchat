package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benjamonnguyen/opendoor-chat-services/commons/config"
	"github.com/benjamonnguyen/opendoor-chat-services/commons/db"
	"github.com/benjamonnguyen/opendoor-chat-services/commons/mq"
	"github.com/benjamonnguyen/opendoor-chat-services/commons/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

var (
	dbClient *mongo.Client
)

func main() {
	start := time.Now()
	cfg := loadConfig()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdownManager := service.GracefulShutdownManager{}

	// dependencies
	// m := mailer.NewMailerSendMailer(cfg.MailerSendApiKey)

	initDbClient(ctx, cfg, shutdownManager)
	startConsumers(ctx, cfg, shutdownManager)
	listenAndServeRoutes(ctx, cfg, shutdownManager)

	log.Info().Msgf("started email-svc after %s", time.Since(start))

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
) {
	connCtx, connCanc := context.WithTimeout(ctx, 10*time.Second)
	defer connCanc()
	dbClient = db.ConnectMongoClient(connCtx, cfg.Mongo)
	shutdownManager.AddHandler(func() {
		if err := dbClient.Disconnect(ctx); err != nil {
			log.Error().Err(err).Msg("failed dbClient.Disconnect")
		}
	})
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

func startConsumers(
	ctx context.Context,
	cfg config.Config,
	shutdownManager service.GracefulShutdownManager,
) {
	consumerCl := mq.NewSplitConsumerClient(
		ctx,
		cfg.Kafka,
		fmt.Sprintf("%s-%s", cfg.Kafka.User, "email-svc"),
	)
	go consumerCl.Poll(ctx)
	shutdownManager.AddHandler(func() {
		consumerCl.Shutdown()
	})
}
