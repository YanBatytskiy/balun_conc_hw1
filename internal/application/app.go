package application

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"lesson1/internal/cli"
	"lesson1/internal/compute"
	"lesson1/internal/config"
	"lesson1/internal/database/storage"
	"lesson1/internal/database/storage/engine"
	"lesson1/internal/lib/logger/slogdiscard"
	"lesson1/internal/lib/logger/slogpretty"
)

type App struct{}

func New() *App {
	return &App{}
}

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func (a *App) Run() {
	rootCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.MustLoad()
	log := setupLogger(cfg.Env)

	log.Info("starting service", slog.String("env", cfg.Env))
	log.Debug("debug message are enabled")

	engine := engine.NewEngine(log)
	storage := storage.NewStorage(log, engine)

	_ = engine
	_ = storage

	compute := compute.NewCompute(log, storage)

	cli := cli.NewCli(log, compute)
	cliCtx, cliErr := cli.Start(rootCtx)

	// _ = compute

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	waitForShutdown(cliCtx, cancel, log, cliErr, stop)
	log.Info("service stoped")
}

func waitForShutdown(
	cliCtx context.Context,
	cancel context.CancelFunc,
	log *slog.Logger,
	cliErr <-chan error,
	stop <-chan os.Signal,
) {
	select {
	case err := <-cliErr:
		if err != nil {
			log.Error("cli error", slog.Any("error", err))
		}
	case <-cliCtx.Done():
	case sig := <-stop:
		log.Error("shutting down application ", slog.String("signal", sig.String()))
		cancel()
	}
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = setupPrettySlog()
	case envDev:
		log = setupPrettySlog()
	case envProd:
		log = slogdiscard.NewDiscardLogger()
	default:
		log = slogdiscard.NewDiscardLogger()

	}
	return log
}

func setupPrettySlog() *slog.Logger {
	opts := slogpretty.PrettyHandlerOptions{
		SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
	}
	handler := opts.NewPrettyHandler(os.Stdout)

	return slog.New(handler)
}
