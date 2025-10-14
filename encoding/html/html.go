package html

import (
	_ "embed"
	"errors"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/dolanor/rip/encoding"
	"github.com/dolanor/rip/encoding/codecwrap"
)

// HandleFuncer is an interface that matches http.ServeMux's HandleFunc method
type HandleFuncer interface {
	HandleFunc(pattern string, handler func(http.ResponseWriter, *http.Request))
}

// htmxHandled make sure the server serves the htmx source file
var htmxHandled sync.Once

func serveHTMX(mux HandleFuncer) {
	htmxHandled.Do(func() {
		mux.HandleFunc("/js/htmx.min.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/javascript")
			w.Header().Set("Cache-Control", "max-age=3600")
			_, err := w.Write(htmxJS)
			if err != nil {
				log.Println("error sending htmx js script file")
				return
			}
		})

		mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write(favicon)
			if err != nil {
				log.Println("error sending favicon.ico")
				return
			}
		})
	})
}

// NewEntityCodec creates a HTML codec that uses pathPrefix for links creation.
func NewEntityCodec(pathPrefix string, opts ...Option) encoding.Codec {

	// TODO: should have a better design so the path shouldn't be passed many times around.
	return codecwrap.Wrap(NewEncoder(pathPrefix, opts...), NewDecoder, MimeTypes...)
}

var MimeTypes = []string{
	"text/html",
}

//go:embed htmx.org@*.min.js
var htmxJS []byte

//go:embed favicon.ico
var favicon []byte

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
	config     EncoderConfig
}

func NewEncoder(pathPrefix string, opts ...Option) func(w io.Writer) *Encoder {
	cfg := EncoderConfig{}
	for _, o := range opts {
		o(&cfg)
	}

	if cfg.mux == nil {
		cfg.mux = http.DefaultServeMux
		serveHTMX(cfg.mux)
	}

	return func(w io.Writer) *Encoder {
		return &Encoder{
			w:          w,
			pathPrefix: pathPrefix,
			config:     cfg,
		}
	}
}

func (e Encoder) Encode(v interface{}) error {
	return htmlEncode(e.pathPrefix, e.config.templatesFS, e.w, editOff, v)
}
