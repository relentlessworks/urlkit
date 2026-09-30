package model

import "errors"

var (
	ErrEmptyURL = errors.New("empty URL")
	ErrNoScheme = errors.New("URL must have a scheme (e.g. https://)")
)
