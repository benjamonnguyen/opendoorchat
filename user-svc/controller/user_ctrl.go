package controller

import (
	"encoding/json"
	"net/http"

	"github.com/benjamonnguyen/gootils/devlog"
	"github.com/benjamonnguyen/opendoor-chat/user-svc/service"
	"github.com/julienschmidt/httprouter"
)

type UserController interface {
	Authenticate(http.ResponseWriter, *http.Request, httprouter.Params)
	// TODO CreateUser
}

type userCtrl struct {
	service service.UserService
}

func NewUserController(service service.UserService) *userCtrl {
	return &userCtrl{
		service: service,
	}
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
	devlog.Printf("userctrl.Authenticate: %+v\n", body)
	if body.Email == "" || body.Password == "" {
		http.Error(w, "required email or password is blank", http.StatusBadRequest)
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
