package initialization

import (
	"log/slog"
	"os"
	"spyder/internal/config"
	"spyder/internal/lib/logger/slogdiscard"
	"spyder/internal/lib/logger/slogpretty"
)

const (
	LoggerLevelInfo = "info"
	LoggerLevelDev  = "dev"
	LoggerLevelProd = "prod"
)

func CreateLogger(cfg *config.Config) (*slog.Logger, error) {

	var log *slog.Logger

	switch cfg.Logger.Level {
	case LoggerLevelInfo:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelInfo},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)

	case LoggerLevelDev:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelDebug},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)

	case LoggerLevelProd:
		log = slogdiscard.NewDiscardLogger()
	default:
		opts := slogpretty.PrettyHandlerOptions{
			SlogOpts: &slog.HandlerOptions{Level: slog.LevelInfo},
		}
		handler := opts.NewPrettyHandler(os.Stdout)

		log = slog.New(handler)

	}
	log.Info("starting service", slog.String("logger level", cfg.Logger.Level))
	return log, nil
}
