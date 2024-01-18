package emailsvc

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type EmailController interface {
	ThreadSearch(http.ResponseWriter, *http.Request, httprouter.Params)
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

func (ctrl *emailController) ThreadSearch(
	w http.ResponseWriter,
	req *http.Request,
	_ httprouter.Params,
) {
	// decode search terms
	var st ThreadSearchTerms
	if err := json.NewDecoder(req.Body).Decode(&st); err != nil {
		http.Error(w, "provide ThreadSearchTerms", http.StatusBadRequest)
		return
	}

	//
	thread, httperr := ctrl.service.ThreadSearch(req.Context(), st)
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
