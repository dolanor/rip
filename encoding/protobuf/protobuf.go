package protobuf

import (
	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, "application/vnd.google.protobuf", "application/protobuf", "application/x-protobuf")
}

var Codec = encoding.Codec{
	NewEncoder: encoding.WrapEncoder(newEncoder),
	NewDecoder: encoding.WrapDecoder(newDecoder),
}