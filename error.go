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

var (
	ErrNotFound = Error{
		Code:   ErrorCodeNotFound,
		Status: http.StatusNotFound,
		Detail: "entity not found",
	}
)

// Error is the error returned by rip.
// It is inspired by JSON-API.
type Error struct {
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `json:"id,omitempty"`

	// Links can contains an About Link or a Type Link.
	Links []Link `json:"links,omitempty"`

	// Status is the HTTP status code applicable to this problem. This SHOULD be provided.
	Status int `json:"status,omitempty"`

	// Code is an application-specific error code.
	Code ErrorCode `json:"code,omitempty"`

	// Title is a short, human-readable summary of the problem that SHOULD NOT change from occurrence to occurrence of the problem, except for purposes of localization.
	Title string `json:"title,omitempty"`

	// Detail is a human-readable explanation specific to this occurrence of the problem
	Detail string `json:"detail,omitempty"`

	// Source is an object containing references to the primary source of the error. It SHOULD include one of its member or be omitted.
	Source ErrorSource `json:"source,omitempty"`
}

// ErrorSource indicates the source error.
// It is based on the JSON API specification.
type ErrorSource struct {
	Pointer   string `json:omitempty`
	Parameter string `json:omitempty`
	Header    string `json:omitempty`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Detail)
}

func writeError(w http.ResponseWriter, accept string, err error) {
	var e Error
	if !errors.As(err, &e) {
		e = Error{
			Detail: err.Error(),
		}
	}

	if e.Status == 0 {
		e.Status = http.StatusInternalServerError
	}

	e.Detail = err.Error()

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
	return "entity not found: " + e.Resource
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
