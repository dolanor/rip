package rip

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/json"
	"github.com/dolanor/rip/internal/ripreflect"
)

// start BackendFunc OMIT

// InputOutputFunc is a function that takes a ctx and an input, and it can return an output or an err.
// It should model any generic backend function that takes input, processes it and returns an output or an error.
type InputOutputFunc[
	Input, Output any,
] func(ctx context.Context, input Input) (output Output, err error)

//end BackendFunc OMIT

// Middleware is an HTTP Middleware that you can add to your handler to handle specific actions like
// logging, authentication, authorization, metrics, ….
type Middleware = func(http.HandlerFunc) http.HandlerFunc

// start HandleEntities OMIT

// HandleEntities associates an urlPath with an entity provider, and handles all HTTP requests in a RESTful way:
//
//	POST   /entities/    : creates the entity
//	GET    /entities/:id : get the entity
//	PUT    /entities/:id : updates the entity (needs to pass the full entity data)
//	DELETE /entities/:id : deletes the entity
//	GET    /entities/    : lists the entities (accepts page and page_size query param)
//
// It also handles fields
//
//	GET    /entities/:id/name : get only the name field of the entity
//	PUT    /entities/:id/name : updates only the name entity field
func HandleEntities[
	Ent any,
	EP EntityProvider[Ent],
](
	urlPath string,
	ep EP,
	options ...EntityRouteOption,
) (path string, handler http.HandlerFunc) {
	// end HandleEntities OMIT
	var cfg entityRouteConfig
	for _, o := range options {
		o(&cfg)
	}

	if len(cfg.codecs.Codecs) == 0 {
		cfg.codecs.Register(json.Codec)
	}

	if cfg.listPageSize == 0 {
		cfg.listPageSize = 20
	}

	if cfg.listPageSizeMax == 0 {
		cfg.listPageSizeMax = 100
	}

	return handleEntityWithPath(urlPath, ep.Create, ep.Get, ep.Update, ep.Delete, ep.List, cfg)
}

type (
	createFunc[Ent any] func(ctx context.Context, ent Ent) (Ent, error)
	getFunc[Ent any]    func(ctx context.Context, id string) (Ent, error)
	updateFunc[Ent any] func(ctx context.Context, ent Ent) error
	deleteFunc          func(ctx context.Context, id string) error
	listFunc[Ent any]   func(ctx context.Context, limit, offset int) ([]Ent, error)
)

func handleEntityWithPath[Ent any](
	urlPath string,
	create createFunc[Ent],
	get getFunc[Ent],
	update updateFunc[Ent],
	deleteFn deleteFunc,
	list listFunc[Ent],
	cfg entityRouteConfig,
) (path string, handler http.HandlerFunc) {
	handler = func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			handleCreate(r.Method, urlPath, create, cfg)(w, r)
		case http.MethodGet:
			_, _, _, _, accept, editMode, err := getIDAndEditMode(w, r, r.Method, urlPath, cfg)
			if err != nil {
				writeError(w, accept, err, cfg)
				return
			}

			if urlPath == r.URL.Path && editMode == encoding.EditOff {
				handleListAll(urlPath, r.Method, list, cfg)(w, r)
				return
			}
			handleGet(urlPath, r.Method, get, cfg)(w, r)
		case http.MethodPut:
			updatePathID(urlPath, r.Method, update, get, cfg)(w, r)
		case http.MethodDelete:
			deletePathID(urlPath, r.Method, deleteFn, cfg)(w, r)
		default:
			badMethodHandler(w, r, cfg)
		}
	}

	for i := len(cfg.middlewares) - 1; i >= 0; i-- {
		// we wrap the handler in the middlewares
		handler = cfg.middlewares[i](handler)
	}

	return urlPath, handler
}

func resID(requestPath, prefixPath string) string {
	pathID := strings.TrimPrefix(requestPath, prefixPath)

	return pathID
}

// decode use the content type to decode the data from r into t.
func decode[T any](r io.Reader, contentType string, cfg entityRouteConfig) (T, error) {
	var t T
	decoder, err := encoding.ContentTypeDecoder(r, contentType, cfg.codecs)
	if err != nil {
		return t, err
	}

	err = decoder.Decode(&t)
	return t, err
}

