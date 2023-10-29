package html

import (
	_ "embed"
	"errors"
	"io"

	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(NewEncoder),
		NewDecoder: encoding.WrapDecoder(NewDecoder),
	}

	encoding.RegisterCodec(codec, "text/html")
}

//go:embed resource.gotpl
var resourceTmpl string

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
