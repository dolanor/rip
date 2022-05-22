package rip

import (
	"encoding/json"
	"encoding/xml"
	"io"
)

var AvailableEncodings = []string{
	"text/json",
	"text/xml",
}

var AvailableCodecs = map[string]Codec{
	"text/json": {NewEncoder: WrapEncoder(json.NewEncoder), NewDecoder: WrapDecoder(json.NewDecoder)},
	"text/xml":  {NewEncoder: WrapEncoder(xml.NewEncoder), NewDecoder: WrapDecoder(xml.NewDecoder)},
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