// TODO: is it used? Delete?
func updateFieldInEntity[Ent any](entity Ent, fieldName string, fieldValue any) (err error) {
	defer func() {
		rerr := recover()
		if rerr != nil {
			// TODO(the): add a real full error with error source, etc
			err = Error{
				Source: ErrorSource{
					Pointer: fieldName,
				},
				Debug: fmt.Errorf("%v: %w", rerr, err).Error(),
			}
		}
	}()
	val := reflect.ValueOf(entity)
	// it should be a pointer type
	val = val.Elem()
	field := val.FieldByNameFunc(func(s string) bool {
		return strings.ToLower(s) == strings.ToLower(fieldName)
	})
	field.Set(reflect.ValueOf(fieldValue))

	return nil
}

func updatePathID[Ent any](urlPath, method string, f updateFunc[Ent], get getFunc[Ent], cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		//TODO add edit mode on
		id, field, _, contentType, accept, _, err := getIDAndEditMode(w, r, method, urlPath, cfg)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		var ent Ent
		if field == "" {
			// if we have no field selected, we just decode the entire entity
			ent, err = decode[Ent](r.Body, contentType, cfg)
			if err != nil {
				writeError(w, accept, fmt.Errorf("bad input format: %w", err), cfg)
				return
			}
		} else {
			err = func() (err error) {
				defer func() {
					rerr := recover()
					if rerr != nil {
						err = Error{
							Source: ErrorSource{
								Pointer: field,
							},
							Debug: fmt.Errorf("%v: %w", rerr, err).Error(),
						}
					}
				}()

				ent, err = get(r.Context(), id)
				if err != nil {
					writeError(w, accept, fmt.Errorf("can not get original entity: %w", err), cfg)
					return err
				}

				st := structOf(ent)

				fieldValue := st.FieldByNameFunc(func(s string) bool {
					return strings.ToLower(s) == strings.ToLower(field)
				})

				if !fieldValue.CanSet() {
					fieldValue = fieldValue.Elem()
					log.Println("fvalue can set: false")
					if fieldValue.CanSet() {
						log.Println("fvalue can set: true again")
					}
				}

				var fieldData any

				decoder, err := encoding.ContentTypeDecoder(r.Body, contentType, cfg.codecs)
				if err != nil {
					return err
				}

				err = decoder.Decode(&fieldData)
				if err != nil {
					writeError(w, accept, fmt.Errorf("can not decode entity field: %w", err), cfg)
					return err
				}
				fieldDataValue := reflect.ValueOf(fieldData)

				if fieldValue.CanSet() {
					fieldValue.Set(fieldDataValue)
				} else {
					writeError(w, accept, fmt.Errorf("can not set entity field: %s", fieldValue.String()), cfg)
				}

				// We've updated the field. We're good to go.
				return nil
			}()
			if err != nil {
				writeError(w, accept, fmt.Errorf("can not decode entity field: %w", err), cfg)
				return
			}
		}

		entID, err := ripreflect.GetID(ent)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		// if the user didn't put an ID in the entity as well as in the path, let's add
		// it to the entity data as well.
		if entID == "" {
			ripreflect.SetID(&ent, id)
		}

		// To update a field, we need to get the entity first, then reflect on it to get the field and change it
		// then we can update the whole entity with updateFunc
		err = f(r.Context(), ent)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		if field != "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		rrw := encoding.RequestResponseWriter{
			ResponseWriter: w,
			Request:        r,
		}

		err = encoding.AcceptEncoder(rrw, accept, encoding.EditOff, cfg.codecs).Encode(ent)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}
	}
}

