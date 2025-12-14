package initialization

import (
	"log/slog"
	"spyder/internal/config"
	"spyder/internal/network"
	"time"
)

var (
	defaultIdleTimeout    = time.Duration(10 * time.Minute)
	defaultMaxConnections = 10
	defaultBufferSize     = 4096
	defaultServerAddress  = "127.0.0.1:3323"
)

func CreateNetwork(log *slog.Logger, cfg *config.NetworkConfig) (*network.TCPServer, error) {
	if log == nil {
		return nil, network.ErrInvalidLogger
	}

	if cfg == nil {
		return nil, network.ErrInvalidConfig
	}

	var options []network.TCPServeOption

	if cfg.IdleTimeout != 0 {
		options = append(options, network.WithServerTCPIdleTimeout(cfg.IdleTimeout))
	} else {
		options = append(options, network.WithServerTCPIdleTimeout(time.Duration(defaultIdleTimeout)))
	}

	if cfg.MaxConnections != 0 {
		options = append(options, network.WithServerTCPMaxConnectionNumber(cfg.MaxConnections))
	} else {
		options = append(options, network.WithServerTCPMaxConnectionNumber(defaultMaxConnections))
	}

	if cfg.BufferSize != 0 {
		options = append(options, network.WithServerTCPBufferSize(cfg.BufferSize))
	} else {
		options = append(options, network.WithServerTCPBufferSize(defaultBufferSize))
	}

	address := ""

	if cfg.Address == "" {
		address = defaultServerAddress
	} else {
		address = cfg.Address
	}

	return network.NewTCPServer(address, log, options...)

}
