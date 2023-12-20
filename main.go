package main

import (
	"context"
	"fmt"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	"github.com/benjamonnguyen/opendoor-chat/commons/db"
	"github.com/benjamonnguyen/opendoor-chat/commons/mq"
	"github.com/benjamonnguyen/opendoor-chat/commons/service"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/consumer"
	emailctrl "github.com/benjamonnguyen/opendoor-chat/email-svc/controller"
	"github.com/benjamonnguyen/opendoor-chat/email-svc/mailer"
	emailrepo "github.com/benjamonnguyen/opendoor-chat/email-svc/repo"
	emailservice "github.com/benjamonnguyen/opendoor-chat/email-svc/service"
	userctrl "github.com/benjamonnguyen/opendoor-chat/user-svc/controller"
	userrepo "github.com/benjamonnguyen/opendoor-chat/user-svc/repo"
	userservice "github.com/benjamonnguyen/opendoor-chat/user-svc/service"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.mongodb.org/mongo-driver/mongo"
)

func main() {
	start := time.Now()
	cfg := loadConfig()
	devlog.Enable(true)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	shutdownManager := service.GracefulShutdownManager{}

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

func loadConfig() config.Config {
	cfg := config.Load(".")
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
	emailCtrl emailctrl.EmailController,
	userCtrl userctrl.UserController,
) {
	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := buildServer(addr, emailCtrl, userCtrl)
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
	log.Info().Msgf("started http server on %s", addr)
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
