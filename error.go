package rip

import (
	"errors"
	"fmt"
	"net/http"
)

type Error struct {
	Status  int
	Err     error
	Message string
}

func (e Error) Error() string {
	return e.Err.Error()
}

func WriteError(w http.ResponseWriter, accept string, err error, msg string) {
	e := Error{
		Err:     err,
		Message: msg,
		Status:  http.StatusInternalServerError,
	}

	var eee BadRequestError
	if errors.As(err, &eee) {
		e.Status = http.StatusBadRequest
	}
	var ee NotFoundError
	if errors.As(err, &ee) {
		e.Status = http.StatusNotFound
	}

	err = AcceptEncoder(w, accept).Encode(e)
	// We can't do anything, we need to make the HTTP server intercept the panic
	if err != nil {
		panic(err)
	}
}

type NotFoundError struct {
	Resource string
}

func (e NotFoundError) Error() string {
	return "resource not found: " + e.Resource
}
