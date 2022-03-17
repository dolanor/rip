package rip

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Txn[Req, Res any] func(ctx context.Context, req Req) (Res, error)

type CreateFn[Res any] func(ctx context.Context, res Res) (Res, error)
type ResFn[ID IDer, Res any] func(ctx context.Context, id ID) (Res, error)
type GetFn[ID IDer, Res any] func(ctx context.Context, id ID) (Res, error)
type UpdateFn[Res any] func(ctx context.Context, res Res) error
type DeleteFn[ID IDer] func(ctx context.Context, id IDer) error

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
		}
	}
}

func UpdatePathID[Res IDer](urlPath, method string, f UpdateFn[Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

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

		if resID.IDString() != res.IDString() {
			http.Error(w, fmt.Sprintf("ID from URL (%s) doesn't match ID in resource (%s)", resID.IDString(), res.IDString()), http.StatusBadRequest)
			return
		}

		err = f(r.Context(), res)
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

func DeletePathID(urlPath, method string, f DeleteFn[IDer]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

		// we don't need the returning resource, it's mostly a no-op
		err := f(r.Context(), &resID)
		if err != nil {
			switch e := err.(type) {
			case NotFoundError:
				http.Error(w, e.Error(), http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func HandleID[ID IDer, Res IDer](method string, f Txn[ID, Res]) http.HandlerFunc {
	//TODO what to use for default path?
	//return HandlePathID("", method, f)
	//TODO fix
	return nil
}

type StringID struct {
	id string
}

func (i *StringID) Identity() string { return string(i.id) }
func (i *StringID) SetID(id string)  { i.id = id }

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
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

		res, err := f(r.Context(), &resID)
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
