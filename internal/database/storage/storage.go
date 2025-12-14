package storage

import (
	"context"
	"log/slog"
	"strings"
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
) (*Storage, error) {
	if log == nil {
		return nil, ErrInvalidLogger
	}

	return &Storage{
		log:            log,
		commandStorage: eng,
		queryStorage:   eng,
	}, nil
}

func (s *Storage) Set(ctx context.Context, key, value string) error {
	err := s.commandStorage.Set(ctx, key, value)
	if err != nil {
		s.log.Error("set failed", slog.String("key", key), slog.Any("err", err))
		return err
	}
	return nil
}

func (s *Storage) Get(ctx context.Context, key string) (string, error) {
	result, err := s.queryStorage.Get(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.log.Debug("get not found", slog.String("key", key))
		} else {
			s.log.Error("get failed", slog.String("key", key), slog.Any("err", err))
		}
		return "", err
	}
	return result, nil
}

func (s *Storage) Del(ctx context.Context, key string) error {
	err := s.commandStorage.Del(ctx, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			s.log.Debug("del not found", slog.String("key", key))
		} else {
			s.log.Error("del failed", slog.String("key", key), slog.Any("err", err))
		}
		return err
	}
	return nil
}
