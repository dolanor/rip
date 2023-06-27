package rip

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"io"

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

type Codec struct {
	NewEncoder NewEncoder
	NewDecoder NewDecoder
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
