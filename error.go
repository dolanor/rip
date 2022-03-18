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
	return e.Message
}

func WriteError(w http.ResponseWriter, accept string, err error) {
	var e Error
	if !errors.As(err, &e) {
		e = Error{
			Message: err.Error(),
			Status:  http.StatusInternalServerError,
		}
	}
	e.Message = err.Error()

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
