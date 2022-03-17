package rip

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type Txn[Req, Res any] func(ctx context.Context, req Req) (Res, error)

func HandleResource[Req, Res any](req Req, save Txn[Req, Res], get Txn[IDer, Res], deleteFn Txn[IDer, Res]) http.HandlerFunc {
	_, f := HandleResourcePath("", req, save, get, deleteFn)
	return f
}

func HandleResourcePath[Req, Res any](urlPath string, req Req, save Txn[Req, Res], get Txn[IDer, Res], deleteFn Txn[IDer, Res]) (path string, handler http.HandlerFunc) {
	return urlPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			Handle(r.Method, save)(w, r)
		case http.MethodGet:
			HandlePathID(urlPath, r.Method, get)(w, r)
		case http.MethodDelete:
			DeletePathID(urlPath, r.Method, deleteFn)(w, r)
		}
	}
}

func DeletePathID[Res any](urlPath, method string, f Txn[IDer, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

		log.Printf("delete: whatting %+v, %s %s", resID, r.URL.Path, urlPath)
		// we don't need the returning resource, it's mostly a no-op
		_, err := f(r.Context(), &resID)
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

func HandleID[Res any](method string, f Txn[IDer, Res]) http.HandlerFunc {
	//TODO what to use for default path?
	return HandlePathID("", method, f)
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

func HandlePathID[Res any](urlPath, method string, f Txn[IDer, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var resID stringID
		resID.FromString(id)

		log.Printf("what: whatting %+v, %s %s", resID, r.URL.Path, urlPath)
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
