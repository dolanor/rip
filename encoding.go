package rip

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/ajg/form"
	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v3"
)

var AvailableEncodings = []string{
	"application/json",
	"text/xml",
	"application/yaml",
	"text/html",
	"application/x-www-form-urlencoded",
	"application/msgpack",
}

var AvailableCodecs = map[string]Codec{
	"application/json":                  {NewEncoder: WrapEncoder(json.NewEncoder), NewDecoder: WrapDecoder(json.NewDecoder)},
	"text/xml":                          {NewEncoder: WrapEncoder(xml.NewEncoder), NewDecoder: WrapDecoder(xml.NewDecoder)},
	"application/yaml":                  {NewEncoder: WrapEncoder(yaml.NewEncoder), NewDecoder: WrapDecoder(yaml.NewDecoder)},
	"text/html":                         {NewEncoder: WrapEncoder(newHTMLEncoder), NewDecoder: WrapDecoder(newHTMLDecoder)},
	"application/x-www-form-urlencoded": {NewEncoder: WrapEncoder(newHTMLFormEncoder), NewDecoder: WrapDecoder(form.NewDecoder)},
	"application/msgpack":               {NewEncoder: WrapEncoder(msgpack.NewEncoder), NewDecoder: WrapDecoder(msgpack.NewDecoder)},
}

type Codec struct {
	NewEncoder NewEncoderFunc
	NewDecoder NewDecoderFunc
}

type NewDecoderFunc func(r io.Reader) Decoder

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

type NewEncoderFunc func(w io.Writer) Encoder

type Encoder interface {
	Encode(v interface{}) error
}

func WrapDecoder[D Decoder, F func(r io.Reader) D](f F) NewDecoderFunc {
	return func(r io.Reader) Decoder {
		return f(r)
	}
}

func WrapEncoder[E Encoder, F func(w io.Writer) E](f F) NewEncoderFunc {
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
		return newHTMLFormEncoder(w)
	}

	return encoder.NewEncoder(w)
}
