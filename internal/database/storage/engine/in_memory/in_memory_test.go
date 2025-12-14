package inmemory_test

import (
	"context"
	"errors"
	"spyder/internal/lib/logger/slogdiscard"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	inmemorystorage "spyder/internal/database/storage/engine/in_memory"
)

func TestEngineSetAndGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(e *inmemorystorage.Engine, ctx context.Context)
		key     string
		want    string
		wantErr string
	}{
		{
			name:  "get ok after set",
			setup: func(e *inmemorystorage.Engine, ctx context.Context) { require.NoError(t, e.Set(ctx, "foo", "bar")) },
			key:   "foo",
			want:  "bar",
		},
		{
			name:    "get not found",
			key:     "missing",
			wantErr: "not found",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			engine, err := inmemorystorage.NewEngine(slogdiscard.NewDiscardLogger())
			require.NoError(t, err)
			if tc.setup != nil {
				tc.setup(engine, ctx)
			}

			got, err := engine.Get(ctx, tc.key)
			if tc.wantErr != "" {
				require.Error(t, err)
				assert.True(t, errors.Is(err, inmemorystorage.ErrNotFound))
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
		setup   func(e *inmemorystorage.Engine, ctx context.Context)
		key     string
		wantErr error
	}{
		{
			name:  "del ok after set",
			setup: func(e *inmemorystorage.Engine, ctx context.Context) { require.NoError(t, e.Set(ctx, "foo", "bar")) },
			key:   "foo",
		},
		{
			name: "del not found",
			key:  "missing",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			engine, err := inmemorystorage.NewEngine(slogdiscard.NewDiscardLogger())
			require.NoError(t, err)
			if tc.setup != nil {
				tc.setup(engine, ctx)
			}

			err = engine.Del(ctx, tc.key)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}

			require.NoError(t, err)
			_, getErr := engine.Get(ctx, tc.key)
			assert.Error(t, getErr)
			assert.Contains(t, getErr.Error(), "not found")
		})
	}
}
