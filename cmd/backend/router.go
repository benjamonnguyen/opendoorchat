package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/emailsvc"
	"github.com/benjamonnguyen/opendoorchat/usersvc"
	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(
	cfg opendoorchat.Config,
	emailsvc emailsvc.EmailController,
	usersvc usersvc.UserController,
) *http.Server {
	router := httprouter.New()
	// email
	router.POST("/email/thread/search", emailsvc.ThreadSearch)
	// user
	router.POST("/user/authenticate", usersvc.Authenticate)
	router.POST("/user", usersvc.CreateUser)

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
