package initialization

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"spyder/internal/config"
)

func TestNewInitilizer_NilConfig(t *testing.T) {
	t.Parallel()

	init, err := NewInitilizer(nil)
	require.Error(t, err)
	require.Nil(t, init)
}

func TestNewInitilizer_InvalidNetwork(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EngineType: "in_memory",
		Network: &config.NetworkConfig{
			Address:        "bad:addr",
			MaxConnections: 1,
			BufferSize:     1,
		},
		Logger: &config.LoggingConfig{
			Level: "prod",
		},
	}

	init, err := NewInitilizer(cfg)
	require.Error(t, err)
	require.Nil(t, init)
}

func TestStartDatabase_Cancel(t *testing.T) {
	t.Parallel()

	cfg := &config.Config{
		EngineType: "in_memory",
		Network: &config.NetworkConfig{
			Address: "127.0.0.1:0",
		},
		Logger: &config.LoggingConfig{
			Level: "prod",
		},
	}

	init, err := NewInitilizer(cfg)
	require.NoError(t, err)
	require.NotNil(t, init)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- init.StartDatabase(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		require.NoError(t, err)
	case <-time.After(time.Second):
		t.Fatal("StartDatabase did not return after cancel")
	}
}
