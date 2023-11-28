package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.WrapCodec(xml.NewEncoder, xml.NewDecoder)

var MimeTypes = []string{
	"application/xml",
	"text/xml",
}
