package app

import (
	"bytes"
	"fmt"
)

type Error interface {
	error
	StatusCode() int
	Status() string
	Message() string
}

type err struct {
	code   int
	status string

	Op  string
	msg string
	Err error
}

var _ Error = (*err)(nil)

func (e *err) Error() string {
	return fmt.Sprintf("%s: %d %s", e.Message(), e.code, e.status)
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

func (e *err) Message() string {
	var buf bytes.Buffer
	if e.Op != "" {
		fmt.Fprint(&buf, e.Op)
	}
	if e.msg != "" {
		fmt.Fprintf(&buf, ": %s", e.msg)
	}
	if e.Err != nil {
		fmt.Fprintf(&buf, ": %s", e.Err.Error())
	}

	return buf.String()
}

func NewErr(code int, status, msg string) *err {
	return &err{
		code:   code,
		status: status,
		msg:    msg,
	}
}

func FromErr(e error, op string) *err {
	if e == nil {
		return nil
	}

	res := &err{Op: op}
	err, _ := e.(Error)
	if err != nil {
		res.code = err.StatusCode()
		res.status = err.Status()
		res.msg = err.Message()
	} else {
		res.Err = e
	}

	return res
}
