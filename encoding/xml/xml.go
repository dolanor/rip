package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding"
)

func init() {

	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.Codec{
	NewEncoder: encoding.WrapEncoder(xml.NewEncoder),
	NewDecoder: encoding.WrapDecoder(xml.NewDecoder),
}

var MimeTypes = []string{
	"application/xml",
	"text/xml",
}
