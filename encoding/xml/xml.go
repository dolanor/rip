package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(xml.NewEncoder, xml.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/xml",
	"text/xml",
}
