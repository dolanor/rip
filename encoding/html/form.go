package html

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ajg/form"
	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/codecwrap"
)

// NewEntityFormCodec creates a HTML Form codec that uses pathPrefix for links creation.
// It will generate a form with editable inputs for each field of your [rip.Entity].
func NewEntityFormCodec(pathPrefix string) encoding.Codec {
	// TODO: should have a better design so the path shouldn't be passed many times around.
	return codecwrap.Wrap(NewFormEncoder(pathPrefix), form.NewDecoder, FormMimeTypes...)
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

//go:embed entity_form.gotpl
var entityFormTmpl string

type FormEncoder struct {
	w          io.Writer
	pathPrefix string
}

func NewFormEncoder(pathPrefix string) func(w io.Writer) *FormEncoder {
	return func(w io.Writer) *FormEncoder {
		return &FormEncoder{
			w:          w,
			pathPrefix: pathPrefix,
		}
	}
}

func (e FormEncoder) Encode(v interface{}) error {
	return htmlEncode(e.pathPrefix, e.w, editOn, v)
}

func htmlEncode(pathPrefix string, w io.Writer, edit editMode, v interface{}) error {
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

	var entities []entity
	var entityName string
	if s.Kind() == reflect.Slice {
		for i := 0; i < s.Len(); i++ {
			s := s.Index(i)

			res := expandFields(s)
			if entityName == "" {
				entityName = res.Name
			}
			entities = append(entities, res)
		}
	} else {
		res := expandFields(s)
		if entityName == "" {
			entityName = res.Name
		}
		entities = append(entities, res)
	}

	pd := pageData{
		PathPrefix: pathPrefix,
		EntityName: entityName,
		Entities:   entities,
	}

	rw, ok := w.(http.ResponseWriter)
	if ok {
		// We force Content-Type to HTML
		rw.Header().Set("Content-Type", "text/html")
	}

	tpl := template.New("entity").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
		"editModeQueryParam": func() string {
			return editModeQueryParam
		},
		"toString": func(a any) string {
			return fmt.Sprintf("%v", a)
		},
	})

	tmplSrc := entityTmpl
	if edit {
		tmplSrc = entityFormTmpl
	}

	tpl, err = tpl.Parse(tmplSrc)
	if err != nil {
		return err
	}

	err = tpl.ExecuteTemplate(w, "entity", pd)
	if err != nil {
		return err
	}

	return nil
}

type field struct {
	Key   string
	Value any
	Type  string
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

	res := entity{
		Name: name,
	}
	switch s.Kind() {
	case reflect.String:
		res.Fields = append(res.Fields, field{"value", s.String(), "string"})
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
			if fName == "ID" {
				res.ID = fVal
			}
			res.Fields = append(res.Fields, field{fName, fVal, fTypeStr})
		}
	default:
		panic("reflect type not handled, yet: " + s.Kind().String())
	}

	return res
}
