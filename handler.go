package rip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Txn[Req, Res any] func(ctx context.Context, req Req) (Res, error)

type CreateFn[Res any] func(ctx context.Context, res Res) (Res, error)
type ResFn[ID IDer, Res any] func(ctx context.Context, id ID) (Res, error)
type GetFn[ID IDer, Res any] func(ctx context.Context, id ID) (Res, error)
type UpdateFn[Res any] func(ctx context.Context, res Res) error
type DeleteFn[ID IDer] func(ctx context.Context, id IDer) error

// TODO add ListFn to deal with list (pagination, etc)

type Creater[Res IDer] interface {
	Create(ctx context.Context, res Res) (Res, error)
}

type Getter[Res IDer] interface {
	Get(ctx context.Context, id IDer) (Res, error)
}

type Updater[Res IDer] interface {
	Update(ctx context.Context, res Res) error
}

type Deleter[Res IDer] interface {
	Delete(ctx context.Context, id IDer) error
}

type ResourceProvider[Res IDer] interface {
	Creater[Res]
	Getter[Res]
	Updater[Res]
	Deleter[Res]
}

func HandleResourceWithPath[Res IDer, RP ResourceProvider[Res]](urlPath string, rp RP) (path string, handler http.HandlerFunc) {
	return handleResourceWithPath(urlPath, rp.Create, rp.Get, rp.Update, rp.Delete)
}

func handleResourceWithPath[Res IDer](urlPath string, create CreateFn[Res], get GetFn[IDer, Res], updateFn UpdateFn[Res], deleteFn DeleteFn[IDer]) (path string, handler http.HandlerFunc) {
	return urlPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			HandleCreate(r.Method, create)(w, r)
		case http.MethodGet:
			HandleGet(urlPath, r.Method, get)(w, r)
		case http.MethodPut:
			UpdatePathID(urlPath, r.Method, updateFn)(w, r)
		case http.MethodDelete:
			DeletePathID(urlPath, r.Method, deleteFn)(w, r)
		default:
			badMethodHandler(w, r)
		}
	}
}

func ProcessRequest(w http.ResponseWriter, r *http.Request, method string, header http.Header) (accept, contentType string, err error) {
	accept, err = BestHeaderValue(r.Header["Accept"], AvailableEncodings)
	if err != nil {
		return "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad accept header format: %v", err)}
	}
	if r.Method != method {
		return "", "", Error{Status: http.StatusMethodNotAllowed, Message: "bad method"}
	}

	contentType, err = BestHeaderValue(r.Header["Content-Type"], AvailableEncodings)
	if err != nil {
		return "", "", Error{Status: http.StatusUnsupportedMediaType, Message: fmt.Sprintf("bad content type header format: %v", err)}
	}

	return "", contentType, nil
}

func resID(requestPath, prefixPath string, id string) stringID {
	pathID := strings.TrimPrefix(requestPath, prefixPath)

	var resID stringID
	resID.FromString(pathID)

	return resID
}

func checkPathID(requestPath, prefixPath string, id string) error {
	rID := resID(requestPath, prefixPath, id)

	if rID.IDString() != id {
		return Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("ID from URL (%s) doesn't match ID in resource (%s)", rID.IDString(), id)}
	}

	return nil
}

// Decode use the content type to decode the data from r into v.
func Decode[T any](r io.Reader, contentType string) (T, error) {
	var t T
	err := ContentTypeDecoder(r, contentType).Decode(&t)
	return t, err
}

func UpdatePathID[Res IDer](urlPath, method string, f UpdateFn[Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, contentType, err := ProcessRequest(w, r, method, r.Header)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

		res, err := Decode[Res](r.Body, contentType)
		if err != nil {
			WriteError(w, accept, fmt.Errorf("bad input format: %w", err))
			return
		}

		err = checkPathID(r.URL.Path, urlPath, res.IDString())
		if err != nil {
			WriteError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
			return
		}

		err = f(r.Context(), res)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

		err = AcceptEncoder(w, accept).Encode(res)
		if err != nil {
			WriteError(w, accept, err)
			return
		}
	}
}

func DeletePathID(urlPath, method string, f DeleteFn[IDer]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, _, err := ProcessRequest(w, r, method, r.Header)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		resID := resID(r.URL.Path, urlPath, id)
		if err != nil {
			WriteError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
			return
		}

		// we don't need the returning resource, it's mostly a no-op
		err = f(r.Context(), &resID)
		if err != nil {
			var e Error
			if errors.As(err, &e) {
				if e.Code != ErrorCodeNotFound {
					WriteError(w, accept, e)
					return
				}
			} else {
				WriteError(w, accept, err)
				return
			}
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

type IDer interface {
	IDString() string
	FromString(s string)
}

type stringID string

func (i *stringID) FromString(s string) {
	*i = stringID(s)
}

func (i stringID) IDString() string {
	return string(i)
}

func HandleGet[Res IDer](urlPath, method string, f GetFn[IDer, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, _, err := ProcessRequest(w, r, method, r.Header)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

		res, err := f(r.Context(), &resID)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

		err = AcceptEncoder(w, accept).Encode(res)
		if err != nil {
			WriteError(w, accept, err)
			return
		}

	}
}

func HandleCreate[Res any](method string, f CreateFn[Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := BestHeaderValue(r.Header["Content-Type"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		var res Res
		err = ContentTypeDecoder(r.Body, contentType).Decode(&res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err = f(r.Context(), res)
		if err != nil {
			switch e := err.(type) {
			case NotFoundError:
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		accept, err := BestHeaderValue(r.Header["Accept"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		w.WriteHeader(http.StatusCreated)
		err = AcceptEncoder(w, accept).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

	}
}

func Handle[Req, Res any](method string, f Txn[Req, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := BestHeaderValue(r.Header["Content-Type"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		var req Req
		err = ContentTypeDecoder(r.Body, contentType).Decode(&req)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		res, err := f(r.Context(), req)
		if err != nil {
			switch e := err.(type) {
			case NotFoundError:
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		accept, err := BestHeaderValue(r.Header["Accept"], AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}
		err = AcceptEncoder(w, accept).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func badMethodHandler(w http.ResponseWriter, r *http.Request) {
	accept, err := BestHeaderValue(r.Header["Accept"], AvailableEncodings)
	if err != nil {
		WriteError(w, accept, fmt.Errorf("bad accept header format: %w", err))
		return
	}

	WriteError(w, accept, Error{Status: http.StatusMethodNotAllowed, Message: "bad method"})
}
