package initialization

import (
	"log/slog"
	inmemory "spyder/internal/database/storage/engine/in_memory"
)

func CreateEngine(log * slog.Logger) (*inmemory.Engine, error) {
	return inmemory.NewEngine(log)
}