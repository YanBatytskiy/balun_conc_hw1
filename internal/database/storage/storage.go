package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"lesson1/internal/database/dberrors"
)

type Storage struct {
	log            *slog.Logger
	commandStorage CommandStorage
	queryStorage   QueryStorage
}

type CommandStorage interface {
	Set(ctx context.Context, key, value string) error
	Del(ctx context.Context, key string) error
}

type QueryStorage interface {
	Get(ctx context.Context, key string) (string, error)
}

func NewStorage(log *slog.Logger, eng interface {
	CommandStorage
	QueryStorage
},
) *Storage {
	return &Storage{
		log:            log,
		commandStorage: eng,
		queryStorage:   eng,
	}
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	const op = "storage.Set"

	err := s.commandStorage.Set(ctx, key, value)
	if err != nil {
		s.log.Error("set failed", slog.String("key", key), slog.Any("err", err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	const op = "storage.Get"

	result, err := s.queryStorage.Get(ctx, key)
	if err != nil {
		if errors.Is(err, dberrors.ErrNotFound) {
			s.log.Info("get not found", slog.String("key", key))
		} else {
			s.log.Error("get failed", slog.String("key", key), slog.Any("err", err))
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return result, nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	const op = "storage.Del"

	err := s.commandStorage.Del(ctx, key)
	if err != nil {
		if errors.Is(err, dberrors.ErrNotFound) {
			s.log.Info("del not found", slog.String("key", key))
		} else {
			s.log.Error("del failed", slog.String("key", key), slog.Any("err", err))
		}
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}
