package rip

import (
	_ "embed"
	"errors"
	"io"
)

//go:embed resource.gotpl
var resourceTmpl string

func NewHTMLEncoder(w io.Writer) *HTMLEncoder {
	return &HTMLEncoder{
		w: w,
	}
}

type HTMLDecoder struct {
	r io.Reader
}

func (e HTMLDecoder) Decode(v interface{}) error {
	return errors.New("html decoder not implemented")
}

func NewHTMLDecoder(r io.Reader) *HTMLDecoder {
	return &HTMLDecoder{
		r: r,
	}
}
