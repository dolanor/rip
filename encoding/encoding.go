package encoding

import (
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// ErrNoEncoderAvailable communicates that a codec is not available.
var ErrNoEncoderAvailable = errors.New("codec not available")

// EditMode allows to select an editable version of a codec.
// eg. HTML Forms for HTML codec
type EditMode bool

const (
	// EditOff indicates that the data is requested as read-only data.
	EditOff EditMode = false

	// EditOn indicates that the data is requested as editable data.
	EditOn EditMode = true
)

// Codecs is a registry of codecs usually related to a route option.
type Codecs struct {
	Codecs           map[string]Codec
	OrderedMimeTypes []string
}

const defaultCodecKey = "default_codec_key"

// Register registers a new codec to the codec registry.
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

// Codec combines an encoder, a decoder and a list of mime types.
type Codec struct {
	NewEncoder func(w io.Writer) Encoder
	NewDecoder func(r io.Reader) Decoder
	MimeTypes  []string
}

// Decoder decodes encoded value from a input stream.
type Decoder interface {
	// Decode decodes encoded value from the input stream into v.
	Decode(v interface{}) error
}

// ContentTypeDecoder decodes the encoded data from r based on the Content-Type header value
// and the codecs that are available.
// If no codec is found, it falls back to JSON.
// FIXME: use another fallback, maybe an error is better
func ContentTypeDecoder(r io.Reader, contentTypeHeader string, codecs Codecs) Decoder {
	decoder, ok := codecs.Codecs[contentTypeHeader]
	if !ok {
		return json.NewDecoder(r)
	}

	return decoder.NewDecoder(r)
}

// Encoder writes encoded value to an output stream.
type Encoder interface {
	// Encode writes the codec data of v to the output stream.
	Encode(v interface{}) error
}

// AcceptEncoder creates an new encoder for w based on the acceptHeader, the edit mode and
// the codecs that are available.
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
