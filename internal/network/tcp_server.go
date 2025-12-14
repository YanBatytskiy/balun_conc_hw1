package network

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"spyder/internal/database"
	"time"

	"golang.org/x/sync/errgroup"
)

type TCPServer struct {
	listener net.Listener

	idleTimeout    time.Duration
	maxConnections int
	bufferSize     int

	log *slog.Logger
}

func NewTCPServer(address string, log *slog.Logger, options ...TCPServeOption) (*TCPServer, error) {
	// const op = "network.NewTCPServer"

	if log == nil {
		return nil, ErrInvalidLogger
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return nil, err
	}

	server := &TCPServer{
		listener: listener,
		log:      log,
	}

	for _, option := range options {
		option(server)
	}

	if server.maxConnections == 0 {
		return nil, ErrInvalidMaxConn
	}

	if server.bufferSize == 0 {
		return nil, ErrInvalidBufferSize
	}

	return server, nil
}

func (server *TCPServer) HandleClientQueries(ctx context.Context, database *database.Database) error {
	const op = "network.HandleClientQueries"

	group, ctx := errgroup.WithContext(ctx)
	defer server.listener.Close()

	server.log.Info("server listening", slog.String("address", server.listener.Addr().String()))

	go func() {
		<-ctx.Done()
		_ = server.listener.Close()
	}()

	var errorReturn error
	sem := make(chan struct{}, server.maxConnections)

	for {
		conn, errorReturn := server.listener.Accept()
		if errorReturn != nil {
			if ctx.Err() != nil { // ctx отменён → listener закрыт
				break
			}

			if errors.Is(errorReturn, net.ErrClosed) {
				errorReturn = nil
				break
			}

			var netError net.Error

			if errors.As(errorReturn, &netError) && netError.Timeout() {
				server.log.Debug("accept timeout",
					slog.String("operation", op),
					slog.String("error", netError.Error()),
				)
				time.Sleep(100 * time.Millisecond)
				continue
			}
			server.log.Debug("failed to accept connection",
				slog.String("operation", op),
				slog.String("error",
					errorReturn.Error()),
			)
			break
		}

		select {
		case sem <- struct{}{}:

			localConn := conn
			group.Go(func() error {
				defer func() { <-sem }()
				defer localConn.Close()
				return server.HandleConnections(ctx, localConn, database)
			})
		default:
			server.log.Debug("maximum connections reached, rejecting new connection")
			_ = conn.Close()
		}
	}

	err := group.Wait()
	if err != nil {
		errorReturn = err
	}

	return errorReturn
}

func (server *TCPServer) HandleConnections(ctx context.Context, conn net.Conn, database *database.Database) error {
	const op = "network.HandleConnections"

	var err error

	defer func() {
		if r := recover(); r != nil {
			server.log.Error("panic in connection handler",
				slog.String("operation", op),
				slog.Any("panic", r))
			err = fmt.Errorf("%s: panic: %v", op, r)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if server.idleTimeout != 0 {
			err = conn.SetReadDeadline(time.Now().Add(server.idleTimeout))
			if err != nil {
				server.log.Debug("failed to set read deadline", slog.String("operation", op))
				break
			}
		}

		query := make([]byte, server.bufferSize)

		num, err := conn.Read(query)
		if err != nil {

			if errors.Is(err, io.EOF) {
				err = nil
				break
			}
			return fmt.Errorf("network.HandleConnections: failed to read from connection: %w", err)
		}

		response, err := database.DatabaseHandler(ctx, string(query[:num]))
		if err != nil {
			_, writeErr := conn.Write([]byte(err.Error()))
			if writeErr != nil {
				return fmt.Errorf("%s: failed to write error to connection: %w", op, writeErr)
			}
			continue
		}

		if server.idleTimeout != 0 {
			err = conn.SetWriteDeadline(time.Now().Add(server.idleTimeout))
			if err != nil {
				server.log.Debug("failed to set write deadline", slog.String("operation", op))
				break
			}
		}
		_, err = conn.Write([]byte(response))
		if err != nil {
			return fmt.Errorf("%s: failed to write to connection: %w", op, err)
		}
	}
	return err
}
