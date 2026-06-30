package testutil_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
)

func TestNewTestDBWithFixtures(t *testing.T) {
	pool := testutil.NewTestDBWithFixtures(t)
	ts := store.NewTickerStore(pool)
	ctx := context.Background()

	all, err := ts.ListActive(ctx, "")
	require.NoError(t, err)
	// AAPL, SPY, BTC, GC=F, ^GSPC, EURUSD=X, ^TNX, TINY, SMOL, BBBY = 10 active
	assert.Len(t, all, 10)

	smol, err := ts.GetBySymbol(ctx, "SMOL", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, smol.IsPinned)

	dead, err := ts.GetBySymbol(ctx, "DEAD", store.AssetClassEquity)
	require.NoError(t, err)
	assert.False(t, dead.Active)
}
