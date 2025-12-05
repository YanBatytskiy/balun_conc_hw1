package storage_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lesson1/internal/database/dberrors"
	"lesson1/internal/database/storage"
	"lesson1/internal/database/storage/engine"
	"lesson1/internal/lib/logger/slogdiscard"
)

func TestStorage(t *testing.T) {
	t.Parallel()

	type (
		setupFn func(ctx context.Context, s *storage.Storage)
	)

	tests := []struct {
		name    string
		setup   setupFn
		run     func(ctx context.Context, s *storage.Storage) (string, error)
		want    string
		wantErr error
	}{
		{
			name: "set ok",
			run: func(ctx context.Context, s *storage.Storage) (string, error) {
				return "", s.Set(ctx, "k", "v")
			},
		},
		{
			name: "get ok after set",
			setup: func(ctx context.Context, s *storage.Storage) {
				require.NoError(t, s.Set(ctx, "k", "v"))
			},
			run: func(ctx context.Context, s *storage.Storage) (string, error) {
				return s.Get(ctx, "k")
			},
			want: "v",
		},
		{
			name:    "get not found",
			run:     func(ctx context.Context, s *storage.Storage) (string, error) { return s.Get(ctx, "missing") },
			wantErr: dberrors.ErrNotFound,
		},
		{
			name: "del ok after set",
			setup: func(ctx context.Context, s *storage.Storage) {
				require.NoError(t, s.Set(ctx, "k", "v"))
			},
			run: func(ctx context.Context, s *storage.Storage) (string, error) {
				return "", s.Del(ctx, "k")
			},
		},
		{
			name:    "del not found",
			run:     func(ctx context.Context, s *storage.Storage) (string, error) { return "", s.Del(ctx, "missing") },
			wantErr: dberrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			logger := slogdiscard.NewDiscardLogger()
			eng := engine.NewEngine(logger)
			s := storage.NewStorage(logger, eng)

			if tc.setup != nil {
				tc.setup(ctx, s)
			}

			got, err := tc.run(ctx, s)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
