package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func writeTempConfig(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	require.NoError(t, os.WriteFile(path, []byte(contents), 0o644))
	return path
}

func TestNewConfig_DefaultsAppliedWhenMissingFields(t *testing.T) {
	path := writeTempConfig(t, `
network:
  engine_address: "127.0.0.1:5555"
logging:
  level: "prod"
`)
	t.Setenv("CONFIG_PATH", path)

	cfg, err := NewConfig()
	require.NoError(t, err)
	require.NotNil(t, cfg)

	require.Equal(t, "127.0.0.1:5555", cfg.Network.Address)
	require.Equal(t, 4096, cfg.Network.MaxMessageSize) // из дефолтов
	require.Equal(t, 100, cfg.Network.MaxConnections)  // из дефолтов
	require.Equal(t, 5*time.Minute, cfg.Network.IdleTimeout)
	require.Equal(t, 4096, cfg.Network.BufferSize)
	require.Equal(t, "tcp", cfg.Network.TypeConn)
	require.Equal(t, "in_memory", cfg.EngineType)
	require.Equal(t, "prod", cfg.Logger.Level)
}

func TestNewConfig_ValidationFailsOnNegative(t *testing.T) {
	path := writeTempConfig(t, `
network:
  engine_address: "127.0.0.1:5555"
  max_connections: -1
logging:
  level: "info"
`)
	t.Setenv("CONFIG_PATH", path)

	cfg, err := NewConfig()
	require.Error(t, err)
	require.Nil(t, cfg)
}
