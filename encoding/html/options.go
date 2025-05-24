package html

import (
	"io/fs"
)

type Option func(cfg *EncoderConfig)

type EncoderConfig struct {
	templatesFS fs.FS
	mux         HandleFuncer
}

func WithTemplatesFS(templatesFS fs.FS) Option {
	return func(cfg *EncoderConfig) {
		cfg.templatesFS = templatesFS
	}
}

func WithServeMux(mux HandleFuncer) Option {
	return func(cfg *EncoderConfig) {
		cfg.mux = mux
	}
}
