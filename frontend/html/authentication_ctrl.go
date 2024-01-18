package html

import (
	"io"
	"net/http"
	"time"

	"github.com/benjamonnguyen/opendoorchat/frontend"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
)

type AuthenticationController struct {
	cl frontend.BackendClient
}

func NewAuthenticationController(cl frontend.BackendClient) *AuthenticationController {
	return &AuthenticationController{
		cl: cl,
	}
}

var errHtml = []byte(
	`<div id="login-status"> <small id="login-status-text" style="color: #FF6161;">
Something went wrong. Please wait a moment and try again.</small></div>`,
)

func (a *AuthenticationController) LogIn(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	const op = "AuthenticationController.LogIn"

	// authenticate
	r.ParseForm()
	resp, err := a.cl.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		log.Error().Str("op", op).Err(err).Send()
		w.Write(errHtml)
		return
	}

	// handle response
	if resp.StatusCode == 200 {
		token, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Error().
				Str("op", op).
				Err(err).
				Msg("failed reading repsonse body")
			w.Write(errHtml)
			return
		}
		http.SetCookie(w, &http.Cookie{
			Name:    frontend.AccessTokenCookieKey,
			Value:   string(token),
			Path:    "/",
			Expires: time.Now().Add(24 * time.Hour * 60),
		})
		// TODO remember login email population
		// if vals.Get("remember") == "true" {
		// 	http.SetCookie(w, &http.Cookie{
		// 		Name:    "OPENDOOR_CHAT_EMAIL",
		// 		Value:   vals.Get("email"),
		// 		Path:    "/",
		// 		Expires: time.Now().Add(24 * time.Hour * 365),
		// 	})
		// }
		// http.SetCookie(w, &http.Cookie{
		// 	Name:    "OPENDOOR_CHAT_REMEMBER_LOGIN",
		// 	Value:   vals.Get("remember"),
		// 	Path:    "/",
		// 	Expires: time.Now().Add(24 * time.Hour * 365),
		// })
		w.Header().Set("HX-Redirect", "/app")
		w.WriteHeader(201)
	} else if resp.StatusCode == http.StatusUnauthorized && resp.Status != "invalid api key" {
		w.Write([]byte(`<div id="login-status"><small id="login-status-text" style="color: #FF6161;">
		The email and/or password you entered are not correct.</small></div>`))
	} else {
		log.Error().Str("op", op).Msg(resp.Status)
		w.Write(errHtml)
	}
}

func (a *AuthenticationController) SignUp(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	const op = "AuthenticationController.SignUp"

	// create user
	r.ParseForm()
	user := frontend.User{
		FirstName: r.FormValue("first-name"),
		LastName:  r.FormValue("last-name"),
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
	}
	resp, err := a.cl.CreateUser(r.Context(), user)
	if err != nil {
		log.Error().Str("op", op).Err(err).Send()
		w.Write(errHtml)
		return
	}

	// handle response
	var html string
	if resp.StatusCode == 201 {
		// TODO onboarding page
		html = `<div id="login-status"><small id="login-status-text">
		You're registered! Please verify your email.</small></div>`
	} else if resp.StatusCode == http.StatusConflict {
		html = `<div id="login-status"><small id="login-status-text" style="color: #FF6161;">
		This email is already in use.</small></div>`
	} else {
		log.Error().Str("op", op).Msg(resp.Status)
		w.Write(errHtml)
		return
		// TODO if problem persists, contact ???
	}
	w.Write([]byte(html))
	// TODO verification email
}

func (a *AuthenticationController) LogOut(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	http.SetCookie(w, &http.Cookie{
		Name:    frontend.AccessTokenCookieKey,
		Value:   "",
		Path:    "/",
		Expires: time.UnixMilli(0),
	})
	w.Header().Add("HX-Redirect", "/app/login")
	w.WriteHeader(201)
}

func (a *AuthenticationController) AuthenticateToken(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	token, _ := r.Cookie(frontend.AccessTokenCookieKey)
	if token != nil {
		// TODO AuthenticateToken impl
		w.WriteHeader(201)
		return
	}
	w.Header().Add("HX-Redirect", "/app/login")
	w.WriteHeader(201)
}
