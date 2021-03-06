package rip

import (
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	ErrorCodeNotFound ErrorCode = 404
	ErrorCodeBadQArg  ErrorCode = 499
)

type Error struct {
	Status  int
	Code    ErrorCode
	Message string
}

func (e Error) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Message)
}

func writeError(w http.ResponseWriter, accept string, err error) {
	var e Error
	if !errors.As(err, &e) {
		e = Error{
			Message: err.Error(),
		}
	}

	if e.Status == 0 {
		e.Status = http.StatusInternalServerError
	}

	e.Message = err.Error()

	var bre badRequestError
	if e.Code == ErrorCodeBadQArg || errors.As(err, &bre) {
		e.Status = http.StatusBadRequest
	}
	var nfe notFoundError
	if e.Code == ErrorCodeNotFound || errors.As(err, &nfe) {
		e.Status = http.StatusNotFound
	}

	w.WriteHeader(e.Status)
	err = acceptEncoder(w, accept).Encode(e)
	// We can't do anything, we need to make the HTTP server intercept the panic
	if err != nil {
		panic(err)
	}
}

type notFoundError struct {
	Resource string
}

func (e notFoundError) Error() string {
	return "resource not found: " + e.Resource
}

type badRequestError struct {
	origin error
}

func (e badRequestError) Error() string {
	return "bad request: " + e.origin.Error()
}

func (e badRequestError) Unwrap() error {
	err := errors.Unwrap(e.origin)
	if err != nil {
		return err
	}
	return nil
}
