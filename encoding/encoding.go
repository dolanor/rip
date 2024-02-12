package encoding

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
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

const defaultCodecKey = "default_codec_key"

func (c *Codecs) Register(codec Codec) {
	_, ok := c.Codecs[defaultCodecKey]
	if !ok {
		c.Codecs[defaultCodecKey] = codec
	}

	for _, mime := range codec.MimeTypes {
		c.Codecs[mime] = codec
		c.OrderedMimeTypes = append(c.OrderedMimeTypes, mime)
	}
}


type Codec struct {
	NewEncoder NewEncoderFunc
	NewDecoder NewDecoderFunc
	MimeTypes  []string
}

type NewDecoderFunc = func(r io.Reader) Decoder

type Decoder interface {
	Decode(v interface{}) error
}

func ContentTypeDecoder(r io.Reader, contentTypeHeader string, codecs Codecs) Decoder {
	decoder, ok := codecs.Codecs[contentTypeHeader]
	if !ok {
		return json.NewDecoder(r)
	}

	return decoder.NewDecoder(r)
}

type NewEncoderFunc = func(w io.Writer) Encoder

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

func WrapCodec[E Encoder, EFunc func(w io.Writer) E, D Decoder, DFunc func(r io.Reader) D](encoderFunc EFunc, decoderFunc DFunc, mimeTypes ...string) Codec {
	return Codec{
		NewEncoder: func(w io.Writer) Encoder { return encoderFunc(w) },
		NewDecoder: func(r io.Reader) Decoder { return decoderFunc(r) },
		MimeTypes:  mimeTypes,
	}
}

func AcceptEncoder(w http.ResponseWriter, acceptHeader string, edit EditMode, codecs Codecs) Encoder {
	// TODO: add some hook to be able to tune this from the codec package
	if acceptHeader == "text/html" && edit {
		formCodec, ok := codecs.Codecs["application/x-www-form-urlencoded"]
		if !ok {
			return &noEncoder{missingEncoder: "application/x-www-form-urlencoded"}
		}
		return formCodec.NewEncoder(w)
	}

	encoder, ok := codecs.Codecs[acceptHeader]
	if !ok {
		encoder, ok := codecs.Codecs[defaultCodecKey]
		if !ok {
			return &noEncoder{missingEncoder: "default"}
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
