package database_test

import (
	"context"
	"errors"
	"spyder/internal/database"
	"spyder/internal/database/compute"
	"spyder/internal/lib/logger/slogdiscard"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	databasemocks "spyder/internal/database/mocks"
)

var (
	errSetFailed = errors.New("set failed")
	errGetFailed = errors.New("get failed")
	errDelFailed = errors.New("del failed")
)

func newDatabaseWithMocks(
	t *testing.T,
) (*database.Database, *databasemocks.MockComputeLayer, *databasemocks.MockStorageLayer) {
	t.Helper()

	computeMock := databasemocks.NewMockComputeLayer(t)
	storageMock := databasemocks.NewMockStorageLayer(t)
	logger := slogdiscard.NewDiscardLogger()

	db, err := database.NewDatabase(logger, computeMock, storageMock)
	require.NoError(t, err)

	return db, computeMock, storageMock
}

func TestDatabaseHandler(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		raw      string
		parsed   []string
		parseErr error
		setup    func(ctx context.Context, storage *databasemocks.MockStorageLayer)
		want     string
		wantErr  error
		wantMsg  string
	}{
		{
			name:   "set ok",
			raw:    "SET key value",
			parsed: []string{compute.CommandSet, "key", "value"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Set(ctx, "key", "value").Return(nil)
			},
			want: "OK",
		},
		{
			name:   "set error",
			raw:    "SET key value",
			parsed: []string{compute.CommandSet, "key", "value"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Set(ctx, "key", "value").Return(errSetFailed)
			},
			wantErr: errSetFailed,
		},
		{
			name:   "get ok",
			raw:    "GET key",
			parsed: []string{compute.CommandGet, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Get(ctx, "key").Return("123", nil)
			},
			want: "VALUE 123",
		},
		{
			name:   "get not found",
			raw:    "GET key",
			parsed: []string{compute.CommandGet, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Get(ctx, "key").Return("", errors.New("HashTable.Get: not found"))
			},
			want: "NOT_FOUND",
		},
		{
			name:   "get error",
			raw:    "GET key",
			parsed: []string{compute.CommandGet, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Get(ctx, "key").Return("", errGetFailed)
			},
			wantErr: errGetFailed,
		},
		{
			name:   "del ok",
			raw:    "DEL key",
			parsed: []string{compute.CommandDel, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Del(ctx, "key").Return(nil)
			},
			want: "DELETED",
		},
		{
			name:   "del not found",
			raw:    "DEL key",
			parsed: []string{compute.CommandDel, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Del(ctx, "key").Return(errors.New("HashTable.Get: not found"))
			},
			want: "NOT_FOUND",
		},
		{
			name:   "del error",
			raw:    "DEL key",
			parsed: []string{compute.CommandDel, "key"},
			setup: func(ctx context.Context, storage *databasemocks.MockStorageLayer) {
				storage.EXPECT().Del(ctx, "key").Return(errDelFailed)
			},
			wantErr: errDelFailed,
		},
		{
			name:    "invalid command token",
			raw:     "BAD key",
			parsed:  []string{"BAD", "key"},
			wantMsg: "database.handler: invalid command",
		},
		{
			name:    "invalid quantity set",
			raw:     "SET onlykey",
			parsed:  []string{compute.CommandSet, "onlykey"},
			wantMsg: "compute.set: invalid quantity of arguments",
		},
		{
			name:    "invalid quantity del",
			raw:     "DEL",
			parsed:  []string{compute.CommandDel},
			wantMsg: "compute.del: invalid quantity of arguments",
		},
		{
			name:    "invalid quantity get",
			raw:     "GET",
			parsed:  []string{compute.CommandGet},
			wantMsg: "compute.get: invalid quantity of arguments",
		},
		{
			name:     "parse invalid command",
			raw:      "set key value",
			parseErr: compute.ErrInvalidCommand,
			wantErr:  compute.ErrInvalidCommand,
			wantMsg:  compute.ErrInvalidCommand.Error(),
		},
		{
			name:     "parse empty command",
			raw:      "   ",
			parseErr: compute.ErrEmptyCommand,
			wantErr:  compute.ErrEmptyCommand,
			wantMsg:  compute.ErrEmptyCommand.Error(),
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			db, computeMock, storageMock := newDatabaseWithMocks(t)
			ctx := context.Background()

			computeMock.EXPECT().ParseAndValidate(ctx, tc.raw).Return(tc.parsed, tc.parseErr)

			if tc.setup != nil {
				tc.setup(ctx, storageMock)
			}

			got, err := db.DatabaseHandler(ctx, tc.raw)

			if tc.wantErr != nil || tc.wantMsg != "" {
				require.Error(t, err)
				if tc.wantErr != nil {
					require.ErrorIs(t, err, tc.wantErr)
				}
				if tc.wantMsg != "" {
					require.EqualError(t, err, tc.wantMsg)
				}
				require.Empty(t, got)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNewDatabase(t *testing.T) {
	t.Parallel()

	logger := slogdiscard.NewDiscardLogger()
	computeMock := databasemocks.NewMockComputeLayer(t)
	storageMock := databasemocks.NewMockStorageLayer(t)

	db, err := database.NewDatabase(logger, computeMock, storageMock)
	require.NoError(t, err)
	require.NotNil(t, db)

	_, err = database.NewDatabase(nil, computeMock, storageMock)
	require.Error(t, err)

	_, err = database.NewDatabase(logger, nil, storageMock)
	require.Error(t, err)

	_, err = database.NewDatabase(logger, computeMock, nil)
	require.Error(t, err)
}
