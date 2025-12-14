package network

import "errors"

var (
	ErrInvalidLogger     = errors.New("invalid logger")
	ErrInvalidConfig     = errors.New("invalid config")
	ErrInvalidMaxConn    = errors.New("max connections is equal zero")
	ErrInvalidBufferSize = errors.New("buffer size is equal zero")
)