func deletePathID(urlPath, method string, f deleteFunc, cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cleanedPath, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path, cfg)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		rID := resID(cleanedPath, urlPath)
		if err != nil {
			writeError(w, accept, fmt.Errorf("incompatible entity id VS path ID: %w", err), cfg)
			return
		}

		// we don't need the returning entity, it's mostly a no-op
		err = f(r.Context(), rID)
		if err != nil {
			var e Error
			if errors.As(err, &e) {
				// if the entity is not found, we don't return 404
				// we should continue with 204/200 as it is idempotent: the entity
				// the entity doesn't exist anymore.
				if e.Code != ErrorCodeNotFound {
					writeError(w, accept, e, cfg)
					return
				}
			} else {
				writeError(w, accept, err, cfg)
				return
			}
		}

		// Handle HTMX delete that returns 200 instead of HTTP 204
		// as explained in: https://htmx.org/attributes/hx-delete/
		// TODO: can I move it in the HTML encoder instead?
		if accept == "text/html" || accept == "application/x-www-form-urlencoded" {
			w.WriteHeader(http.StatusOK)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func getEntityField(entityPrefix, requestPath string) (id, field string) {
	requestPath = strings.TrimPrefix(requestPath, entityPrefix)

	requestPath = strings.TrimRight(requestPath, "/")
	requestPath = strings.TrimLeft(requestPath, "/")

	if strings.Contains(requestPath, "/") {
		id, field = path.Split(requestPath)
		id = strings.TrimRight(id, "/")
	} else {
		id = strings.TrimRight(requestPath, "/")
	}

	return id, field
}

func getIDAndEditMode(w http.ResponseWriter, r *http.Request, method string, urlPath string, cfg entityRouteConfig) (id, field, cleanedPath, contentType, accept string, editMode encoding.EditMode, err error) {
	vals := r.URL.Query()
	editMode = encoding.EditOff
	if vals.Get("mode") == "edit" {
		editMode = encoding.EditOn
	}

	cleanedPath, accept, contentType, err = preprocessRequest(r.Method, method, r.Header, r.URL.Path, cfg)
	if err != nil {
		return id, field, cleanedPath, contentType, accept, editMode, err
	}

	id = strings.TrimPrefix(cleanedPath, urlPath)
	if id == "" {
		return id, field, cleanedPath, contentType, accept, editMode, err
	}

	id, field = getEntityField(urlPath, cleanedPath)

	return id, field, cleanedPath, contentType, accept, editMode, nil
}

func structOf(v any) reflect.Value {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Pointer {
		val = val.Elem()
	}
	return val
}

func fieldValue(st reflect.Value, field string) any {
	val := st.FieldByNameFunc(func(f string) bool {
		if strings.ToLower(f) != strings.ToLower(field) {
			return false
		}
		return true
	})

	switch val.Kind() {
	case reflect.Bool:
		return val.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return val.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return val.Uint()
	case reflect.Float32, reflect.Float64:
		return val.Float()
	case reflect.Complex64, reflect.Complex128:
		return val.Complex()
	case reflect.String:
		return val.String()
	case reflect.Slice,
		reflect.Array,
		reflect.Map,
		reflect.Struct,
		reflect.Chan:
		return val.Interface()
	default:
		return val.String()
	}
}

func handleGet[Ent any](urlPath, method string, f getFunc[Ent], cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, field, _, _, accept, editMode, err := getIDAndEditMode(w, r, method, urlPath, cfg)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		res, err := f(r.Context(), id)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		var ret any = res
		if field != "" {
			st := structOf(res)
			ret = fieldValue(st, field)
		}
		rrw := encoding.RequestResponseWriter{
			ResponseWriter: w,
			Request:        r,
		}

		err = encoding.AcceptEncoder(rrw, accept, editMode, cfg.codecs).Encode(ret)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}
	}
}

func handleListAll[Ent any](urlPath, method string, f listFunc[Ent], cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		_, accept, _, err := preprocessRequest(r.Method, method, r.Header, r.URL.Path, cfg)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		pageSize := cfg.listPageSize
		if r.URL.Query().Has("page_size") {
			pageSizeErr := Error{
				Status: http.StatusBadRequest,
				Source: ErrorSource{
					Parameter: "page_size",
				},
			}

			pageSizeStr := r.URL.Query().Get("page_size")
			pageSizeUint, err := strconv.ParseUint(pageSizeStr, 10, 64)
			if err != nil {
				pageSizeErr.Detail = `malformed "page_size" query parameter`
				writeError(w, accept, pageSizeErr, cfg)
				return
			}

			if int(pageSizeUint) > cfg.listPageSizeMax {
				pageSizeErr.Detail = fmt.Sprintf(`%q query parameter cannot be bigger than: %d`, "page_size", cfg.listPageSizeMax)
				writeError(w, accept, pageSizeErr, cfg)
				return
			}

			if int(pageSizeUint) == 0 {
				pageSizeErr.Detail = fmt.Sprintf(`%q query parameter should be > 0`, "page_size")
				writeError(w, accept, pageSizeErr, cfg)
				return
			}

			pageSize = int(pageSizeUint)
		}

		offset := 0
		if r.URL.Query().Has("page") {
			pageStr := r.URL.Query().Get("page")
			page, err := strconv.ParseUint(pageStr, 10, 64)
			if err != nil {
				err := Error{
					Status: http.StatusBadRequest,
					Detail: `malformed "page" query parameter`,
					Source: ErrorSource{
						Parameter: "page",
					},
				}
				writeError(w, accept, err, cfg)
				return
			}

			// we switch to 0 index (page 1 = page 0)
			if page > 0 {
				page--
			}

			offset = int(page) * pageSize
		}

		limit := offset + pageSize

		ents, err := f(r.Context(), offset, limit)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff, cfg.codecs).Encode(ents)
		if err != nil {
			writeError(w, accept, err, cfg)
			return
		}
	}
}

