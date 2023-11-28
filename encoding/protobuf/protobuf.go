package protobuf

import (
	"github.com/dolanor/rip/encoding"
)

func init() {
	encoding.RegisterCodec(Codec, MimeTypes...)
}

var Codec = encoding.WrapCodec(newEncoder, newDecoder)

var MimeTypes = []string{
	"application/vnd.google.protobuf",
	"application/protobuf",
	"application/x-protobuf",
}
