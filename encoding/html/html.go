package html

import (
	_ "embed"
	"errors"
	"io"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/codecwrap"
)

// NewEntityCodec creates a HTML codec that uses pathPrefix for links creation.
func NewEntityCodec(pathPrefix string) encoding.Codec {
	// TODO: should have a better design so the path shouldn't be passed many times around.
	return codecwrap.Wrap(NewEncoder(pathPrefix), NewDecoder, MimeTypes...)
}

var MimeTypes = []string{
	"text/html",
}

//go:embed entity.gotpl
var entityTmpl string

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (e Decoder) Decode(v interface{}) error {
	return errors.New("html decoder not implemented")
}

type Encoder struct {
	w          io.Writer
	pathPrefix string
}

func NewEncoder(pathPrefix string) func(w io.Writer) *Encoder {
	return func(w io.Writer) *Encoder {
		return &Encoder{
			w:          w,
			pathPrefix: pathPrefix,
		}
	}
}

func (e Encoder) Encode(v interface{}) error {
	return htmlEncode(e.pathPrefix, e.w, editOff, v)
}
