package inmemory

import (
	"context"
	"log/slog"
)

type Engine struct {
	log           *slog.Logger
	commandEngine CommandEngine
	queryEngine   QueryEngine
}

type CommandEngine struct {
	hashTable *HashTable
}
type QueryEngine struct {
	hashTable *HashTable
}

func NewEngine(log *slog.Logger) (*Engine, error) {
	if log == nil {
		return nil, ErrInvalidLogger
	}

	hashTable := NewHashTable()
	return &Engine{
		log:           log,
		commandEngine: CommandEngine{hashTable: hashTable},
		queryEngine:   QueryEngine{hashTable: hashTable},
	}, nil
}

func (e *Engine) Set(ctx context.Context, key, value string) error {
	_ = ctx

	err := e.commandEngine.hashTable.Set(key, value)
	if err != nil {
		return err
	}

	return nil
}

func (e *Engine) Get(ctx context.Context, key string) (string, error) {
	_ = ctx

	result, err := e.queryEngine.hashTable.Get(key)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (e *Engine) Del(ctx context.Context, key string) error {
	_ = ctx

	err := e.commandEngine.hashTable.Del(key)
	if err != nil {
		return err
	}

	return nil
}
