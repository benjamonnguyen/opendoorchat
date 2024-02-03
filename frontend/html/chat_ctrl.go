package html

import (
	"net/http"
	"strings"

	"github.com/benjamonnguyen/opendoorchat/frontend/be"
	"github.com/julienschmidt/httprouter"
)

type ChatController struct {
	cl *be.Client
}

func NewChatController(cl *be.Client) *ChatController {
	return &ChatController{
		cl: cl,
	}
}

func (ctrl *ChatController) ChatView(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	//
	http.ServeFile(w, r, "frontend/public/new-chat.html")
}

func (ctrl *ChatController) CreateChat(
	w http.ResponseWriter,
	r *http.Request,
	p httprouter.Params,
) {
	// TODO TODO TODO
	const op = "ChatController.CreateChat"
	// get params
	r.ParseForm()
	var (
		recipients []string
		subject    = r.FormValue("subject")
		text       = r.FormValue("text")
	)
	for _, r := range strings.Split(r.FormValue("recipients"), ",") {
		recipients = append(recipients, r)
	}

	// introspect to authenticate and get user info

	// create chat

	// ctrl.cl.CreateChat()

	// add message

	// respond

}
