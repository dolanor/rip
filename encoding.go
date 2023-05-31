package rip

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"errors"
	"html/template"
	"io"
	"net/http"
	"reflect"
	"strings"
	"time"

	"github.com/ajg/form"
)

var AvailableEncodings = []string{
	"application/json",
	"text/xml",
	"text/html",
	"application/x-www-form-urlencoded",
}

var AvailableCodecs = map[string]Codec{
	"application/json":                  {NewEncoder: WrapEncoder(json.NewEncoder), NewDecoder: WrapDecoder(json.NewDecoder)},
	"text/xml":                          {NewEncoder: WrapEncoder(xml.NewEncoder), NewDecoder: WrapDecoder(xml.NewDecoder)},
	"text/html":                         {NewEncoder: WrapEncoder(NewHTMLEncoder), NewDecoder: WrapDecoder(NewHTMLDecoder)},
	"application/x-www-form-urlencoded": {NewEncoder: WrapEncoder(NewHTMLFormEncoder), NewDecoder: WrapDecoder(form.NewDecoder)},
}

type HTMLFormEncoder struct {
	w io.Writer
}

func NewHTMLFormEncoder(w io.Writer) *HTMLFormEncoder {
	return &HTMLFormEncoder{
		w: w,
	}
}

func (e HTMLFormEncoder) Encode(v interface{}) error {
	return HTMLEncode(e.w, EditOn, v)
}

type HTMLEncoder struct {
	w io.Writer
}

func (e HTMLEncoder) Encode(v interface{}) error {
	return HTMLEncode(e.w, EditOff, v)
}

func HTMLEncode(w io.Writer, edit EditMode, v interface{}) error {
	s := reflect.ValueOf(v)
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
	}

	var resources []resource
	if s.Kind() == reflect.Slice {
		for i := 0; i < s.Len(); i++ {
			s := s.Index(i)

			res := expandFields(s)

			resources = append(resources, res)
		}
	} else {
		res := expandFields(s)
		resources = append(resources, res)
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
	})

	tmplSrc := resourceTmpl
	if edit {
		tmplSrc = resourceFormTmpl
	}

	tpl, err := tpl.Parse(tmplSrc)
	if err != nil {
		return err
	}

	err = tpl.Execute(w, resources)
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

func expandFields(s reflect.Value) resource {
	t := s.Type()
	name := t.Name()
	// log.Println("encode: name:", name)
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

//go:embed resource.gotpl
var resourceTmpl string

//go:embed resource_form.gotpl
var resourceFormTmpl string

func NewHTMLEncoder(w io.Writer) *HTMLEncoder {
	return &HTMLEncoder{
		w: w,
	}
}

type HTMLDecoder struct {
	r io.Reader
}

func (e HTMLDecoder) Decode(v interface{}) error {
	return errors.New("not implemented")
}

func NewHTMLDecoder(r io.Reader) *HTMLDecoder {
	return &HTMLDecoder{
		r: r,
	}
}

type Codec struct {
	NewEncoder NewEncoder
	NewDecoder NewDecoder
}

func WrapDecoder[D Decoder, F func(r io.Reader) D](f F) func(r io.Reader) Decoder {
	return func(r io.Reader) Decoder {
		return f(r)
	}
}

func WrapEncoder[E Encoder, F func(w io.Writer) E](f F) func(w io.Writer) Encoder {
	return func(w io.Writer) Encoder {
		return f(w)
	}
}

type NewDecoder func(r io.Reader) Decoder

type Decoder interface {
	Decode(v interface{}) error
}

func contentTypeDecoder(r io.Reader, contentTypeHeader string) Decoder {
	decoder, ok := AvailableCodecs[contentTypeHeader]
	if !ok {
		return json.NewDecoder(r)
	}

	return decoder.NewDecoder(r)
}

type NewEncoder func(w io.Writer) Encoder

type Encoder interface {
	Encode(v interface{}) error
}

type EditMode bool

const (
	EditOff EditMode = false
	EditOn  EditMode = true
)

func acceptEncoder(w io.Writer, acceptHeader string, edit EditMode) Encoder {
	encoder, ok := AvailableCodecs[acceptHeader]
	if !ok {
		return json.NewEncoder(w)
	}

	if acceptHeader == "text/html" && edit {
		return NewHTMLFormEncoder(w)
	}

	return encoder.NewEncoder(w)
}
