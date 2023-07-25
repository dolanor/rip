package rip

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/dolanor/rip/encoding"
)

// ErrorCode maps errors from the ResourceProvider implementation to HTTP
// status code.
type ErrorCode int

const (
	// ErrorCodeNotFound happens when a resource with an id is not found.
	ErrorCodeNotFound ErrorCode = http.StatusNotFound

	// ErrorCodeBadQArg happens when a user gives a wrongly formatted header `; q=X.Y` argument.
	ErrorCodeBadQArg ErrorCode = 499
)

// Error is the error returned by rip.
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
	if errors.Is(err, encoding.ErrNoEncoderAvailable) {
		e.Status = http.StatusNotAcceptable
	}

	encoder := encoding.AcceptEncoder(w, accept, encoding.EditOff)
	if e.Status == http.StatusNotAcceptable {
		// if we have encoding problems, we will use json as default
		// to serialize the error to user
		encoder = json.NewEncoder(w)
	}

	w.WriteHeader(e.Status)
	err = encoder.Encode(e)
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
