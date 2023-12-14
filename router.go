package main

import (
	"net/http"

	emailctrl "github.com/benjamonnguyen/opendoor-chat/email-svc/controller"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(addr string, emailCtrl emailctrl.EmailController) *http.Server {
	router := httprouter.New()
	router.POST("/email/thread/search", emailCtrl.ThreadSearch)

	n := negroni.Classic()
	n.UseHandler(router)

	return &http.Server{
		Addr:    addr,
		Handler: n,
	}
}
