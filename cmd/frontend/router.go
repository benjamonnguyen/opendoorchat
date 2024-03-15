package main

import (
	"net/http"

	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/benjamonnguyen/opendoorchat/frontend/be"
	"github.com/benjamonnguyen/opendoorchat/frontend/html"
	"github.com/benjamonnguyen/opendoorchat/frontend/ws"
	"github.com/gorilla/websocket"
	"github.com/rs/zerolog/log"
	"github.com/urfave/negroni"
)

func buildServer(
	cfg frontend.Config,
	hub *ws.Hub,
	cl *http.Client,
	authenticationCtrl *html.AuthenticationController,
) *http.Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	//
	http.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})
	// App pages
	http.HandleFunc("GET /app", func(w http.ResponseWriter, r *http.Request) {
		// TODO revisit this back button caching dilemma (https://bugzilla.mozilla.org/show_bug.cgi?id=112564)
		http.ServeFile(w, r, "public/app.html")
	})
	http.HandleFunc("GET /app/login", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/login.html")
	})
	http.HandleFunc("GET /app/signup", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/signup.html")
	})
	http.HandleFunc("GET /{path...}", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "public/"+r.PathValue("path"))
	})
	// TODO /app/demo get demo data to populate UI and allow user to click around, but don't allow mutation

	// auth endpoints
	http.HandleFunc("POST /auth/login", authenticationCtrl.LogIn)
	http.HandleFunc("POST /auth/signup", authenticationCtrl.SignUp)
	http.HandleFunc("GET /auth/logout", authenticationCtrl.LogOut)
	http.HandleFunc("GET /api/authenticate-token", authenticationCtrl.AuthenticateToken)

	// backend endpoints
	backendCl := be.NewClient(cl, cfg.Backend.BaseUrl)
	chatCtrl := html.NewChatController(backendCl)
	http.HandleFunc("GET /api/chat-view", chatCtrl.ChatView)
	http.HandleFunc("POST /api/chat", chatCtrl.CreateChat)

	// WS
	http.HandleFunc("GET /ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed ws upgrade")
			return
		}
		hub.Register(ws.NewClient(hub, conn))
	})

	//
	n := negroni.Classic()
	n.UseHandler(http.DefaultServeMux)

	//
	return &http.Server{
		Addr:    cfg.Address,
		Handler: n,
	}
}
