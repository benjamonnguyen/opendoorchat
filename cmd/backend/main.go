package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/kafka"
	"github.com/benjamonnguyen/opendoorchat/mailersend"
	"github.com/benjamonnguyen/opendoorchat/mongodb"
	"github.com/benjamonnguyen/opendoorchat/services/emailsvc"
	"github.com/benjamonnguyen/opendoorchat/services/emailsvc/consumer"
	"github.com/benjamonnguyen/opendoorchat/services/usersvc"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	start := time.Now()
	cfg := loadConfig()
	devlog.Init(true, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdownManager := opendoorchat.GracefulShutdownManager{}

	// dependencies
	m := mailersend.NewMailer(cfg.MailerSendApiKey)

	// repositories
	dbClient := initDbClient(ctx, cfg, shutdownManager)
	emailRepo := mongodb.NewEmailRepo(cfg, dbClient)
	userRepo := mongodb.NewUserRepo(cfg, dbClient)

	// services
	emailService := emailsvc.NewEmailService(emailRepo)
	userService := usersvc.NewUserService(userRepo)

	// controllers
	emailCtrl := emailsvc.NewEmailController(emailService)
	userCtrl := usersvc.NewUserController(userService)

	startEmailSvcConsumers(ctx, cfg, shutdownManager, emailService, m)
	listenAndServeRoutes(ctx, cfg, shutdownManager, emailCtrl, userCtrl)

	log.Info().Msgf("started application after %s", time.Since(start).Truncate(time.Second))

	shutdownManager.ShutdownOnInterrupt(20 * time.Second)
}

func loadConfig() opendoorchat.Config {
	cfg := opendoorchat.LoadConfig(".")
	lvl, err := zerolog.ParseLevel(cfg.LogLevel)
	if err != nil {
		lvl = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(lvl)

	return cfg
}

func initDbClient(
	ctx context.Context,
	cfg opendoorchat.Config,
	shutdownManager opendoorchat.GracefulShutdownManager,
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
	cfg opendoorchat.Config,
	shutdownManager opendoorchat.GracefulShutdownManager,
	emailCtrl emailsvc.EmailController,
	userCtrl usersvc.UserController,
) {
	srv := buildServer(cfg, emailCtrl, userCtrl)
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
	log.Info().Msgf("started http server on %s", srv.Addr)
}

func startEmailSvcConsumers(
	ctx context.Context,
	cfg opendoorchat.Config,
	shutdownManager opendoorchat.GracefulShutdownManager,
	emailService emailsvc.EmailService,
	m emailsvc.Mailer,
) {
	consumerCl := kafka.NewSplitConsumerClient(
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
