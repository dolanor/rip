package rip

import (
	"context"
	"log"
	"net/http"
	"strings"
)

type Txn[Req, Res any] func(ctx context.Context, req Req) (Res, error)

type Identifier[T any] interface {
	Identity() T
	SetID(id T)
}

func HandleResource[Req, Res any](req Req, save Txn[Req, Res], get Txn[string, Res]) http.HandlerFunc {
	_, f := HandleResourcePath("", req, save, get)
	return f
}

func HandleResourcePath[Req, Res any](urlPath string, req Req, save Txn[Req, Res], get Txn[string, Res]) (string, http.HandlerFunc) {
	return urlPath, func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			Handle(r.Method, save)(w, r)
		case http.MethodGet:
			HandlePathID(urlPath, r.Method, get)(w, r)
		}
	}
}

func HandleID[Res any, ID Identifier[string]](method string, f Txn[string, Res]) http.HandlerFunc {
	return HandlePathID("", method, f)
}

type StringID struct {
	id string
}

func (i *StringID) Identity() string { return string(i.id) }
func (i *StringID) SetID(id string)  { i.id = id }

func HandlePathID[Res any](urlPath, method string, f Txn[string, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		id := strings.TrimPrefix(r.URL.Path, urlPath)

		var req = StringID{}
		req.SetID(id)

		log.Printf("what: whatting %+v, %s %s", req, r.URL.Path, urlPath)
		res, err := f(r.Context(), id)
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
