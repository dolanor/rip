package encoding

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"slices"
)

var ErrNoEncoderAvailable = errors.New("codec not available")

type EditMode bool

const (
	EditOff EditMode = false
	EditOn  EditMode = true
)

type Codecs struct {
	Codecs           map[string]Codec
	OrderedMimeTypes []string
}

func (c *Codecs) Register(codec Codec, mimes ...string) {
	for _, mime := range mimes {
		c.Codecs[mime] = codec
		c.OrderedMimeTypes = append(c.OrderedMimeTypes, mime)
	}
}

var availableCodecs = Codecs{
	Codecs:           map[string]Codec{},
	OrderedMimeTypes: []string{},
}

func AvailableCodecs() Codecs {
	codecs := maps.Clone(availableCodecs.Codecs)
	mimeTypes := slices.Clone(availableCodecs.OrderedMimeTypes)
	return Codecs{
		Codecs:           codecs,
		OrderedMimeTypes: mimeTypes,
	}
}

func RegisterCodec(codec Codec, mimes ...string) {
	availableCodecs.Register(codec, mimes...)
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
	decoder, ok := availableCodecs.Codecs[contentTypeHeader]
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
		formCodec, ok := availableCodecs.Codecs["application/x-www-form-urlencoded"]
		if !ok {
			return &noEncoder{}
		}
		return formCodec.NewEncoder(w)
	}

	encoder, ok := availableCodecs.Codecs[acceptHeader]
	if !ok {

		encoder, ok := availableCodecs.Codecs["default"]
		if !ok {
			return &noEncoder{}
		}
		return encoder.NewEncoder(w)
	}

	return encoder.NewEncoder(w)
}

type noEncoder struct {
	missingEncoder string
}

func (e *noEncoder) Encode(v interface{}) error {
	return fmt.Errorf("%q: %w", e.missingEncoder, ErrNoEncoderAvailable)
}
