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
)

var AvailableEncodings = []string{
	"text/json",
	"text/xml",
	"text/html",
}

var AvailableCodecs = map[string]Codec{
	"text/json": {NewEncoder: WrapEncoder(json.NewEncoder), NewDecoder: WrapDecoder(json.NewDecoder)},
	"text/xml":  {NewEncoder: WrapEncoder(xml.NewEncoder), NewDecoder: WrapDecoder(xml.NewDecoder)},
	"text/html": {NewEncoder: WrapEncoder(NewHTMLEncoder), NewDecoder: WrapDecoder(NewHTMLDecoder)},
}

type HTMLFormEncoder struct {
	w io.Writer
}

func (e HTMLFormEncoder) Encode(v interface{}) error {
	return HTMLEncode(e.w, true, v)
}

type HTMLEncoder struct {
	w io.Writer
}

//go:embed resource.gotpl
var resourceTmpl string

//go:embed resource_form.gotpl
var resourceFormTmpl string

func (e HTMLEncoder) Encode(v interface{}) error {
	return HTMLEncode(e.w, false, v)
}

func HTMLEncode(w io.Writer, edit bool, v interface{}) error {
	s := reflect.ValueOf(v)
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
	}
	t := s.Type()
	name := t.Name()
	//log.Println("encode: name:", name)

	rw, ok := w.(http.ResponseWriter)
	if ok {
		// We force Content-Type to HTML
		rw.Header().Set("Content-Type", "text/html")
	}

	type field struct {
		Key   string
		Value any
	}

	type resource struct {
		ID     string
		Name   string
		Fields []field
	}

	res := resource{
		Name: name,
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fName := t.Field(i).Name
		//fType := f.Type()
		fVal := f.Interface()
		//log.Printf("fieldX : %s: %+v || %+v", fName, f.Type(), f.Interface())
		res.Fields = append(res.Fields, field{fName, fVal})
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

	err = tpl.Execute(w, res)
	if err != nil {
		return err
	}

	return nil
}

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

func acceptEncoder(w io.Writer, acceptHeader string) Encoder {
	encoder, ok := AvailableCodecs[acceptHeader]
	if !ok {
		return json.NewEncoder(w)
	}

	return encoder.NewEncoder(w)
}
