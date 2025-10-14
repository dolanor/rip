package html

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/codecwrap"
	"github.com/dolanor/rip/encoding/html/templates"
	"github.com/dolanor/rip/internal/ripreflect"
)

type entityMetadata struct {
	PathPrefix string
	Name       string
	Entity     any
}

const HXRequest = "Hx-Request" // it gets normalized by net/http from HX-Request to Hx-Request

// NewEntityFormCodec creates a HTML Form codec that uses pathPrefix for links creation.
// It will generate a form with editable inputs for each field of your [github.com/dolanor/rip.Entity].
func NewEntityFormCodec(pathPrefix string, opts ...Option) encoding.Codec {
	// TODO: should have a better design so the path shouldn't be passed many times around.
	return codecwrap.Wrap(NewFormEncoder(pathPrefix, opts...), form.NewDecoder, FormMimeTypes...)
}

var FormMimeTypes = []string{
	"application/x-www-form-urlencoded",
}

// editMode is a duplicate of rip.EditMode that allows to break a dependency cycle
type editMode bool

const (
	editOff editMode = false
	editOn  editMode = true
)

const editModeQueryParam = "mode=edit"

const (
	entityFormPageTmpl = "entity_form_page"
	entityFormTmpl     = "entity_form"
)

type FormEncoder struct {
	w          io.Writer
	pathPrefix string
	config     EncoderConfig
}

func NewFormEncoder(pathPrefix string, opts ...Option) func(w io.Writer) *FormEncoder {
	cfg := EncoderConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.mux == nil {
		cfg.mux = http.DefaultServeMux
	}
	serveHTMX(cfg.mux)

	return func(w io.Writer) *FormEncoder {
		return &FormEncoder{
			w:          w,
			pathPrefix: pathPrefix,
			config:     cfg,
		}
	}
}

func (e FormEncoder) Encode(v interface{}) error {
	return htmlEncode(e.pathPrefix, e.config.templatesFS, e.w, editOn, v)
}

func htmlEncode(pathPrefix string, templatesFS fs.FS, w io.Writer, edit editMode, v interface{}) error {
	err, _ := v.(error)
	if err != nil {
		// TODO: handle error better
		w.Write([]byte(err.Error()))
		return nil
	}

	s := reflect.ValueOf(v)
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
	}

	var pd any
	var entities []entity
	var entityName string
	isList := false

	if s.Kind() == reflect.Slice ||
		s.Kind() == reflect.Array {
		isList = true

		for i := 0; i < s.Len(); i++ {
			s := s.Index(i)

			ent := expandFields(s)
			if entityName == "" {
				entityName = ent.Name
			}
			entities = append(entities, ent)

			pd = pageData{
				PathPrefix: pathPrefix,
				EntityName: entityName,

				Entities: entities,
			}
		}
	} else {
		res := expandFields(s)
		if entityName == "" {
			entityName = res.Name
		}

		pd = entityMetadata{
			PathPrefix: pathPrefix,
			Name:       entityName,

			Entity: res,
		}
	}

	rw, ok := w.(http.ResponseWriter)
	if ok {
		// We force Content-Type to HTML
		rw.Header().Set("Content-Type", "text/html")
	}

	isHTMX := isHTMXRequest(w)
	tmplSrc := selectTemplate(edit, isList, isHTMX)

	tpl := template.New("html").Funcs(template.FuncMap{
		"idField":     ripreflect.FieldIDString,
		"idFieldName": ripreflect.FieldIDName,
		"toLower":     strings.ToLower,
		"editModeQueryParam": func() string {
			return editModeQueryParam
		},
		"toString": func(a any) string {
			return fmt.Sprintf("%v", a)
		},
		"wrapMetadata": func(pathPrefix string, a any) any {
			return entityMetadata{
				PathPrefix: pathPrefix,

				Entity: a,
			}
		},
	})

	if templatesFS == nil {
		templatesFS = templates.Files
	}
	tpl, err = tpl.ParseFS(templatesFS, "*.gotpl")
	if err != nil {
		return err
	}

	err = tpl.ExecuteTemplate(w, tmplSrc, pd)
	if err != nil {
		return err
	}

	return nil
}

func isHTMXRequest(w io.Writer) bool {
	rrw, ok := w.(encoding.RequestResponseWriter)
	if !ok {
		return false
	}

	htmxReq, ok := rrw.Request.Header[HXRequest]
	if !ok {
		return false
	}
	isHTMXStr := strings.ToLower(htmxReq[0])

	return isHTMXStr == "true"
}

func selectTemplate(edit editMode, isList, isHTMX bool) string {
	tmplSrc := entityPageTmpl

	switch {
	case isList:
		tmplSrc = entityListPageTmpl
	case edit == true:
		tmplSrc = entityFormPageTmpl
	case isHTMX && edit == true:
		tmplSrc = entityFormTmpl
	case isHTMX && edit == false:
		tmplSrc = entityTmpl
	}

	return tmplSrc
}

type field struct {
	Key   string
	Value any
	Type  string
	IsID  bool
}

type entity struct {
	ID     any
	Name   string
	Fields []field
}

type pageData struct {
	PathPrefix string
	EntityName string

	Entities []entity
}

func expandFields(s reflect.Value) entity {
	t := s.Type()
	_, name, _ := strings.Cut(t.String(), ".")
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
		t = s.Type()
	}

	ent := entity{
		Name: name,
	}

	switch s.Kind() {
	case reflect.String:
		ent.Fields = append(ent.Fields, field{"value", s.String(), "string", false})
	case reflect.Struct:
		for i := 0; i < s.NumField(); i++ {
			f := s.Field(i)
			fName := t.Field(i).Name
			fVal := f.Interface()

			fTypeStr := ""
			switch f.Type() {
			case reflect.TypeOf(time.Time{}):
				fTypeStr = "time.Time"
			}
			if f.Type() == reflect.TypeOf(time.Time{}) {
				fVal = f.Interface().(time.Time).Format(time.RFC3339)
			}
			var isID bool
			if fName == "ID" || ripreflect.HasRIPIDField(t.Field(i)) {
				// TODO: check if it works correctly, even when a struct has an .ID field and
				// a field with a `rip:"id"` struct tag somewhere else.
				ent.ID = fVal
				isID = true
			}
			ent.Fields = append(ent.Fields, field{fName, fVal, fTypeStr, isID})
		}
	default:
		panic("reflect type not handled, yet: " + s.Kind().String())
	}

	return ent
}
