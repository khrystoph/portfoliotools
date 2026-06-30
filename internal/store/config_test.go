package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
)

func TestSystemConfigStore_Get_ExistingKey(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewSystemConfigStore(pool)
	ctx := context.Background()

	// migration 009 seeds mcap_threshold_usd = 100000000
	val, err := s.Get(ctx, "mcap_threshold_usd")
	require.NoError(t, err)
	assert.Equal(t, "100000000", val)
}

func TestSystemConfigStore_Get_MissingKey(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewSystemConfigStore(pool)
	ctx := context.Background()

	_, err := s.Get(ctx, "nonexistent_key")
	require.Error(t, err)
	assert.True(t, store.IsNotFound(err))
}

func TestSystemConfigStore_Set_UpsertBehavior(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewSystemConfigStore(pool)
	ctx := context.Background()

	require.NoError(t, s.Set(ctx, "mcap_threshold_usd", "200000000"))
	val, err := s.Get(ctx, "mcap_threshold_usd")
	require.NoError(t, err)
	assert.Equal(t, "200000000", val)

	require.NoError(t, s.Set(ctx, "new_key", "hello"))
	val, err = s.Get(ctx, "new_key")
	require.NoError(t, err)
	assert.Equal(t, "hello", val)
}

func TestSystemConfigStore_GetInt64(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewSystemConfigStore(pool)
	ctx := context.Background()

	n, err := s.GetInt64(ctx, "mcap_threshold_usd", 0)
	require.NoError(t, err)
	assert.Equal(t, int64(100_000_000), n)
}

func TestSystemConfigStore_GetInt64_Fallback(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewSystemConfigStore(pool)
	ctx := context.Background()

	n, err := s.GetInt64(ctx, "no_such_key", int64(42))
	require.NoError(t, err)
	assert.Equal(t, int64(42), n)
}
