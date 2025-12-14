package compute_test

import (
	"context"
	"spyder/internal/database/compute"
	"spyder/internal/lib/logger/slogdiscard"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseAndValidate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		raw     string
		want    []string
		wantErr error
	}{
		{
			name: "set ok",
			raw:  "SET key value",
			want: []string{"SET", "key", "value"},
		},
		{
			name: "trims command",
			raw:  "   GET   key   ",
			want: []string{"GET", "key"},
		},
		{
			name:    "invalid command syntax",
			raw:     "set key value",
			wantErr: compute.ErrInvalidCommand,
		},
		{
			name:    "invalid argument syntax",
			raw:     "SET key !",
			wantErr: compute.ErrInvalidArgument,
		},
		{
			name:    "empty command",
			raw:     "   ",
			wantErr: compute.ErrEmptyCommand,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			logger := slogdiscard.NewDiscardLogger()
			c, err := compute.NewCompute(logger)
			require.NoError(t, err)

			tokens, err := c.ParseAndValidate(context.Background(), tc.raw)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Nil(t, tokens)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, tokens)
		})
	}
}
