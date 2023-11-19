package main

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/urfave/negroni"
)

func buildServer(addr string) *http.Server {
	router := httprouter.New()

	n := negroni.Classic()
	n.UseHandler(router)

	return &http.Server{
		Addr:    addr,
		Handler: n,
	}
}
