package rip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// RequestResponseFunc is a function that takes a ctx and a request, and it can return a response or an err.
type RequestResponseFunc[Request, Response any] func(ctx context.Context, request Request) (response Response, err error)

type CreateFn[Rsc any] func(ctx context.Context, res Rsc) (Rsc, error)
type ResFn[ID IdentifiableResource, Rsc any] func(ctx context.Context, id ID) (Rsc, error)
type GetFn[ID IdentifiableResource, Rsc any] func(ctx context.Context, id ID) (Rsc, error)
type UpdateFn[Rsc any] func(ctx context.Context, res Rsc) error
type DeleteFn[ID IdentifiableResource] func(ctx context.Context, id IdentifiableResource) error
type ListFn[Rsc any] func(ctx context.Context) ([]Rsc, error)

// Creater creates a resource that can be identified.
type Creater[Rsc IdentifiableResource] interface {
	Create(ctx context.Context, res Rsc) (Rsc, error)
}

// Getter gets a resource with its id.
type Getter[Rsc IdentifiableResource] interface {
	Get(ctx context.Context, id IdentifiableResource) (Rsc, error)
}

// Updater updates an identifiable resource.
type Updater[Rsc IdentifiableResource] interface {
	Update(ctx context.Context, res Rsc) error
}

// Deleter deletes a resource with its id.
type Deleter[Rsc IdentifiableResource] interface {
	Delete(ctx context.Context, id IdentifiableResource) error
}

// Lister lists a group of resources.
type Lister[Rsc any] interface {
	ListAll(ctx context.Context) ([]Rsc, error)
}

// ResourceProvider provides identifiable resources.
type ResourceProvider[Rsc IdentifiableResource] interface {
	Creater[Rsc]
	Getter[Rsc]
	Updater[Rsc]
	Deleter[Rsc]
	Lister[Rsc]
}

// HandleResource associates an urlPath with a resource provider, and handles all HTTP requests in a RESTful way.
func HandleResource[Rsc IdentifiableResource, RP ResourceProvider[Rsc]](urlPath string, rp RP, mids ...func(http.HandlerFunc) http.HandlerFunc) (path string, handler http.HandlerFunc) {
	return handleResourceWithPath(urlPath, rp.Create, rp.Get, rp.Update, rp.Delete, rp.ListAll, mids...)
}

func handleResourceWithPath[Rsc IdentifiableResource](urlPath string, create CreateFn[Rsc], get GetFn[IdentifiableResource, Rsc], updateFn UpdateFn[Rsc], deleteFn DeleteFn[IdentifiableResource], list ListFn[Rsc], mids ...func(http.HandlerFunc) http.HandlerFunc) (path string, handler http.HandlerFunc) {
	handler = func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(r.Method, create)(w, r)
		case http.MethodGet:
			if urlPath == r.URL.Path {
				handleListAll(urlPath, r.Method, list)(w, r)
				return
			}
			handleGet(urlPath, r.Method, get)(w, r)
		case http.MethodPut:
			updatePathID(urlPath, r.Method, updateFn)(w, r)
		case http.MethodDelete:
			deletePathID(urlPath, r.Method, deleteFn)(w, r)
		default:
			badMethodHandler(w, r)
		}
	}

	for i := len(mids) - 1; i >= 0; i-- {
		// we wrap the handler in the middlewares
		handler = mids[i](handler)

	}

	return urlPath, handler
}

func resID(requestPath, prefixPath string, id string) stringID {
	pathID := strings.TrimPrefix(requestPath, prefixPath)

	var resID stringID
	resID.IDFromString(pathID)

	return resID
}

func checkPathID(requestPath, prefixPath string, id string) error {
	rID := resID(requestPath, prefixPath, id)

	if rID.IDString() != id {
		return Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("ID from URL (%s) doesn't match ID in resource (%s)", rID.IDString(), id)}
	}

	return nil
}

// decode use the content type to decode the data from r into v.
func decode[T any](r io.Reader, contentType string) (T, error) {
	var t T
	err := contentTypeDecoder(r, contentType).Decode(&t)
	return t, err
}

func updatePathID[Rsc IdentifiableResource](urlPath, method string, f UpdateFn[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, contentType, err := preprocessRequest(r.Method, method, r.Header)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		res, err := decode[Rsc](r.Body, contentType)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad input format: %w", err))
			return
		}

		err = checkPathID(r.URL.Path, urlPath, res.IDString())
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
			return
		}

		err = f(r.Context(), res)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = acceptEncoder(w, accept).Encode(res)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func deletePathID(urlPath, method string, f DeleteFn[IdentifiableResource]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, _, err := preprocessRequest(r.Method, method, r.Header)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		resID := resID(r.URL.Path, urlPath, id)
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
			return
		}

		// we don't need the returning resource, it's mostly a no-op
		err = f(r.Context(), &resID)
		if err != nil {
			var e Error
			if errors.As(err, &e) {
				if e.Code != ErrorCodeNotFound {
					writeError(w, accept, e)
					return
				}
			} else {
				writeError(w, accept, err)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// IdentifiableResource is a resource that can be identifiable by an string.
type IdentifiableResource interface {
	IDString() string
	IDFromString(s string) error
}

type stringID string

func (i *stringID) IDFromString(s string) error {
	*i = stringID(s)
	return nil
}

func (i stringID) IDString() string {
	return string(i)
}

func handleGet[Rsc IdentifiableResource](urlPath, method string, f GetFn[IdentifiableResource, Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, _, err := preprocessRequest(r.Method, method, r.Header)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.IDFromString(id)

		res, err := f(r.Context(), &resID)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = acceptEncoder(w, accept).Encode(res)
		if err != nil {
			writeError(w, accept, err)
			return
		}

	}
}

func handleListAll[Rsc any](urlPath, method string, f ListFn[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, _, err := preprocessRequest(r.Method, method, r.Header)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		rscs, err := f(r.Context())
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = acceptEncoder(w, accept).Encode(rscs)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func handleCreate[Rsc any](method string, f CreateFn[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header["Content-Type"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		var res Rsc
		err = contentTypeDecoder(r.Body, contentType).Decode(&res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err = f(r.Context(), res)
		if err != nil {
			switch e := err.(type) {
			case notFoundError:
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		accept, err := bestHeaderValue(r.Header["Accept"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = acceptEncoder(w, accept).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

// Handle is a generic HTTP handler that takes a
func Handle[Req, Rsp any](method string, f RequestResponseFunc[Req, Rsp]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header["Content-Type"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		var req Req
		err = contentTypeDecoder(r.Body, contentType).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := f(r.Context(), req)
		if err != nil {
			switch e := err.(type) {
			case notFoundError:
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		accept, err := bestHeaderValue(r.Header["Accept"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}
		err = acceptEncoder(w, accept).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func badMethodHandler(w http.ResponseWriter, r *http.Request) {
	accept, err := bestHeaderValue(r.Header["Accept"], AvailableEncodings)
	if err != nil {
		writeError(w, accept, fmt.Errorf("bad accept header format: %w", err))
		return
	}

	writeError(w, accept, Error{Status: http.StatusMethodNotAllowed, Message: "bad method"})
}