func handleCreate[Ent any](method, urlPath string, f createFunc[Ent], cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, err := contentNegociateBestHeaderValue(r.Header, "Accept", cfg.codecs.OrderedMimeTypes)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad accept header format: %w", err), cfg)
			return
		}

		if r.Method != method {
			writeError(w, accept, fmt.Errorf("method not allowed: %s", r.Method), cfg)
			return
		}

		contentType, err := contentNegociateBestHeaderValue(r.Header, "Content-Type", cfg.codecs.OrderedMimeTypes)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad content type header format: %w", err), cfg)
			return
		}

		res, err := decode[Ent](r.Body, contentType, cfg)
		if err != nil {
			writeError(w, accept, fmt.Errorf("decode POST body: %w", err), cfg)
			return
		}

		res, err = f(r.Context(), res)
		if err != nil {
			writeError(w, accept, fmt.Errorf("entity provider create: %w", err), cfg)
			return
		}

		w.WriteHeader(http.StatusCreated)

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff, cfg.codecs).Encode(res)
		if err != nil {
			writeError(w, accept, fmt.Errorf("encode POST body: %w", err), cfg)
			return
		}
	}
}

// start Handle OMIT

// Handle is a generic HTTP handler that maps an HTTP method to a InputOutputFunc f.
func Handle[
	Input, Output any,
](
	method string, f InputOutputFunc[Input, Output],
	options ...EntityRouteOption,
) http.HandlerFunc {
	// end Handle OMIT

	var cfg entityRouteConfig
	for _, o := range options {
		o(&cfg)
	}

	if len(cfg.codecs.Codecs) == 0 {
		cfg.codecs.Register(json.Codec)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		accept, err := contentNegociateBestHeaderValue(r.Header, "Accept", cfg.codecs.OrderedMimeTypes)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad accept header format: %w", err), cfg)
			return
		}

		if r.Method != method {
			writeError(w, accept, fmt.Errorf("method not allowed: %s", r.Method), cfg)
			return
		}

		contentType, err := contentNegociateBestHeaderValue(r.Header, "Content-Type", cfg.codecs.OrderedMimeTypes)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad content type header format: %w", err), cfg)
			return
		}

		req, err := decode[Input](r.Body, contentType, cfg)
		if err != nil {
			writeError(w, accept, fmt.Errorf("decode %s body: %w", r.Method, err), cfg)
			return
		}

		res, err := f(r.Context(), req)
		if err != nil {
			writeError(w, accept, fmt.Errorf("handle: %w", err), cfg)
			return
		}

		err = encoding.AcceptEncoder(w, accept, encoding.EditOff, cfg.codecs).Encode(res)
		if err != nil {
			writeError(w, accept, fmt.Errorf("encode %s body: %w", r.Method, err), cfg)
			return
		}
	}
}

func badMethodHandler(w http.ResponseWriter, r *http.Request, cfg entityRouteConfig) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accept, err := contentNegociateBestHeaderValue(r.Header, "Accept", cfg.codecs.OrderedMimeTypes)
		if err != nil {
			writeError(w, accept, fmt.Errorf("bad accept header format: %w", err), cfg)
			return
		}

		writeError(w, accept, Error{Status: http.StatusMethodNotAllowed, Detail: "bad method"}, cfg)
	}
}
