package encoding

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"io"

	"github.com/vmihailenco/msgpack/v5"
	"gopkg.in/yaml.v3"
)

type EditMode bool

const (
	EditOff EditMode = false
	EditOn  EditMode = true
)

var AvailableEncodings = []string{
	"application/json",
	"text/xml",
	"application/yaml",
	"text/html",
	"application/x-www-form-urlencoded",
	"application/msgpack",
}

var availableCodecs = map[string]Codec{
	"application/json":    {NewEncoder: WrapEncoder(json.NewEncoder), NewDecoder: WrapDecoder(json.NewDecoder)},
	"text/xml":            {NewEncoder: WrapEncoder(xml.NewEncoder), NewDecoder: WrapDecoder(xml.NewDecoder)},
	"application/yaml":    {NewEncoder: WrapEncoder(yaml.NewEncoder), NewDecoder: WrapDecoder(yaml.NewDecoder)},
	"application/msgpack": {NewEncoder: WrapEncoder(msgpack.NewEncoder), NewDecoder: WrapDecoder(msgpack.NewDecoder)},
}

func RegisterCodec(mime string, codec Codec) {
	availableCodecs[mime] = codec
}

type Codec struct {
	NewEncoder NewEncoderFunc
	NewDecoder NewDecoderFunc
}

type NewDecoderFunc func(r io.Reader) Decoder

type Decoder interface {
	Decode(v interface{}) error
}

func ContentTypeDecoder(r io.Reader, contentTypeHeader string) Decoder {
	decoder, ok := availableCodecs[contentTypeHeader]
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

func AcceptEncoder(w io.Writer, acceptHeader string, edit EditMode) Encoder {
	encoder, ok := availableCodecs[acceptHeader]
	if !ok {
		return json.NewEncoder(w)
	}

	if acceptHeader == "text/html" && edit {
		return availableCodecs["application/x-www-form-urlencoded"].NewEncoder(w)
	}

	return encoder.NewEncoder(w)
}
