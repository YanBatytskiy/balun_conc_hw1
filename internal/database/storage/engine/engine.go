package engine

import (
	"context"
	"fmt"
	"log/slog"

	hashtable "lesson1/internal/database/hash_table"
)

type Engine struct {
	log           *slog.Logger
	commandEngine CommandEngine
	queryEngine   QueryEngine
}

type CommandEngine struct {
	hashTable *hashtable.HashTable
}
type QueryEngine struct {
	hashTable *hashtable.HashTable
}

func NewEngine(log *slog.Logger) *Engine {
	hashTable := hashtable.NewHashTable()
	return &Engine{
		log:           log,
		commandEngine: CommandEngine{hashTable: hashTable},
		queryEngine:   QueryEngine{hashTable: hashTable},
	}
}

func (e *Engine) Set(ctx context.Context, key, value string) error {
	const op = "engine.Set"
	_ = ctx

	err := e.commandEngine.hashTable.Set(key, value)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (e *Engine) Get(ctx context.Context, key string) (string, error) {
	const op = "engine.Get"
	_ = ctx

	result, err := e.queryEngine.hashTable.Get(key)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (e *Engine) Del(ctx context.Context, key string) error {
	const op = "engine.Del"
	_ = ctx

	err := e.commandEngine.hashTable.Del(key)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
