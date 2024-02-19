package codecwrap

import (
	"io"

	"github.com/dolanor/rip/encoding"
)

// Wrap is a helper that allows to easily wrap codecs from standard library into a rip codec.
func Wrap[E encoding.Encoder, EFunc func(w io.Writer) E, D encoding.Decoder, DFunc func(r io.Reader) D](encoderFunc EFunc, decoderFunc DFunc, mimeTypes ...string) encoding.Codec {
	return encoding.Codec{
		NewEncoder: func(w io.Writer) encoding.Encoder { return encoderFunc(w) },
		NewDecoder: func(r io.Reader) encoding.Decoder { return decoderFunc(r) },
		MimeTypes:  mimeTypes,
	}
}
