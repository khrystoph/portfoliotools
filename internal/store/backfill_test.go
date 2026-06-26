package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
)

func TestBackfillStore_CreateAndComplete(t *testing.T) {
	pool := testutil.NewTestDB(t)
	bs := store.NewBackfillStore(pool)
	ctx := context.Background()

	runID, err := bs.Create(ctx)
	require.NoError(t, err)
	assert.Greater(t, runID, int64(0))

	err = bs.Complete(ctx, runID, 100, 2)
	require.NoError(t, err)

	run, err := bs.Get(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, store.BackfillStatusCompleted, run.Status)
	assert.Equal(t, 100, run.TickersProcessed)
	assert.Equal(t, 2, run.TickersFailed)
	assert.NotNil(t, run.CompletedAt)
}

func TestBackfillStore_Fail(t *testing.T) {
	pool := testutil.NewTestDB(t)
	bs := store.NewBackfillStore(pool)
	ctx := context.Background()

	runID, err := bs.Create(ctx)
	require.NoError(t, err)

	require.NoError(t, bs.Fail(ctx, runID, "connection timeout"))

	run, err := bs.Get(ctx, runID)
	require.NoError(t, err)
	assert.Equal(t, store.BackfillStatusFailed, run.Status)
	require.NotNil(t, run.ErrorMsg)
	assert.Equal(t, "connection timeout", *run.ErrorMsg)
}

func TestBackfillStore_LogTicker(t *testing.T) {
	pool := testutil.NewTestDB(t)
	bs := store.NewBackfillStore(pool)
	ts := store.NewTickerStore(pool)
	ctx := context.Background()

	runID, err := bs.Create(ctx)
	require.NoError(t, err)

	tickerID, err := ts.Upsert(ctx, store.Ticker{
		Symbol:        "GOOG",
		Name:          "Alphabet Inc.",
		AssetClass:    store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca,
		Currency:      "USD",
		Active:        true,
	})
	require.NoError(t, err)

	dur := 42
	log := store.BackfillTickerLog{
		RunID:         runID,
		TickerID:      tickerID,
		Status:        store.BackfillTickerSuccess,
		CandlesStored: 1,
		DurationMS:    &dur,
	}
	require.NoError(t, bs.LogTicker(ctx, log))
}
