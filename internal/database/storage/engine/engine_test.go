package engine_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"lesson1/internal/database/dberrors"
	"lesson1/internal/database/storage/engine"
	"lesson1/internal/lib/logger/slogdiscard"
)

func TestEngineSetAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(e *engine.Engine, ctx context.Context)
		key     string
		want    string
		wantErr error
	}{
		{
			name:  "get ok after set",
			setup: func(e *engine.Engine, ctx context.Context) { require.NoError(t, e.Set(ctx, "foo", "bar")) },
			key:   "foo",
			want:  "bar",
		},
		{
			name:    "get not found",
			key:     "missing",
			wantErr: dberrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			e := engine.NewEngine(slogdiscard.NewDiscardLogger())
			if tc.setup != nil {
				tc.setup(e, ctx)
			}

			got, err := e.Get(ctx, tc.key)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestEngineDel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(e *engine.Engine, ctx context.Context)
		key     string
		wantErr error
	}{
		{
			name:  "del ok after set",
			setup: func(e *engine.Engine, ctx context.Context) { require.NoError(t, e.Set(ctx, "foo", "bar")) },
			key:   "foo",
		},
		{
			name:    "del not found",
			key:     "missing",
			wantErr: dberrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			e := engine.NewEngine(slogdiscard.NewDiscardLogger())
			if tc.setup != nil {
				tc.setup(e, ctx)
			}

			err := e.Del(ctx, tc.key)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			_, getErr := e.Get(ctx, tc.key)
			assert.ErrorIs(t, getErr, dberrors.ErrNotFound)
		})
	}
}
