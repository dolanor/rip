package domain

import "errors"

var (
	ErrAppNotFound       = errors.New("not found")
	ErrAppNotImplemented = errors.New("not implemented")
)
