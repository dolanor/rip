package encoding

import (
	_ "embed"
	"encoding/json"
	"errors"
	"io"
	"maps"
	"net/http"
)

var ErrNoEncoderAvailable = errors.New("codec not available")

type EditMode bool

const (
	EditOff EditMode = false
	EditOn  EditMode = true
)

var availableCodecs = map[string]Codec{}

func AvailableCodecs() map[string]Codec {
	codecs := maps.Clone(availableCodecs)
	return codecs
}

func RegisterCodec(codec Codec, mimes ...string) {
	for _, mime := range mimes {
		availableCodecs[mime] = codec
	}
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

func WrapCodec[E Encoder, EFunc func(w io.Writer) E, D Decoder, DFunc func(r io.Reader) D](encoderFunc EFunc, decoderFunc DFunc) Codec {
	return Codec{
		NewEncoder: func(w io.Writer) Encoder { return encoderFunc(w) },
		NewDecoder: func(r io.Reader) Decoder { return decoderFunc(r) },
	}
}

func AcceptEncoder(w http.ResponseWriter, acceptHeader string, edit EditMode) Encoder {
	// TODO: add some hook to be able to tune this from the codec package
	if acceptHeader == "text/html" && edit {
		formCodec, ok := availableCodecs["application/x-www-form-urlencoded"]
		if !ok {
			return &noEncoder{}
		}
		return formCodec.NewEncoder(w)
	}

	encoder, ok := availableCodecs[acceptHeader]
	if !ok {

		encoder, ok := availableCodecs["default"]
		if !ok {
			return &noEncoder{}
		}
		return encoder.NewEncoder(w)
	}

	return encoder.NewEncoder(w)
}

type noEncoder struct{}

func (e *noEncoder) Encode(v interface{}) error {
	return ErrNoEncoderAvailable
}
