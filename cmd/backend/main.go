package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/email-svc/consumer"
	emailctrl "github.com/benjamonnguyen/opendoorchat/email-svc/controller"
	"github.com/benjamonnguyen/opendoorchat/email-svc/mailer"
	emailrepo "github.com/benjamonnguyen/opendoorchat/email-svc/repo"
	emailservice "github.com/benjamonnguyen/opendoorchat/email-svc/service"
	"github.com/benjamonnguyen/opendoorchat/kafka"
	"github.com/benjamonnguyen/opendoorchat/mongodb"
	userctrl "github.com/benjamonnguyen/opendoorchat/user-svc/controller"
	userrepo "github.com/benjamonnguyen/opendoorchat/user-svc/repo"
	userservice "github.com/benjamonnguyen/opendoorchat/user-svc/service"
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
	m := mailer.NewMailerSendMailer(cfg.MailerSendApiKey)

	// repositories
	dbClient := initDbClient(ctx, cfg, shutdownManager)
	emailRepo := emailrepo.NewMongoEmailRepo(cfg, dbClient)
	userRepo := userrepo.NewMongoUserRepo(cfg, dbClient)

	// services
	emailService := emailservice.NewEmailService(emailRepo)
	userService := userservice.NewUserService(userRepo)

	// controllers
	emailCtrl := emailctrl.NewEmailController(emailService)
	userCtrl := userctrl.NewUserController(userService)

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
	emailCtrl emailctrl.EmailController,
	userCtrl userctrl.UserController,
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
	emailService emailservice.EmailService,
	m mailer.Mailer,
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
