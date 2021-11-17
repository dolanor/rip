package rip

import (
	"context"
	"net/http"
)

type Txn[Req, Res any] func(ctx context.Context, req Req) (Res, error)

type Identifier[T any] interface {
	Identity() T
}

func HandleResource[Req, Res any, ID Identifier[string]](req Req, save Txn[Req, Res], get Txn[ID, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			Handle(r.Method, save)(w, r)
		case http.MethodGet:
			HandleID(r.Method, get)(w, r)
		}
	}
}

func HandleID[Res any, ID Identifier[string]](method string, f Txn[ID, Res]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := BestHeaderValue(r.Header["Content-Type"])
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		var id string
		err = ContentTypeDecoder(r.Body, contentType).Decode(&id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req ID
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

		accept, err := BestHeaderValue(r.Header["Accept"])
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

		contentType, err := BestHeaderValue(r.Header["Content-Type"])
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

		accept, err := BestHeaderValue(r.Header["Accept"])
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
