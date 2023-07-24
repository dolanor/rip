package rip

import (
	_ "embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"
)

const editModeQueryParam = "mode=edit"

type EditMode bool

const (
	EditOff EditMode = false
	EditOn  EditMode = true
)

//go:embed resource_form.gotpl
var resourceFormTmpl string

type htmlFormEncoder struct {
	w io.Writer
}

func newHTMLFormEncoder(w io.Writer) *htmlFormEncoder {
	return &htmlFormEncoder{
		w: w,
	}
}

func (e htmlFormEncoder) Encode(v interface{}) error {
	return htmlEncode(e.w, EditOn, v)
}

type htmlEncoder struct {
	w io.Writer
}

func (e htmlEncoder) Encode(v interface{}) error {
	return htmlEncode(e.w, EditOff, v)
}

func htmlEncode(w io.Writer, edit EditMode, v interface{}) error {
	s := reflect.ValueOf(v)
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
	}

	var resources []resource
	var resourceName string
	if s.Kind() == reflect.Slice {
		for i := 0; i < s.Len(); i++ {
			s := s.Index(i)

			res := expandFields(s)
			if resourceName == "" {
				resourceName = res.Name
			}
			resources = append(resources, res)
		}
	} else {
		res := expandFields(s)
		if resourceName == "" {
			resourceName = res.Name
		}
		resources = append(resources, res)
	}

	pd := pageData{
		ResourceName: resourceName,
		Resources:    resources,
	}

	rw, ok := w.(http.ResponseWriter)
	if ok {
		// We force Content-Type to HTML
		rw.Header().Set("Content-Type", "text/html")
	}

	tpl := template.New("resource").Funcs(template.FuncMap{
		"toLower": strings.ToLower,
		// FIXME: use real plural i18n lib
		"toPlural": func(s string) string {
			return s + "s"
		},
		"editModeQueryParam": func() string {
			return editModeQueryParam
		},
		"toString": func(a any) string {
			return fmt.Sprintf("%v", a)
		},
	})

	tmplSrc := resourceTmpl
	if edit {
		tmplSrc = resourceFormTmpl
	}

	tpl, err := tpl.Parse(tmplSrc)
	if err != nil {
		return err
	}

	err = tpl.ExecuteTemplate(w, "resource", pd)
	if err != nil {
		return err
	}

	return nil
}

type field struct {
	Key   string
	Value any
}

type resource struct {
	ID     any
	Name   string
	Fields []field
}

type pageData struct {
	PagePath     string
	ResourceName string

	Resources []resource
}

func expandFields(s reflect.Value) resource {
	t := s.Type()
	// log.Println("encode: name:", name)
	_, name, _ := strings.Cut(t.String(), ".")
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
		t = s.Type()
	}

	res := resource{
		Name: name,
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fName := t.Field(i).Name
		// fType := f.Type()
		fVal := f.Interface()
		if f.Type() == reflect.TypeOf(time.Time{}) {
			fVal = f.Interface().(time.Time).Format(time.RFC3339)
		}
		// log.Printf("fieldX : %s: %+v || %+v", fName, f.Type(), f.Interface())
		if fName == "ID" {
			res.ID = fVal
		}
		res.Fields = append(res.Fields, field{fName, fVal})
	}

	return res
}
