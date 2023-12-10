package httputil

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
)

type HttpError interface {
	error
	StatusCode() int
}

type httpError struct {
	err  error
	code int
}

func (e httpError) StatusCode() int {
	return e.code
}

func (e httpError) Error() string {
	return fmt.Sprintf("status code %d: %s", e.code, e.err.Error())
}

func NewHttpError(statusCode int, msg string) httpError {
	return httpError{
		err:  errors.New(msg),
		code: statusCode,
	}
}

func HttpErrorFromErr(err error) httpError {
	code := http.StatusInternalServerError
	re := regexp.MustCompile(`status code: (\d+)`)
	match := re.FindStringSubmatch(err.Error())
	if len(match) > 1 {
		c, _ := strconv.Atoi(match[1])
		code = c
	}
	return httpError{
		err:  err,
		code: code,
	}
}
