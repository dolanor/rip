package rip

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type ErrorCode int

const (
	ErrorCodeBadQArg ErrorCode = 499
)

type Error struct {
	Status  int
	Err     error
	Code    ErrorCode
	Message string
}

func (e Error) MarshalJSON() ([]byte, error) {
	type ex struct {
		Status  int
		Err     string
		Code    int
		Message string
	}
	exx := ex{
		Status:  e.Status,
		Err:     e.Err.Error(),
		Code:    int(e.Code),
		Message: e.Message,
	}
	return json.Marshal(exx)
}

func (e Error) Error() string {
	return e.Err.Error() + " " + e.Message
}

func WriteError(w http.ResponseWriter, accept string, err error, msg string) {
	var e Error

	if !errors.As(err, &e) {
		e = Error{
			Err:     err,
			Message: msg,
			Status:  http.StatusInternalServerError,
		}
	}

	var eee BadRequestError
	if e.Code == ErrorCodeBadQArg || errors.As(err, &eee) {
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

type BadRequestError struct {
	origin error
}

func (e BadRequestError) Error() string {
	return "bad request: " + e.origin.Error()
}

func (e BadRequestError) Unwrap() error {
	err := errors.Unwrap(e.origin)
	if err != nil {
		return err
	}
	return nil
}

func BadRequestErr(err error) error {
	var e BadRequestError
	e.origin = fmt.Errorf("BAD: %w", err)
	return e
}
