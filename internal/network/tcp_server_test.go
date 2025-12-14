package network

import (
	"context"
	"errors"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"spyder/internal/database"
	"spyder/internal/database/compute"
	databasemocks "spyder/internal/database/mocks"
	"spyder/internal/database/storage"
	inmemory "spyder/internal/database/storage/engine/in_memory"
	"spyder/internal/lib/logger/slogdiscard"
)

func newDB(t *testing.T) *database.Database {
	t.Helper()

	log := slogdiscard.NewDiscardLogger()
	comp, err := compute.NewCompute(log)
	require.NoError(t, err)

	eng, err := inmemory.NewEngine(log)
	require.NoError(t, err)

	stor, err := storage.NewStorage(log, eng)
	require.NoError(t, err)

	db, err := database.NewDatabase(log, comp, stor)
	require.NoError(t, err)
	return db
}

func newMockDB(t *testing.T) (*database.Database, *databasemocks.MockComputeLayer, *databasemocks.MockStorageLayer) {
	t.Helper()

	computeMock := databasemocks.NewMockComputeLayer(t)
	storageMock := databasemocks.NewMockStorageLayer(t)
	logger := slogdiscard.NewDiscardLogger()

	db, err := database.NewDatabase(logger, computeMock, storageMock)
	require.NoError(t, err)

	return db, computeMock, storageMock
}

type stubListener struct {
	mu    sync.Mutex
	conns []net.Conn
	idx   int
}

func (s *stubListener) Accept() (net.Conn, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.idx >= len(s.conns) {
		return nil, net.ErrClosed
	}
	c := s.conns[s.idx]
	s.idx++
	return c, nil
}

func (s *stubListener) Close() error {
	return nil
}

func (s *stubListener) Addr() net.Addr {
	return &net.TCPAddr{}
}

func TestNewTCPServer_Table(t *testing.T) {
	t.Parallel()

	type args struct {
		addr    string
		options []TCPServeOption
		logNil  bool
	}

	tests := []struct {
		name      string
		args      args
		wantErr   bool
		errTarget error
	}{
		{
			name: "nil logger",
			args: args{
				addr:    "127.0.0.1:0",
				logNil:  true,
				options: []TCPServeOption{WithServerTCPBufferSize(1), WithServerTCPMaxConnectionNumber(1)},
			},
			wantErr:   true,
			errTarget: ErrInvalidLogger,
		},
		{
			name: "invalid address format",
			args: args{
				addr:    "127.0.0.1",
				options: []TCPServeOption{WithServerTCPBufferSize(1), WithServerTCPMaxConnectionNumber(1)},
			},
			wantErr: true,
		},
		{
			name: "address already in use",
			args: func() args {
				ln, err := net.Listen("tcp", "127.0.0.1:0")
				require.NoError(t, err)
				t.Cleanup(func() { _ = ln.Close() })
				return args{
					addr:    ln.Addr().String(),
					options: []TCPServeOption{WithServerTCPBufferSize(1), WithServerTCPMaxConnectionNumber(1)},
				}
			}(),
			wantErr: true,
		},
		{
			name: "zero buffer size",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(0), WithServerTCPMaxConnectionNumber(1)},
			},
			wantErr:   true,
			errTarget: ErrInvalidBufferSize,
		},
		{
			name: "zero max connections",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(1), WithServerTCPMaxConnectionNumber(0)},
			},
			wantErr:   true,
			errTarget: ErrInvalidMaxConn,
		},
		{
			name: "success minimal",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(8), WithServerTCPMaxConnectionNumber(1)},
			},
		},
		{
			name: "success with idle timeout",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(16), WithServerTCPMaxConnectionNumber(2), WithServerTCPIdleTimeout(time.Second)},
			},
		},
		{
			name: "large buffer and connections",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(8192), WithServerTCPMaxConnectionNumber(50)},
			},
		},
		{
			name: "default buffer error when not set",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPMaxConnectionNumber(5)},
			},
			wantErr:   true,
			errTarget: ErrInvalidBufferSize,
		},
		{
			name: "default max conn error when not set",
			args: args{
				addr:    "127.0.0.1:0",
				options: []TCPServeOption{WithServerTCPBufferSize(128)},
			},
			wantErr:   true,
			errTarget: ErrInvalidMaxConn,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var log = slogdiscard.NewDiscardLogger()
			if tc.args.logNil {
				log = nil
			}

			srv, err := NewTCPServer(tc.args.addr, log, tc.args.options...)
			if tc.wantErr {
				require.Error(t, err)
				if tc.errTarget != nil {
					require.ErrorIs(t, err, tc.errTarget)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, srv)
			_ = srv.listener.Close()
		})
	}
}

