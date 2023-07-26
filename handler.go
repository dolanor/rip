package rip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dolanor/rip/encoding"
)

// RequestResponseFunc is a function that takes a ctx and a request, and it can return a response or an err.
type RequestResponseFunc[Request, Response any] func(ctx context.Context, request Request) (response Response, err error)

// ResourceCreater creates a resource that can be identified.
type ResourceCreater[Rsc IdentifiableResource] interface {
	Create(ctx context.Context, res Rsc) (Rsc, error)
}

// ResourceGetter gets a resource with its id.
type ResourceGetter[Rsc IdentifiableResource] interface {
	Get(ctx context.Context, id IdentifiableResource) (Rsc, error)
}

// ResourceUpdater updates an identifiable resource.
type ResourceUpdater[Rsc IdentifiableResource] interface {
	Update(ctx context.Context, res Rsc) error
}

// ResourceDeleter deletes a resource with its id.
type ResourceDeleter[Rsc IdentifiableResource] interface {
	Delete(ctx context.Context, id IdentifiableResource) error
}

// ResourceLister lists a group of resources.
type ResourceLister[Rsc any] interface {
	ListAll(ctx context.Context) ([]Rsc, error)
}

// ResourceProvider provides identifiable resources.
type ResourceProvider[Rsc IdentifiableResource] interface {
	ResourceCreater[Rsc]
	ResourceGetter[Rsc]
	ResourceUpdater[Rsc]
	ResourceDeleter[Rsc]
	ResourceLister[Rsc]
}

// HandleResource associates an urlPath with a resource provider, and handles all HTTP requests in a RESTful way:
//
//	POST   /resources/    : creates the resource
//	GET    /resources/:id : get the resource
//	PUT    /resources/:id : updates the resource (needs to pass the full resource data)
//	DELETE /resources/:id : deletes the resource
//	GET    /resources/    : lists the resources
func HandleResource[Rsc IdentifiableResource, RP ResourceProvider[Rsc]](urlPath string, rp RP, mids ...func(http.HandlerFunc) http.HandlerFunc) (path string, handler http.HandlerFunc) {
	return handleResourceWithPath(urlPath, rp.Create, rp.Get, rp.Update, rp.Delete, rp.ListAll, mids...)
}

type (
	createFunc[Rsc any]                       func(ctx context.Context, res Rsc) (Rsc, error)
	getFunc[ID IdentifiableResource, Rsc any] func(ctx context.Context, id ID) (Rsc, error)
	updateFunc[Rsc any]                       func(ctx context.Context, res Rsc) error
	deleteFunc[ID IdentifiableResource]       func(ctx context.Context, id IdentifiableResource) error
	listFunc[Rsc any]                         func(ctx context.Context) ([]Rsc, error)
)

func handleResourceWithPath[Rsc IdentifiableResource](urlPath string, create createFunc[Rsc], get getFunc[IdentifiableResource, Rsc], update updateFunc[Rsc], deleteFn deleteFunc[IdentifiableResource], list listFunc[Rsc], mids ...func(http.HandlerFunc) http.HandlerFunc) (path string, handler http.HandlerFunc) {
	handler = func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(r.Method, create)(w, r)
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

func resID(requestPath, prefixPath string, id string) stringID {
	pathID := strings.TrimPrefix(requestPath, prefixPath)

	var resID stringID
	resID.IDFromString(pathID)

	return resID
}

func checkPathID(requestPath, prefixPath string, id string) error {
	rID := resID(requestPath, prefixPath, id)

	if rID.IDString() != id {
		return ripError{Status: http.StatusBadRequest, Message: fmt.Sprintf("ID from URL (%s) doesn't match ID in resource (%s)", rID.IDString(), id)}
	}

	return nil
}

// decode use the content type to decode the data from r into v.
func decode[T any](r io.Reader, contentType string) (T, error) {
	var t T
	err := encoding.ContentTypeDecoder(r, contentType).Decode(&t)
	return t, err
}

func updatePathID[Rsc IdentifiableResource](urlPath, method string, f updateFunc[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanedPath, accept, contentType, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		// TODO: use the correct encoder (www-urlform?)
		res, err := decode[Rsc](r.Body, contentType)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad input format: %w", err))
			return
		}

		err = checkPathID(cleanedPath, urlPath, res.IDString())
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
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

func deletePathID(urlPath, method string, f deleteFunc[IdentifiableResource]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanedPath, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		id := strings.TrimPrefix(cleanedPath, urlPath)

		rID := resID(cleanedPath, urlPath, id)
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible resource id VS path ID: %w", err))
			return
		}

		// we don't need the returning resource, it's mostly a no-op
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

// IdentifiableResource is a resource that can be identified by an string.
type IdentifiableResource interface {
	// IDString returns an ID in form of a string.
	IDString() string

	// IDFromString serialize an ID from s.
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
		id = "new"
	}
	return id, accept, editMode, nil
}

func handleGet[Rsc IdentifiableResource](urlPath, method string, f getFunc[IdentifiableResource, Rsc]) http.HandlerFunc {
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

func handleListAll[Rsc any](urlPath, method string, f listFunc[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path)
		if err != nil {
			writeError(w, accept, err)
			return
		}

		rscs, err := f(r.Context())
		if err != nil {
			writeError(w, accept, err)
			return
		}

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff).Encode(rscs)
		if err != nil {
			writeError(w, accept, err)
			return
		}
	}
}

func handleCreate[Rsc IdentifiableResource](method string, f createFunc[Rsc]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header, "Content-Type", encoding.AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		res, err := decode[Rsc](r.Body, contentType)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var newResource bool
		if res.IDString() == "0" {
			newResource = true
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

		accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		if newResource {
			http.Redirect(w, r, "/users/"+res.IDString(), http.StatusSeeOther)
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
func Handle[Request, Response any](method string, f RequestResponseFunc[Request, Response]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != method {
			http.Error(w, "bad method", http.StatusMethodNotAllowed)
			return
		}

		contentType, err := bestHeaderValue(r.Header, "Content-Type", encoding.AvailableEncodings)
		if err != nil {
			http.Error(w, "bad content type header format", http.StatusBadRequest)
			return
		}

		req, err := decode[Request](r.Body, contentType)
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

		accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableEncodings)
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
	accept, err := bestHeaderValue(r.Header, "Accept", encoding.AvailableEncodings)
	if err != nil {
		writeError(w, accept, fmt.Errorf("bad accept header format: %w", err))
		return
	}

	writeError(w, accept, ripError{Status: http.StatusMethodNotAllowed, Message: "bad method"})
}
