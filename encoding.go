package rip

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
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

type HTMLEncoder struct {
	w io.Writer
}

func (e HTMLEncoder) Encode(v interface{}) error {
	s := reflect.ValueOf(v)
	if s.Kind() == reflect.Pointer {
		s = s.Elem()
	}
	t := s.Type()
	name := t.Name()

	rw, ok := e.w.(http.ResponseWriter)
	if ok {
		// We force Content-Type to HTML
		rw.Header().Set("Content-Type", "text/html")
	}

	// TODO: replace with template
	_, err := e.w.Write([]byte(fmt.Sprintf(`<h2>%s</h2>`, name)))
	if err != nil {
		return err
	}
	_, err = e.w.Write([]byte(`<div hx-target="this" hx-swap="outerHTML">`))
	if err != nil {
		return err
	}
	for i := 0; i < s.NumField(); i++ {
		f := s.Field(i)
		fName := t.Field(i).Name
		//fType := f.Type()
		fVal := f.Interface()

		_, err = e.w.Write([]byte(fmt.Sprintf(`%s<div><label>%s</label>: %v</div>`, "\n", fName, fVal)))
		if err != nil {
			return err
		}

	}

	// TODO: replace with template
	_, err = e.w.Write([]byte(`<button hx-get="/users/2/edit" class="btn btn-primary">Click to edit</button></div>`))
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
