package main

import (
	"net/http"
	"time"

	emailctrl "github.com/benjamonnguyen/opendoor-chat/email-svc/controller"
	userctrl "github.com/benjamonnguyen/opendoor-chat/user-svc/controller"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(
	addr string,
	emailCtrl emailctrl.EmailController,
	userCtrl userctrl.UserController,
) *http.Server {
	router := httprouter.New()
	// email
	router.POST("/email/thread/search", emailCtrl.ThreadSearch)
	// user
	router.POST("/user/authenticate", userCtrl.Authenticate)
	router.POST("/user", userCtrl.CreateUser)

	n := negroni.Classic()
	n.UseHandler(router)

	// TODO auth middleware

	return &http.Server{
		Addr:         addr,
		Handler:      n,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
}
