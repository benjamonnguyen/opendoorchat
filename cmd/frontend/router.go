package main

import (
	"net/http"

	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/benjamonnguyen/opendoorchat/frontend/be"
	"github.com/benjamonnguyen/opendoorchat/frontend/html"
	"github.com/benjamonnguyen/opendoorchat/frontend/ws"
	"github.com/gorilla/websocket"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"github.com/urfave/negroni"
)

func buildServer(
	cfg frontend.Config,
	addr string,
	hub *ws.Hub,
	cl *http.Client,
	authenticationCtrl *html.AuthenticationController,
) *http.Server {
	upgrader := websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	router := httprouter.New()
	//
	router.GET("/", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		w.Write([]byte("Hello, World!"))
	})
	// App pages
	router.GET("/app", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		// TODO revisit this back button caching dilemma (https://bugzilla.mozilla.org/show_bug.cgi?id=112564)
		http.ServeFile(w, r, "frontend/public/app.html")
	})
	router.GET("/app/login", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		http.ServeFile(w, r, "frontend/public/login.html")
	})
	router.GET("/app/signup", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		http.ServeFile(w, r, "frontend/public/signup.html")
	})
	// TODO /app/demo get demo data to populate UI and allow user to click around, but don't allow mutation

	// auth endpoints
	router.POST("/auth/login", authenticationCtrl.LogIn)
	router.POST("/auth/signup", authenticationCtrl.SignUp)
	router.GET("/auth/logout", authenticationCtrl.LogOut)
	router.GET("/api/authenticate-token", authenticationCtrl.AuthenticateToken)

	// backend endpoints
	backendCl := be.NewClient(cl, cfg.Backend.BaseUrl)
	chatCtrl := html.NewChatController(backendCl)
	router.GET("/api/chat-view", chatCtrl.ChatView)
	router.POST("/api/chat", chatCtrl.CreateChat)

	// WS
	router.GET("/ws", func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Error().Err(err).Msg("failed ws upgrade")
			return
		}
		hub.Register(ws.NewClient(hub, conn))
	})

	// CSS
	router.GET(
		"/css/:file",
		func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			http.ServeFile(w, r, "frontend/public/css/"+p.ByName("file"))
		},
	)

	// Assets
	router.GET(
		"/assets/*filepath",
		func(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
			http.ServeFile(w, r, "frontend/public/"+p.ByName("filepath"))
		},
	)

	//
	n := negroni.Classic()
	n.UseHandler(router)

	//
	return &http.Server{
		Addr:    addr,
		Handler: n,
	}
}
