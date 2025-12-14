package network

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWithClientOptions(t *testing.T) {
	t.Parallel()

	t.Run("address", func(t *testing.T) {
		client := &TCPClient{}
		WithClientTCPAddress("addr")(client)
		require.Equal(t, "addr", client.address)
	})

	t.Run("idle timeout positive", func(t *testing.T) {
		client := &TCPClient{}
		WithClientTCPIdleTimeout(2 * time.Second)(client)
		require.Equal(t, 2*time.Second, client.idleTimeout)
	})

	t.Run("idle timeout negative falls back", func(t *testing.T) {
		client := &TCPClient{}
		WithClientTCPIdleTimeout(-time.Second)(client)
		require.Equal(t, 5*time.Minute, client.idleTimeout)
	})

	t.Run("max message size positive", func(t *testing.T) {
		client := &TCPClient{}
		WithClientTCPMaxMessageSize(16)(client)
		require.Equal(t, 16, client.maxMessageSize)
	})

	t.Run("max message size negative falls back", func(t *testing.T) {
		client := &TCPClient{}
		WithClientTCPMaxMessageSize(-1)(client)
		require.Equal(t, 4098, client.maxMessageSize)
	})
}

func TestWithServerOptions(t *testing.T) {
	t.Parallel()

	t.Run("idle timeout positive", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPIdleTimeout(3 * time.Second)(server)
		require.Equal(t, 3*time.Second, server.idleTimeout)
	})

	t.Run("idle timeout negative falls back", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPIdleTimeout(-time.Second)(server)
		require.Equal(t, 5*time.Minute, server.idleTimeout)
	})

	t.Run("max connections positive", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPMaxConnectionNumber(10)(server)
		require.Equal(t, 10, server.maxConnections)
	})

	t.Run("max connections negative falls back", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPMaxConnectionNumber(-1)(server)
		require.Equal(t, 100, server.maxConnections)
	})

	t.Run("buffer size positive", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPBufferSize(64)(server)
		require.Equal(t, 64, server.bufferSize)
	})

	t.Run("buffer size negative falls back", func(t *testing.T) {
		server := &TCPServer{}
		WithServerTCPBufferSize(-1)(server)
		require.Equal(t, 4098, server.bufferSize)
	})
}
