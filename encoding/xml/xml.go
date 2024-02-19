package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding/codecwrap"
)

var Codec = codecwrap.Wrap(xml.NewEncoder, xml.NewDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/xml",
	"text/xml",
}
