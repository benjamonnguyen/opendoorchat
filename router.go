package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/benjamonnguyen/opendoor-chat/commons/config"
	emailctrl "github.com/benjamonnguyen/opendoor-chat/email-svc/controller"
	userctrl "github.com/benjamonnguyen/opendoor-chat/user-svc/controller"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(
	cfg config.Config,
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
	// n.UseFunc(func(w http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	// 	// TODO authentication middleware
	// 	if r.Header.Get("Authorization") != cfg.ApiKey {
	// 		http.Error(w, "invalid api key", http.StatusUnauthorized)
	// 		return
	// 	}
	// 	next(w, r)
	// })
	n.UseHandler(router)

	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      n,
		ReadTimeout:  time.Minute,
		WriteTimeout: time.Minute,
	}
}
