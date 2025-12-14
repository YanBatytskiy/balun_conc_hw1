package storage_test

import (
	"context"
	"errors"
	"spyder/internal/database/storage"
	"spyder/internal/lib/logger/slogdiscard"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inmemorystorage "spyder/internal/database/storage/engine/in_memory"
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
			wantErr: inmemorystorage.ErrNotFound,
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
			name: "del not found",
			run:  func(ctx context.Context, s *storage.Storage) (string, error) { return "", s.Del(ctx, "missing") },
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			logger := slogdiscard.NewDiscardLogger()
			eng, err := inmemorystorage.NewEngine(logger)
			require.NoError(t, err)

			s, err := storage.NewStorage(logger, eng)
			require.NoError(t, err)

			if tc.setup != nil {
				tc.setup(ctx, s)
			}

			got, err := tc.run(ctx, s)

			if tc.wantErr != nil {
				require.Error(t, err)
				assert.True(t, errors.Is(err, tc.wantErr))
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}
