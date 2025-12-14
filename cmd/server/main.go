package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"spyder/internal/config"
	"spyder/internal/initialization"
	"syscall"
	"time"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.NewConfig()
	if err != nil {
		log.Printf("failed to load config: %v", err)
		stop()
		os.Exit(1)
	}

	init, err := initialization.NewInitilizer(cfg)
	if err != nil {
		log.Printf("failed to initialize: %v", err)
		stop()
		os.Exit(1)
	}

	go func() {
		err = init.StartDatabase(ctx)
		if err != nil {
			init.Log.Error("failed to start database", slog.String("error", err.Error()))
		}
	}()

	waitForShutdown(ctx, init.Log)
	init.Log.Info("service stoped")
}

func waitForShutdown(
	ctx context.Context,
	log *slog.Logger,
) {
	<-ctx.Done()
	log.Info("received shutdown signal ")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	<-shutdownCtx.Done()
	log.Info("shutdown complete")
}
