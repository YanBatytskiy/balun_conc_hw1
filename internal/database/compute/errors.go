package compute

import "errors"

var (
	ErrInvalidCommand  = errors.New("invalid syntax of command")
	ErrInvalidArgument = errors.New("invalid syntax of argument")
	ErrEmptyCommand    = errors.New("empty command")
)
