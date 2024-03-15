package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/benjamonnguyen/opendoorchat/frontend/html"
	"github.com/benjamonnguyen/opendoorchat/frontend/ws"
	"github.com/benjamonnguyen/opendoorchat/keycloak"
)

func main() {
	// config
	cfgFile := flag.String("cfg", "config.yml", "configuration file")
	flag.Parse()
	cfg, err := frontend.LoadConfig(*cfgFile)
	if err != nil {
		log.Fatal(err)
	}
	devlog.Init(true, nil)

	// graceful shutdown setup
	ctx, cancel := context.WithCancel(context.Background())
	interruptCh := make(chan os.Signal, 1)
	signal.Notify(interruptCh, os.Interrupt)
	go func() {
		<-interruptCh
		cancel()
	}()

	// ws
	hub := ws.NewHub()
	go hub.Run(ctx)

	// clients
	cl := &http.Client{
		Timeout: time.Minute,
	}
	authCl := keycloak.NewAuthClient(cl, cfg.Keycloak)
	userRepo := keycloak.NewUserRepo(cl, cfg.Keycloak)
	authenticationCtrl := html.NewAuthenticationController(authCl, userRepo)

	// server
	srv := buildServer(cfg, hub, cl, authenticationCtrl)
	go func() {
		if err := srv.ListenAndServe(); err != nil {
			log.Println("ListenAndServe:", err)
		}
	}()
	log.Println("started http server at", srv.Addr)

	<-ctx.Done()
	// graceful shutdown
	start := time.Now()
	shtudownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	srv.Shutdown(shtudownCtx)

	log.Printf("completed graceful shutdown after %s", time.Since(start))
}
