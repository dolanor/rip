package protobuf

import (
	"github.com/dolanor/rip/encoding"
)

var Codec = encoding.WrapCodec(newEncoder, newDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/vnd.google.protobuf",
	"application/protobuf",
	"application/x-protobuf",
}
