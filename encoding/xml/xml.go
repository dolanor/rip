package xml

import (
	"encoding/xml"

	"github.com/dolanor/rip/encoding"
)

func init() {

	encoding.RegisterCodec("text/xml", encoding.Codec{NewEncoder: encoding.WrapEncoder(xml.NewEncoder), NewDecoder: encoding.WrapDecoder(xml.NewDecoder)})
}
