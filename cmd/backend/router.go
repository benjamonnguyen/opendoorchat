package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/benjamonnguyen/opendoorchat/backend"
	"github.com/benjamonnguyen/opendoorchat/backend/emailsvc"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(
	cfg backend.Config,
	emailsvc emailsvc.EmailController,
) *http.Server {
	router := httprouter.New()
	// email
	router.POST("/email/thread/search", emailsvc.ThreadSearch)

	n := negroni.Classic()
	n.UseHandler(router)

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      n,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
}
