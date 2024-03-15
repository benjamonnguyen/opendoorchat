package html

import (
	"log"
	"net/http"
	"time"

	app "github.com/benjamonnguyen/opendoorchat"
	"github.com/benjamonnguyen/opendoorchat/keycloak"
)

type AuthenticationController struct {
	cl       *keycloak.AuthClient
	userRepo app.UserRepo
}

func NewAuthenticationController(
	cl *keycloak.AuthClient,
	userRepo app.UserRepo,
) *AuthenticationController {
	return &AuthenticationController{
		cl:       cl,
		userRepo: userRepo,
	}
}

var errHtml = []byte(
	`<div id="login-status"> <small id="login-status-text" style="color: #FF6161;">
Something went wrong. Please wait a moment and try again.</small></div>`,
)

func (a *AuthenticationController) LogIn(w http.ResponseWriter, r *http.Request) {
	const op = "AuthenticationController.LogIn"
	minTime := time.Now().Add(time.Second)
	// authenticate
	r.ParseForm()
	_, refreshToken, err := a.cl.RequestAccessToken(
		r.Context(),
		"",
		r.FormValue("email"),
		r.FormValue("password"),
	)
	time.Sleep(time.Until(minTime)) // ensure loading animation lasts at least set duration
	if err != nil {
		log.Println(app.FromErr(err, op))
		if err.StatusCode() == 401 {
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
		Name:    app.REFRESH_TOKEN_COOKIE_KEY,
		Value:   refreshToken,
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

func (a *AuthenticationController) SignUp(w http.ResponseWriter, r *http.Request) {
	const op = "AuthenticationController.SignUp"
	minTime := time.Now().Add(time.Second)
	// create user
	r.ParseForm()
	user := keycloak.User{
		FirstName: r.FormValue("first-name"),
		LastName:  r.FormValue("last-name"),
		Email:     r.FormValue("email"),
		Credentials: []keycloak.CredentialRepresentation{
			{Type: "password", Value: r.FormValue("password")},
		},
		Enabled: true,
		// TODO RequiredActions verifyEmail
	}
	err := a.userRepo.CreateUser(r.Context(), user)
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
	time.Sleep(time.Until(minTime)) // ensure loading animation lasts at least set duration

	//
	w.Write([]byte(`<div id="login-status"><small id="login-status-text">
	You're registered! Please verify your email.</small></div>`))
}

func (a *AuthenticationController) LogOut(w http.ResponseWriter, r *http.Request) {
	const op = "AuthenticationController.LogOut"
	// logout
	token, _ := r.Cookie(app.REFRESH_TOKEN_COOKIE_KEY)
	if token != nil {
		if err := a.cl.LogOut(r.Context(), token.Value); err != nil {
			log.Println(app.FromErr(err, op))
		}

		// expire cookie
		http.SetCookie(w, &http.Cookie{
			Name:    app.REFRESH_TOKEN_COOKIE_KEY,
			Value:   "",
			Path:    "/",
			Expires: time.UnixMilli(0),
		})
	}

	// redirect to login regardless of result
	w.Header().Add("HX-Redirect", "/app/login")
	w.WriteHeader(200)
}

func (a *AuthenticationController) AuthenticateToken(w http.ResponseWriter, r *http.Request) {
	token, _ := r.Cookie(app.REFRESH_TOKEN_COOKIE_KEY)
	if token != nil {
		// TODO AuthenticateToken impl
		w.WriteHeader(201)
		return
	}
	w.Header().Add("HX-Redirect", "/app/login")
	w.WriteHeader(201)
}
