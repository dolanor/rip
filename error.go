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

	// ErrorCodeNotImplemented is when the endpoint is not implemented.
	ErrorCodeNotImplemented ErrorCode = http.StatusNotImplemented

	// errorCodeBadQArg happens when a user gives a wrongly formatted header `; q=X.Y` argument.
	errorCodeBadQArg ErrorCode = 499
)

var (
	// ErrNotFound represents when a resource is not found.
	// It can also be used if a user without proper authorization
	// should not know if a resource exists or not.
	ErrNotFound = Error{
		Code:   ErrorCodeNotFound,
		Status: http.StatusNotFound,
		Detail: "entity not found",
	}

	// ErrNotImplemented communicates if a specific entity function is not
	// implemented.
	ErrNotImplemented = Error{
		Code:   ErrorCodeNotImplemented,
		Status: http.StatusNotImplemented,
		Detail: "not implemented",
	}
)

// Error is the error returned by rip.
// It is inspired by JSON-API.
type Error struct {
	// ID is a unique identifier for this particular occurrence of the problem.
	ID string `json:"id,omitempty"`

	// Links can contains an About Link or a Type Link.
	Links []ErrorLink `json:"links,omitempty"`

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

	// Debug contains debug information, not to be read by a user of the app, but by a technical user trying to fix problems.
	Debug string `json:"debug,omitempty"`
}

// ErrorSource indicates the source error.
// It is based on the JSON API specification: https://jsonapi.org/format/#error-objects
type ErrorSource struct {
	// Pointer is a JSON Pointer [RFC6901] to the value in the request document
	// that caused the error [e.g. "/data" for a primary data object,
	// or "/data/attributes/title" for a specific attribute].
	// This MUST point to a value in the request document that exists;
	// if it doesn’t, the client SHOULD simply ignore the pointer.
	Pointer string `json:"pointer,omitempty"`

	// Parameter indicates which URI query parameter caused the error.
	Parameter string `json:"parameter,omitempty"`

	// Header indicates the name of a single request header which caused the error.
	Header string `json:"header,omitempty"`
}

func (e Error) Error() string {
	return fmt.Sprintf("%d - %s", e.Code, e.Detail)
}

func writeError(w http.ResponseWriter, accept string, err error, options *RouteOptions) {
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
	if e.Code == errorCodeBadQArg || errors.As(err, &bre) {
		e.Status = http.StatusBadRequest
	}
	var nfe notFoundError
	if e.Code == ErrorCodeNotFound || errors.As(err, &nfe) {
		e.Status = http.StatusNotFound
	}
	if errors.Is(err, encoding.ErrNoEncoderAvailable) {
		e.Status = http.StatusNotAcceptable
	}

	encoder := encoding.AcceptEncoder(w, accept, encoding.EditOff, options.codecs)
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

// ErrorLink represents a RFC8288 web link.
type ErrorLink struct {
	// HRef is a URI-reference [RFC3986 Section 4.1] pointing to the link’s target.
	HRef string `json:"href,omitempty"`

	// Rel indicates the link’s relation type. The string MUST be a valid link relation type.
	Rel string `json:"rel,omitempty"`

	// DescribedBy is a link to a description document (e.g. OpenAPI or JSON Schema) for the link target.
	DescribedBy *ErrorLink `json:"describedby,omitempty"`

	// Title serves as a label for the destination of a link such that it can be used as a human-readable identifier (e.g., a menu entry).
	Title string `json:"title,omitempty"`

	// Type indicates the media type of the link’s target.
	Type string `json:"type,omitempty"`

	// HRefLang indicates the language(s) of the link’s target. An array of strings indicates that the link’s target is available in multiple languages. Each string MUST be a valid language tag [RFC5646].
	HRefLang []string `json:"hreflang,omitempty"`
}
