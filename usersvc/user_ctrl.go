package usersvc

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"
)

type UserController interface {
	Authenticate(http.ResponseWriter, *http.Request, httprouter.Params)
	CreateUser(http.ResponseWriter, *http.Request, httprouter.Params)
}

type userCtrl struct {
	service UserService
}

func NewUserController(service UserService) *userCtrl {
	return &userCtrl{
		service: service,
	}
}

func (ctrl *userCtrl) CreateUser(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// validate body
	var user User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		log.Error().Err(err).Msg("CreateUser: failed decode")
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if err := user.Validate(); err != nil {
		log.Error().Err(err).Interface("user", user).Msg("CreateUser: failed validate")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// hash password
	hashedPw, err := bcrypt.GenerateFromPassword([]byte(user.Password), 0)
	if err != nil {
		log.Error().Err(err).Msg("CreateUser: failed GenerateFromPassword")
	}
	user.Password = string(hashedPw)

	//
	if httperr := ctrl.service.CreateUser(r.Context(), user); httperr != nil {
		http.Error(w, httperr.Status(), httperr.StatusCode())
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (ctrl *userCtrl) Authenticate(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// validate body
	var body struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if body.Email == "" || body.Password == "" {
		http.Error(w, "required email or password is empty", http.StatusBadRequest)
		return
	}

	//
	token, httperr := ctrl.service.Authenticate(r.Context(), body.Email, body.Password)
	if httperr != nil {
		http.Error(w, httperr.Status(), httperr.StatusCode())
		return
	}

	//
	w.Header().Add("Content-Type", "plain/text")
	w.Write([]byte(token))
}
