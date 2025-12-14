package network

import "time"

type TCPClientOption func(*TCPClient)

type TCPServeOption func(*TCPServer)

func WithClientTCPAddress(address string) TCPClientOption {
	return func(client *TCPClient) {
		client.address = address
	}
}

func WithClientTCPIdleTimeout(timeout time.Duration) TCPClientOption {
	return func(client *TCPClient) {
		if timeout < 0 {
			timeout = 5 * time.Minute
		}
		client.idleTimeout = timeout
	}
}

func WithClientTCPMaxMessageSize(maxMessageSize int) TCPClientOption {
	return func(client *TCPClient) {
		if maxMessageSize < 0 {
			maxMessageSize = 4098
		}
		client.maxMessageSize = maxMessageSize
	}
}

func WithServerTCPIdleTimeout(timeout time.Duration) TCPServeOption {
	return func(server *TCPServer) {
		if timeout < 0 {
			timeout = 5 * time.Minute
		}
		server.idleTimeout = timeout
	}
}

func WithServerTCPMaxConnectionNumber(maxConnections int) TCPServeOption {
	return func(server *TCPServer) {
		if maxConnections < 0 {
			maxConnections = 100
		}
		server.maxConnections = maxConnections
	}
}

func WithServerTCPBufferSize(bufferSize int) TCPServeOption {
	return func(server *TCPServer) {
		if bufferSize < 0 {
			bufferSize = 4098
		}
		server.bufferSize = bufferSize
	}
}
