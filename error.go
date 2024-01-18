package app

import (
	"bytes"
	"fmt"
	"net/http"
)

type Error interface {
	error
	StatusCode() int
	Status() string
}

type err struct {
	code   int
	status string

	Op  string
	Msg string
	Err error
}

func (e *err) Error() string {
	var buf bytes.Buffer
	if e.Op != "" {
		fmt.Fprint(&buf, e.Op)
	}
	if e.Msg != "" {
		fmt.Fprintf(&buf, ": %s", e.Msg)
	}
	if e.Err != nil {
		fmt.Fprintf(&buf, ": %s", e.Msg)
	}
	return buf.String()
}

func (e *err) StatusCode() int {
	if e.code == 0 {
		return 500
	}
	return e.code
}

func (e *err) Status() string {
	return e.status
}

func NewErr(code int, status, msg string) *err {
	return &err{
		code:   code,
		status: status,
		Msg:    msg,
	}
}

func FromErr(e error, op string) *err {
	if e == nil {
		return nil
	}

	code := http.StatusInternalServerError
	status := ""
	httperr, _ := e.(Error)
	if httperr != nil {
		code = httperr.StatusCode()
		status = httperr.Status()
	}

	return &err{
		code:   code,
		status: status,
		Msg:    e.Error(),
		Op:     op,
	}
}
