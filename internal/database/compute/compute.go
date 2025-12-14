package compute

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

//go:generate go run github.com/vektra/mockery/v3@v3.6.1 --config ../../.mockery.yaml

type Compute struct {
	log *slog.Logger
}

func NewCompute(log *slog.Logger) (*Compute, error) {
	const op = "compute.NewCompute"

	if log == nil {
		return nil, fmt.Errorf("%s: logger is invaled", op)
	}

	return &Compute{log: log}, nil
}

func (c *Compute) ParseAndValidate(_ context.Context, raw string) ([]string, error) {
	const op = "compute.parse"

	tokens := strings.Fields(strings.TrimSpace(raw))

	if len(tokens) == 0 {
		c.log.Debug("empty command", slog.String("operation", op))
		return nil, ErrEmptyCommand
	}

	ok := ValidateCommand(tokens[0])

	if !ok {
		c.log.Debug("invalid syntax of command", slog.String("operation", op), slog.String("token", tokens[0]))
		return nil, ErrInvalidCommand
	}

	tokens[0] = strings.ToUpper(tokens[0])

	for _, token := range tokens {
		ok := ValidateArgument(token)
		if !ok {
			c.log.Debug("invalid syntax of argument", slog.String("operation", op), slog.String("token", token))
			return nil, ErrInvalidArgument
		}
	}

	return tokens, nil
}
