package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding"
)

func init() {
	codec := encoding.Codec{
		NewEncoder: encoding.WrapEncoder(xml.NewEncoder),
		NewDecoder: encoding.WrapDecoder(xml.NewDecoder),
	}

	encoding.RegisterCodec(codec, "application/xml", "text/xml")
}
