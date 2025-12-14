package initialization

import (
	"context"
	"fmt"
	"log/slog"
	"spyder/internal/config"
	"spyder/internal/database"
	"spyder/internal/database/compute"
	"spyder/internal/database/storage"
	"spyder/internal/network"

	inmemory "spyder/internal/database/storage/engine/in_memory"
)

type Initializer struct {
	Log    *slog.Logger
	engine *inmemory.Engine
	server *network.TCPServer
}

func NewInitilizer(cfg *config.Config) (*Initializer, error) {
	const op = "initialization.NewInitilizer"

	if cfg == nil {
		return nil, fmt.Errorf("%s: failed to initilize: config is invaled", op)
	}

	log, err := CreateLogger(cfg)
	if err != nil {
		return nil, err
	}

	engine, err := CreateEngine(log)
	if err != nil {
		return nil, err
	}

	network, err := CreateNetwork(log, cfg.Network)
	if err != nil {
		return nil, err
	}

	return &Initializer{
		Log:    log,
		engine: engine,
		server: network,
	}, nil
}

func (init *Initializer) StartDatabase(ctx context.Context) error {
	const op = "initialization.SartDatabase"

	_ = ctx

	compute, err := compute.NewCompute(init.Log)
	if err != nil {
		init.Log.Info("failed to create compute layer", slog.String("error", err.Error()))
		return fmt.Errorf("%s: failed to create compute layer", op)
	}

	storage, err := storage.NewStorage(init.Log, init.engine)
	if err != nil {
		init.Log.Info("failed to create storage layer", slog.String("error", err.Error()))
		return fmt.Errorf("%s: failed to create storage layer", op)
	}

	database, err := database.NewDatabase(init.Log, compute, storage)
	if err != nil {
		init.Log.Info("failed to create database", slog.String("error", err.Error()))
		return fmt.Errorf("%s: failed to create database", op)
	}

	err = init.server.HandleClientQueries(ctx, database)
	if err != nil {
		init.Log.Debug("failed to handle clients's query %w", slog.String("err", err.Error()))
	}

	return err
}
