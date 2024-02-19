package protobuf

import "github.com/dolanor/rip/encoding/codecwrap"

var Codec = codecwrap.Wrap(newEncoder, newDecoder, MimeTypes...)

var MimeTypes = []string{
	"application/vnd.google.protobuf",
	"application/protobuf",
	"application/x-protobuf",
}
