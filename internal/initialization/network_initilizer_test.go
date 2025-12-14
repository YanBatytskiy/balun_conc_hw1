package initialization

import (
	"net"
	"reflect"
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/require"

	"spyder/internal/config"
	"spyder/internal/lib/logger/slogdiscard"
	"spyder/internal/network"
)

func TestCreateNetwork_NilArgs(t *testing.T) {
	t.Parallel()

	_, err := CreateNetwork(nil, &config.NetworkConfig{})
	require.ErrorIs(t, err, network.ErrInvalidLogger)

	logger := slogdiscard.NewDiscardLogger()
	_, err = CreateNetwork(logger, nil)
	require.ErrorIs(t, err, network.ErrInvalidConfig)
}

func TestCreateNetwork_UsesDefaultsWhenZero(t *testing.T) {
	t.Parallel()

	logger := slogdiscard.NewDiscardLogger()
	cfg := &config.NetworkConfig{}

	srv, err := CreateNetwork(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	ln := extractListener(t, srv)
	require.Equal(t, defaultServerAddress, ln.Addr().String())
	require.Equal(t, defaultMaxConnections, extractIntField(t, srv, "maxConnections"))
	require.Equal(t, defaultBufferSize, extractIntField(t, srv, "bufferSize"))
	require.Equal(t, defaultIdleTimeout, extractDurationField(t, srv, "idleTimeout"))

	_ = ln.Close()
}

func TestCreateNetwork_UsesConfigValues(t *testing.T) {
	t.Parallel()

	logger := slogdiscard.NewDiscardLogger()
	cfg := &config.NetworkConfig{
		Address:        "127.0.0.1:0",
		MaxConnections: 5,
		BufferSize:     128,
		IdleTimeout:    time.Second,
	}

	srv, err := CreateNetwork(logger, cfg)
	require.NoError(t, err)
	require.NotNil(t, srv)

	ln := extractListener(t, srv)
	require.Equal(t, cfg.MaxConnections, extractIntField(t, srv, "maxConnections"))
	require.Equal(t, cfg.BufferSize, extractIntField(t, srv, "bufferSize"))
	require.Equal(t, cfg.IdleTimeout, extractDurationField(t, srv, "idleTimeout"))

	_ = ln.Close()
}

func extractListener(t *testing.T, srv *network.TCPServer) net.Listener {
	t.Helper()
	v := reflect.ValueOf(srv).Elem().FieldByName("listener")
	ptr := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	return ptr.Interface().(net.Listener)
}

func extractIntField(t *testing.T, srv *network.TCPServer, name string) int {
	t.Helper()
	v := reflect.ValueOf(srv).Elem().FieldByName(name)
	ptr := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	return int(ptr.Int())
}

func extractDurationField(t *testing.T, srv *network.TCPServer, name string) time.Duration {
	t.Helper()
	v := reflect.ValueOf(srv).Elem().FieldByName(name)
	ptr := reflect.NewAt(v.Type(), unsafe.Pointer(v.UnsafeAddr())).Elem()
	return time.Duration(ptr.Int())
}
