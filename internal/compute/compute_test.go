package compute_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	compute "lesson1/internal/compute"
	computemocks "lesson1/internal/compute/mocks"
	"lesson1/internal/database/dberrors"
	"lesson1/internal/lib/logger/slogdiscard"
)

var (
	errSetFailed = errors.New("set failed")
	errGetFailed = errors.New("get failed")
	errDelFailed = errors.New("del failed")
)

func newComputeWithMocks(t *testing.T) (*compute.Compute, *computemocks.MockCommandCompute, *computemocks.MockQueryCompute) {
	t.Helper()

	cmdMock := computemocks.NewMockCommandCompute(t)
	queryMock := computemocks.NewMockQueryCompute(t)
	logger := newTestLogger()

	combined := struct {
		compute.CommandCompute
		compute.QueryCompute
	}{
		CommandCompute: cmdMock,
		QueryCompute:   queryMock,
	}

	return compute.NewCompute(logger, combined), cmdMock, queryMock
}

func TestComputeHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		setup   func(ctx context.Context, cmd *computemocks.MockCommandCompute, q *computemocks.MockQueryCompute)
		want    string
		wantErr error
	}{
		{
			name:  "set ok",
			input: "SET key value",
			setup: func(ctx context.Context, cmd *computemocks.MockCommandCompute, _ *computemocks.MockQueryCompute) {
				cmd.EXPECT().Set(ctx, "key", "value").Return(nil)
			},
			want: "OK",
		},
		{
			name:  "set error",
			input: "SET key value",
			setup: func(ctx context.Context, cmd *computemocks.MockCommandCompute, _ *computemocks.MockQueryCompute) {
				cmd.EXPECT().Set(ctx, "key", "value").Return(errSetFailed)
			},
			wantErr: errSetFailed,
		},
		{
			name:  "get ok",
			input: "GET key",
			setup: func(ctx context.Context, _ *computemocks.MockCommandCompute, q *computemocks.MockQueryCompute) {
				q.EXPECT().Get(ctx, "key").Return("123", nil)
			},
			want: "VALUE 123",
		},
		{
			name:  "get not found",
			input: "GET key",
			setup: func(ctx context.Context, _ *computemocks.MockCommandCompute, q *computemocks.MockQueryCompute) {
				q.EXPECT().Get(ctx, "key").Return("", dberrors.ErrNotFound)
			},
			want: "NOT_FOUND",
		},
		{
			name:  "get error",
			input: "GET key",
			setup: func(ctx context.Context, _ *computemocks.MockCommandCompute, q *computemocks.MockQueryCompute) {
				q.EXPECT().Get(ctx, "key").Return("", errGetFailed)
			},
			wantErr: errGetFailed,
		},
		{
			name:  "del ok",
			input: "DEL key",
			setup: func(ctx context.Context, cmd *computemocks.MockCommandCompute, _ *computemocks.MockQueryCompute) {
				cmd.EXPECT().Del(ctx, "key").Return(nil)
			},
			want: "DELETED",
		},
		{
			name:  "del not found",
			input: "DEL key",
			setup: func(ctx context.Context, cmd *computemocks.MockCommandCompute, _ *computemocks.MockQueryCompute) {
				cmd.EXPECT().Del(ctx, "key").Return(dberrors.ErrNotFound)
			},
			want: "NOT_FOUND",
		},
		{
			name:  "del error",
			input: "DEL key",
			setup: func(ctx context.Context, cmd *computemocks.MockCommandCompute, _ *computemocks.MockQueryCompute) {
				cmd.EXPECT().Del(ctx, "key").Return(errDelFailed)
			},
			wantErr: errDelFailed,
		},
		{
			name:    "invalid command",
			input:   "BAD key",
			wantErr: compute.ErrInvalidCommand,
		},
		{
			name:    "invalid syntax command",
			input:   "set key value",
			wantErr: compute.ErrInvalidSyntaxCommand,
		},
		{
			name:    "invalid argument",
			input:   "SET key !",
			wantErr: compute.ErrInvalidSyntaxArg,
		},
		{
			name:    "empty command",
			input:   "   ",
			wantErr: compute.ErrEmptyCommand,
		},
		{
			name:    "invalid quantity set",
			input:   "SET onlykey",
			wantErr: compute.ErrInvalidQuantity,
		},
		{
			name:    "invalid quantity del",
			input:   "DEL",
			wantErr: compute.ErrInvalidQuantity,
		},
		{
			name:    "invalid quantity get",
			input:   "GET",
			wantErr: compute.ErrInvalidQuantity,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			c, cmdMock, queryMock := newComputeWithMocks(t)
			ctx := context.Background()

			if tc.setup != nil {
				tc.setup(ctx, cmdMock, queryMock)
			}

			got, err := c.ComputeHandler(ctx, tc.input)

			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				require.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func newTestLogger() *slog.Logger {
	return slogdiscard.NewDiscardLogger()
}
