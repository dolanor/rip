package rip

import (
	_ "embed"
	"errors"
	"io"
)

//go:embed resource.gotpl
var resourceTmpl string

func newHTMLEncoder(w io.Writer) *htmlEncoder {
	return &htmlEncoder{
		w: w,
	}
}

type htmlDecoder struct {
	r io.Reader
}

func (e htmlDecoder) Decode(v interface{}) error {
	return errors.New("html decoder not implemented")
}

func newHTMLDecoder(r io.Reader) *htmlDecoder {
	return &htmlDecoder{
		r: r,
	}
}
