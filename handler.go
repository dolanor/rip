package rip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/dolanor/rip/encoding"
)

// InputOutputFunc is a function that takes a ctx and an input, and it can return an output or an err.
type InputOutputFunc[Input, Output any] func(ctx context.Context, in Input) (out Output, err error)

// Middleware is an HTTP Middleware that you can add to your handler to handle specific actions like
// logging, authentication, authorization, metrics, ….
type Middleware func(http.HandlerFunc) http.HandlerFunc

// HandleEntity associates an urlPath with an entity provider, and handles all HTTP requests in a RESTful way:
//
//	POST   /entities/    : creates the entity
//	GET    /entities/:id : get the entity
//	PUT    /entities/:id : updates the entity (needs to pass the full entity data)
//	DELETE /entities/:id : deletes the entity
//	GET    /entities/    : lists the entities
func HandleEntity[
	Ent Entity,
	EP EntityProvider[Ent],
](
	urlPath string,
	ep EP,
	middlewares ...Middleware,
) (path string, handler http.HandlerFunc) {
	return handleEntityWithPath(urlPath, ep.Create, ep.Get, ep.Update, ep.Delete, ep.ListAll, middlewares...)
}

type (
	createFunc[Ent any]         func(ctx context.Context, ent Ent) (Ent, error)
	getFunc[ID Entity, Ent any] func(ctx context.Context, id ID) (Ent, error)
	updateFunc[Ent any]         func(ctx context.Context, ent Ent) error
	deleteFunc[ID Entity]       func(ctx context.Context, id Entity) error
	listFunc[Ent any]           func(ctx context.Context) ([]Ent, error)
)

func handleEntityWithPath[Ent Entity](urlPath string, create createFunc[Ent], get getFunc[Entity, Ent], update updateFunc[Ent], deleteFn deleteFunc[Entity], list listFunc[Ent], mids ...Middleware) (path string, handler http.HandlerFunc) {
	handler = func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(r.Method, urlPath, create)(w, r)
		case http.MethodGet:
			_, accept, editMode, err := getIDAndEditMode(w, r, r.Method, urlPath)
			if err != nil {
				writeError(w, accept, err)
				return
			}

			if urlPath == r.URL.Path && editMode == encoding.EditOff {
				handleListAll(urlPath, r.Method, list)(w, r)
				return
			}
			handleGet(urlPath, r.Method, get)(w, r)
		case http.MethodPut:
			updatePathID(urlPath, r.Method, update)(w, r)
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

func resID(requestPath, prefixPath string) stringID {
	pathID := strings.TrimPrefix(requestPath, prefixPath)

	var resID stringID
	resID.IDFromString(pathID)

	return resID
}

func checkPathID(requestPath, prefixPath string, id string) error {
	rID := resID(requestPath, prefixPath)

	if rID.IDString() != id {
		return ripError{Status: http.StatusBadRequest, Message: fmt.Sprintf("ID from URL (%s) doesn't match ID in entity (%s)", rID.IDString(), id)}
	}

	return nil
}

// decode use the content type to decode the data from r into v.
func decode[T any](r io.Reader, contentType string) (T, error) {
	var t T
	err := encoding.ContentTypeDecoder(r, contentType).Decode(&t)
	return t, err
}

func updatePathID[Ent Entity](urlPath, method string, f updateFunc[Ent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanedPath, accept, contentType, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		// TODO: use the correct encoder (www-urlform?)
		res, err := decode[Ent](r.Body, contentType)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad input format: %w", err))
			return
		}

		err = checkPathID(cleanedPath, urlPath, res.IDString())
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible entity id VS path ID: %w", err))
			return
		}

		err = f(r.Context(), res)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff).Encode(res)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func deletePathID(urlPath, method string, f deleteFunc[Entity]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanedPath, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		rID := resID(cleanedPath, urlPath)
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible entity id VS path ID: %w", err))
			return
		}

		// we don't need the returning entity, it's mostly a no-op
		err = f(r.Context(), &rID)
		if err != nil {
			var e ripError
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

		http.Redirect(w, r, urlPath, http.StatusSeeOther)
	}
}

func getIDAndEditMode(w http.ResponseWriter, r *http.Request, method string, urlPath string) (id string, accept string, editMode encoding.EditMode, err error) {
	vals := r.URL.Query()
	editMode = encoding.EditOff
	if vals.Get("mode") == "edit" {
		editMode = encoding.EditOn
	}

	cleanedPath, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
	if err != nil {
		return id, accept, editMode, err
	}

	id = strings.TrimPrefix(cleanedPath, urlPath)
	if id == "" {
		id = NewEntityID
	}
	return id, accept, editMode, nil
}

func handleGet[Ent Entity](urlPath, method string, f getFunc[Entity, Ent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, accept, editMode, err := getIDAndEditMode(w, r, method, urlPath)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		var resID stringID
		resID.IDFromString(id)

		res, err := f(r.Context(), &resID)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = encoding.AcceptEncoder(w, accept, editMode).Encode(res)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func handleListAll[Ent any](urlPath, method string, f listFunc[Ent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		ents, err := f(r.Context())
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff).Encode(ents)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func handleCreate[Ent Entity](method, urlPath string, f createFunc[Ent]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header, "Content-Type", encoding.AvailableCodecs().OrderedMimeTypes)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		res, err := decode[Ent](r.Body, contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var newEntity bool
		if res.IDString() == "0" {
			newEntity = true
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

		accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableCodecs().OrderedMimeTypes)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		if newEntity {
			http.Redirect(w, r, path.Join(urlPath, res.IDString()), http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusCreated)

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

// Handle is a generic HTTP handler that maps an HTTP method to a RequestResponseFunc f.
func Handle[Input, Output any](method string, f InputOutputFunc[Input, Output]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header, "Content-Type", encoding.AvailableCodecs().OrderedMimeTypes)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		req, err := decode[Input](r.Body, contentType)
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

		accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableCodecs().OrderedMimeTypes)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}
		err = encoding.AcceptEncoder(w, accept, encoding.EditOff).Encode(res)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
}

func badMethodHandler(w http.ResponseWriter, r *http.Request) {
	accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableCodecs().OrderedMimeTypes)
	if err != nil {
		writeError(w, accept, fmt.Errorf("bad accept header format: %w", err))
		return
	}

	writeError(w, accept, ripError{Status: http.StatusMethodNotAllowed, Message: "bad method"})
}
