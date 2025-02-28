package html

import (
	"io/fs"

	"github.com/dolanor/rip"
)

type Option func(cfg *EncoderConfig)

type EncoderConfig struct {
	templatesFS fs.FS
}

func WithTemplatesFS(templatesFS fs.FS) Option {
	return func(cfg *EncoderConfig) {
		cfg.templatesFS = templatesFS
	}
}

func WithServeMux(mux rip.HTTPServeMux) Option {
	return func(cfg *EncoderConfig) {
		// serve HTMX from mux
		serveHTMX(mux)
	}
}
