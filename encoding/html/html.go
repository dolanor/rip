package html

import (
	_ "embed"
	"errors"
	"io"

	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(NewEncoder, NewDecoder)

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
	w io.Writer
}

func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w: w,
	}
}

func (e Encoder) Encode(v interface{}) error {
	return htmlEncode(e.w, editOff, v)
}
