package html

import (
	"log"
	"net/http"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/frontend/be"
	"github.com/julienschmidt/httprouter"
)

type AuthenticationController struct {
	cl *be.Client
}

func NewAuthenticationController(cl *be.Client) *AuthenticationController {
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
	token, err := a.cl.Authenticate(r.Context(), r.FormValue("email"), r.FormValue("password"))
	if err != nil {
		log.Println(app.FromErr(err, op))
		if err.StatusCode() == 401 && err.Status() != "invalid api key" {
			w.Write(
				[]byte(
					`<div id="login-status"><small id="login-status-text" style="color: #FF6161;">
					The email and/or password you entered are not correct.</small></div>`,
				),
			)
			return
		}
		w.Write(errHtml)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    be.AccessTokenCookieKey,
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
}

func (a *AuthenticationController) SignUp(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	const op = "AuthenticationController.SignUp"

	// create user
	r.ParseForm()
	user := be.User{
		FirstName: r.FormValue("first-name"),
		LastName:  r.FormValue("last-name"),
		Email:     r.FormValue("email"),
		Password:  r.FormValue("password"),
	}
	err := a.cl.CreateUser(r.Context(), user)
	if err != nil {
		log.Println(app.FromErr(err, op))
		if err.StatusCode() == http.StatusConflict {
			w.Write(
				[]byte(
					`<div id="login-status"><small id="login-status-text" style="color: #FF6161;">
					This email is already in use.</small></div>`,
				))
			return
		}
		w.Write(errHtml)
		return
	}

	//
	w.Write([]byte(`<div id="login-status"><small id="login-status-text">
	You're registered! Please verify your email.</small></div>`))
	// TODO verification email
}

func (a *AuthenticationController) LogOut(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	http.SetCookie(w, &http.Cookie{
		Name:    be.AccessTokenCookieKey,
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
	token, _ := r.Cookie(be.AccessTokenCookieKey)
	if token != nil {
		// TODO AuthenticateToken impl
		w.WriteHeader(201)
		return
	}
	w.Header().Add("HX-Redirect", "/app/login")
	w.WriteHeader(201)
}
