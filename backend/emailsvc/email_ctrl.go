package emailsvc

import (
	"encoding/json"
	"net/http"
)

type EmailController interface {
	ThreadSearch(http.ResponseWriter, *http.Request)
}

var _ EmailController = (*emailController)(nil)

type emailController struct {
	service EmailService
}

func NewEmailController(service EmailService) *emailController {
	return &emailController{
		service: service,
	}
}

func (ctrl *emailController) ThreadSearch(w http.ResponseWriter, r *http.Request) {
	// decode search terms
	var st ThreadSearchTerms
	if err := json.NewDecoder(r.Body).Decode(&st); err != nil {
		http.Error(w, "provide ThreadSearchTerms", http.StatusBadRequest)
		return
	}

	//
	thread, httperr := ctrl.service.ThreadSearch(r.Context(), st)
	if httperr != nil {
		http.Error(w, "failed ThreadSearch: "+httperr.Error(), httperr.StatusCode())
		return
	}

	//
	data, err := json.Marshal(thread)
	if err != nil {
		http.Error(w, "failed Marshal: "+err.Error(), 500)
		return
	}

	//
	w.Header().Add("Content-Type", "application/json")
	w.Write(data)
}
