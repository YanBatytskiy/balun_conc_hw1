package compute

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"lesson1/internal/command"
	"lesson1/internal/database/dberrors"
)

var (
	ErrEmptyCommand         = errors.New("empty command")
	ErrInvalidCommand       = errors.New("invalid command")
	ErrInvalidArg           = errors.New("invalid argument")
	ErrInvalidQuantity      = errors.New("invalid quantity of arguments")
	ErrInvalidSyntaxCommand = errors.New("invalid syntax of command")
	ErrInvalidSyntaxArg     = errors.New("invalid syntax of argument")
)

//go:generate go run github.com/vektra/mockery/v3@v3.6.1 --config ../../.mockery.yaml

type CommandCompute interface {
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
}

type QueryCompute interface {
	Get(ctx context.Context, key string) (string, error)
}

type Compute struct {
	log            *slog.Logger
	commandCompute CommandCompute
	queryCompute   QueryCompute
}

func NewCompute(log *slog.Logger, cmd interface {
	CommandCompute
	QueryCompute
},
) *Compute {
	return &Compute{
		log:            log,
		commandCompute: cmd,
		queryCompute:   cmd,
	}
}

func (c *Compute) ComputeHandler(ctx context.Context, raw string) (string, error) {
	const op = "compute.ComputeHandler"

	tokens, err := c.ParseAndValidate(ctx, raw)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	c.log.Info("command start", slog.String("cmd", tokens[0]))

	switch tokens[0] {
	case command.CommandSet:
		return c.handleSet(ctx, tokens)
	case command.CommandGet:
		return c.handleGet(ctx, tokens)
	case command.CommandDel:
		return c.handleDel(ctx, tokens)
	default:
		c.log.Info("invalid command")

		return "", fmt.Errorf("%s: %w", op, ErrInvalidCommand)
	}
}

func (c *Compute) ParseAndValidate(_ context.Context, raw string) ([]string, error) {
	const op = "compute.parse"

	tokens := strings.Fields(strings.TrimSpace(raw))

	if len(tokens) == 0 {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyCommand)
	}

	ok := ValidateCommand(tokens[0])

	if !ok {
		c.log.Info("invalid syntax of command", slog.String(" ", tokens[0]))
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidSyntaxCommand)
	}

	tokens[0] = strings.ToUpper(tokens[0])

	for _, token := range tokens {
		ok := ValidateArgument(token)
		if !ok {
			c.log.Info("invalid syntax of argument", slog.String(" ", token))
			return nil, fmt.Errorf("%s: %w", op, ErrInvalidSyntaxArg)
		}
	}

	return tokens, nil
}

func (c *Compute) handleSet(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.set"

	if len(tokens)-1 != command.CommandSetQ {
		c.log.Info("must be two arguments")
		return "", ErrInvalidQuantity
	}

	err := c.commandCompute.Set(ctx, tokens[1], tokens[2])
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	c.log.Info("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "OK", nil
}

// our mocked QueryService method
func (c *Compute) handleGet(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.get"

	if len(tokens)-1 != command.CommandGetQ {
		c.log.Info("must be one arguments")
		return "", fmt.Errorf("%w", ErrInvalidQuantity)
	}

	result, err := c.queryCompute.Get(ctx, tokens[1])
	if err != nil {
		if errors.Is(err, dberrors.ErrNotFound) {
			return "NOT_FOUND", nil
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	c.log.Info("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "VALUE " + result, nil
}

func (c *Compute) handleDel(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.del"

	if len(tokens)-1 != command.CommandDelQ {
		c.log.Info("must be one arguments")
		return "", fmt.Errorf("%w", ErrInvalidQuantity)
	}

	err := c.commandCompute.Del(ctx, tokens[1])
	if err != nil {
		if errors.Is(err, dberrors.ErrNotFound) {
			return "NOT_FOUND", nil
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	c.log.Info("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "DELETED", nil
}