func TestHandleConnections_Table(t *testing.T) {
	logger := slogdiscard.NewDiscardLogger()

	tests := []struct {
		name          string
		buffer        int
		query         string
		preCancel     bool
		wantResp      string
		wantErr       bool
		errContains   string
		errIsCanceled bool
	}{
		{
			name:          "context canceled before read",
			buffer:        16,
			preCancel:     true,
			wantErr:       true,
			errIsCanceled: true,
		},
		{
			name:          "set ok",
			buffer:        32,
			query:         "SET foo bar",
			wantResp:      "OK",
			errIsCanceled: true,
		},
		{
			name:        "invalid command",
			buffer:      32,
			query:       "BAD",
			wantErr:     true,
			errContains: "invalid command",
		},
		{
			name:        "truncated command due to small buffer",
			buffer:      4,
			query:       "SET foo bar",
			wantErr:     true,
			errContains: "invalid quantity",
		},
		{
			name:          "handle get not found",
			buffer:        32,
			query:         "GET missing",
			wantResp:      "NOT_FOUND",
			errIsCanceled: true,
		},
		{
			name:          "handle del ok",
			buffer:        32,
			query:         "DEL missing",
			wantResp:      "DELETED",
			errIsCanceled: true,
		},
		{
			name:          "multiple commands same connection stop on cancel",
			buffer:        64,
			query:         "SET a b",
			wantResp:      "OK",
			errIsCanceled: true,
		},
		{
			name:          "buffer exact size command",
			buffer:        len("SET x y"),
			query:         "SET x y",
			wantResp:      "OK",
			errIsCanceled: true,
		},
		{
			name:          "buffer larger than command",
			buffer:        128,
			query:         "SET large buf",
			wantResp:      "OK",
			errIsCanceled: true,
		},
		{
			name:          "get after set using same db",
			buffer:        64,
			query:         "SET q w",
			wantResp:      "OK",
			errIsCanceled: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			db := newDB(t)

			server := &TCPServer{
				bufferSize: tc.buffer,
				log:        logger,
			}

			ctx, cancel := context.WithCancel(context.Background())
			if tc.preCancel {
				cancel()
			}
			defer cancel()

			c1, c2 := net.Pipe()
			defer c2.Close()

			errCh := make(chan error, 1)
			go func() {
				defer c1.Close()
				errCh <- server.HandleConnections(ctx, c1, db)
			}()

			if tc.query != "" && !tc.preCancel {
				_ = c2.SetWriteDeadline(time.Now().Add(time.Second))
				_, _ = c2.Write([]byte(tc.query))
				_ = c2.SetReadDeadline(time.Now().Add(time.Second))
				buf := make([]byte, tc.buffer)
				n, _ := c2.Read(buf)
				if tc.wantResp != "" {
					require.Equal(t, tc.wantResp, string(buf[:n]))
				}
			}

			if !tc.preCancel {
				cancel()
			}

			err := <-errCh
			if tc.wantErr {
				require.Error(t, err)
				if tc.errContains != "" {
					require.ErrorContains(t, err, tc.errContains)
				}
				if tc.errIsCanceled {
					require.True(t, errors.Is(err, context.Canceled))
				}
				return
			}

			if tc.errIsCanceled {
				require.True(t, errors.Is(err, context.Canceled))
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestHandleClientQueries_MaxConnectionsRejectsExtra(t *testing.T) {
	t.Parallel()

	srvConn1, cliConn1 := net.Pipe()
	srvConn2, cliConn2 := net.Pipe()

	listener := &stubListener{conns: []net.Conn{srvConn1, srvConn2}}

	server := &TCPServer{
		listener:       listener,
		bufferSize:     4,
		maxConnections: 1,
		log:            slogdiscard.NewDiscardLogger(),
	}

	db := newDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- server.HandleClientQueries(ctx, db)
	}()

	time.Sleep(50 * time.Millisecond)

	_ = cliConn2.SetReadDeadline(time.Now().Add(200 * time.Millisecond))
	buf := make([]byte, 1)
	_, err := cliConn2.Read(buf)
	require.Error(t, err)

	cancel()
	_ = cliConn1.Close()
	_ = cliConn2.Close()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("HandleClientQueries did not return after cancel")
	}
}

func TestHandleConnections_ReadTimeout(t *testing.T) {
	t.Parallel()

	server := &TCPServer{
		bufferSize:  8,
		idleTimeout: 10 * time.Millisecond,
		log:         slogdiscard.NewDiscardLogger(),
	}

	db := newDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	srv, cli := net.Pipe()
	defer cli.Close()

	errCh := make(chan error, 1)
	go func() {
		defer srv.Close()
		errCh <- server.HandleConnections(ctx, srv, db)
	}()

	// Не пишем в соединение, ждём таймаут чтения.
	err := <-errCh
	require.Error(t, err)
}

func TestHandleConnections_EOFIsNotError(t *testing.T) {
	t.Parallel()

	server := &TCPServer{
		bufferSize: 8,
		log:        slogdiscard.NewDiscardLogger(),
	}
	db := newDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c1, c2 := net.Pipe()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		defer c1.Close()
		errCh <- server.HandleConnections(ctx, c1, db)
	}()

	_ = c2.Close() // закрываем сразу, сервер должен вернуть nil, а не ошибку

	err := <-errCh
	require.NoError(t, err)
}

func TestHandleConnections_RecoversFromPanic(t *testing.T) {
	t.Parallel()

	server := &TCPServer{
		bufferSize: 16,
		log:        slogdiscard.NewDiscardLogger(),
	}

	db, computeMock, _ := newMockDB(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	computeMock.EXPECT().ParseAndValidate(mock.Anything, "PANIC").RunAndReturn(func(context.Context, string) ([]string, error) {
		panic("boom")
	})

	c1, c2 := net.Pipe()
	defer c2.Close()

	errCh := make(chan error, 1)
	go func() {
		defer c1.Close()
		errCh <- server.HandleConnections(ctx, c1, db)
	}()

	_, _ = c2.Write([]byte("PANIC"))

	err := <-errCh
	require.Error(t, err)
	require.Contains(t, err.Error(), "panic")
}
