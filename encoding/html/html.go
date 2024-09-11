package html

import (
	"embed"
	_ "embed"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/codecwrap"
)

// htmxHandled make sure the server serves the htmx source file
var htmxHandled sync.Once

// NewEntityCodec creates a HTML codec that uses pathPrefix for links creation.
func NewEntityCodec(pathPrefix string) encoding.Codec {
	htmxHandled.Do(func() {
		http.HandleFunc("/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(htmxJS)
			if err != nil {
				log.Println("error sending htmx js script file")
				return
			}
		})
	})

	// TODO: should have a better design so the path shouldn't be passed many times around.
	return codecwrap.Wrap(NewEncoder(pathPrefix), NewDecoder, MimeTypes...)
}

var MimeTypes = []string{
	"text/html",
}

//go:embed htmx.org@*.min.js
var htmxJS []byte

//go:embed *.gotpl
var templateFiles embed.FS

const (
	entityPageTmpl     = "entity_page"
	entityListPageTmpl = "entity_list_page"

	entityTmpl = "entity"
)

type Decoder struct {
	r io.Reader
}

func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{
		r: r,
	}
}

func (e Decoder) Decode(v interface{}) error {
	return errors.New("html decoder not implemented")
}

type Encoder struct {
	w          io.Writer
	pathPrefix string
}

func NewEncoder(pathPrefix string) func(w io.Writer) *Encoder {
	return func(w io.Writer) *Encoder {
		return &Encoder{
			w:          w,
			pathPrefix: pathPrefix,
		}
	}
}

func (e Encoder) Encode(v interface{}) error {
	return htmlEncode(e.pathPrefix, e.w, editOff, v)
}
