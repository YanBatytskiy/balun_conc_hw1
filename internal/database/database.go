package database

import (
	"context"
	"fmt"
	"log/slog"
	"spyder/internal/database/compute"
	"strings"
)

type ComputeLayer interface {
	ParseAndValidate(_ context.Context, raw string) ([]string, error)
}

type StorageLayer interface {
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
	Get(ctx context.Context, key string) (string, error)
}

type Database struct {
	Compute ComputeLayer
	Storage StorageLayer
	log     *slog.Logger
}

func NewDatabase(
	log *slog.Logger,
	computeLayer ComputeLayer,
	storageLayer StorageLayer,
) (*Database, error) {
	const op = "database.NewDatabase"

	if computeLayer == nil {
		return nil, fmt.Errorf("%s: compute is invaled", op)
	}
	if storageLayer == nil {
		return nil, fmt.Errorf("%s: storage is invaled", op)
	}
	if log == nil {
		return nil, fmt.Errorf("%s: logger is invaled", op)
	}
	return &Database{
		Compute: computeLayer,
		Storage: storageLayer,
		log:     log,
	}, nil
}

func (db *Database) DatabaseHandler(ctx context.Context, raw string) (string, error) {
	const op = "database.handler"

	tokens, err := db.Compute.ParseAndValidate(ctx, raw)
	if err != nil {
		db.log.Debug("failed to parse command", slog.String("operation", op), slog.String("error", err.Error()))
		return "", err
	}

	db.log.Debug("command start", slog.String("cmd", tokens[0]))

	switch tokens[0] {
	case compute.CommandSet:
		return db.handleSet(ctx, tokens)
	case compute.CommandGet:
		return db.handleGet(ctx, tokens)
	case compute.CommandDel:
		return db.handleDel(ctx, tokens)
	default:
		db.log.Debug("invalid command")

		return "", fmt.Errorf("%s: invalid command", op)
	}
}

func (db *Database) handleSet(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.set"

	if len(tokens)-1 != compute.CommandSetQ {
		db.log.Info("must be two arguments")
		return "", fmt.Errorf("%s: invalid quantity of arguments", op)
	}

	err := db.Storage.Set(ctx, tokens[1], tokens[2])
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	db.log.Info("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "OK", nil
}

// our mocked QueryService method.
func (db *Database) handleGet(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.get"

	if len(tokens)-1 != compute.CommandGetQ {
		db.log.Debug("must be one arguments")
		return "", fmt.Errorf("%s: invalid quantity of arguments", op)
	}

	result, err := db.Storage.Get(ctx, tokens[1])
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return "NOT_FOUND", nil
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	db.log.Debug("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "VALUE " + result, nil
}

func (db *Database) handleDel(ctx context.Context, tokens []string) (string, error) {
	const op = "compute.del"

	if len(tokens)-1 != compute.CommandDelQ {
		db.log.Debug("must be one arguments")
		return "", fmt.Errorf("%s: invalid quantity of arguments", op)
	}

	err := db.Storage.Del(ctx, tokens[1])
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return "NOT_FOUND", nil
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}

	db.log.Debug("command success", slog.String("cmd", tokens[0]), slog.String("key", tokens[1]))
	return "DELETED", nil
}
