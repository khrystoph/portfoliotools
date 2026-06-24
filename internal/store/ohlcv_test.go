package store_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
)

func setupTickerForOHLCV(t *testing.T, ts *store.TickerStore) int64 {
	t.Helper()
	id, err := ts.Upsert(context.Background(), store.Ticker{
		Symbol:        "TEST",
		Name:          "Test Corp",
		AssetClass:    store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca,
		Currency:      "USD",
		Active:        true,
	})
	require.NoError(t, err)
	return id
}

func TestOHLCVStore_Upsert(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	os := store.NewOHLCVStore(pool)
	ctx := context.Background()

	tickerID := setupTickerForOHLCV(t, ts)
	date := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)

	candle := store.OHLCVDaily{
		TickerID:  tickerID,
		TradeDate: date,
		Open:      150.0,
		High:      155.0,
		Low:       149.0,
		Close:     153.0,
		Volume:    1_000_000,
		Source:    store.SourceAlpaca,
	}

	require.NoError(t, os.Upsert(ctx, candle))

	// Upsert same date with new close — must overwrite
	candle.Close = 154.0
	require.NoError(t, os.Upsert(ctx, candle))

	rows, err := os.GetRange(ctx, tickerID, 10)
	require.NoError(t, err)
	require.Len(t, rows, 1)
	assert.Equal(t, 154.0, rows[0].Close)
}

func TestOHLCVStore_UpsertBatch(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	os := store.NewOHLCVStore(pool)
	ctx := context.Background()

	tickerID := setupTickerForOHLCV(t, ts)

	candles := make([]store.OHLCVDaily, 5)
	for i := range candles {
		candles[i] = store.OHLCVDaily{
			TickerID:  tickerID,
			TradeDate: time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC),
			Open:      float64(100 + i),
			High:      float64(105 + i),
			Low:       float64(98 + i),
			Close:     float64(102 + i),
			Volume:    500_000,
			Source:    store.SourceAlpaca,
		}
	}

	require.NoError(t, os.UpsertBatch(ctx, candles))

	rows, err := os.GetRange(ctx, tickerID, 10)
	require.NoError(t, err)
	assert.Len(t, rows, 5)
}

func TestOHLCVStore_UpsertBatch_RollsBackOnError(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	os := store.NewOHLCVStore(pool)
	ctx := context.Background()

	tickerID := setupTickerForOHLCV(t, ts)

	// First candle is valid; second has ticker_id=0 (violates FK constraint)
	candles := []store.OHLCVDaily{
		{
			TickerID:  tickerID,
			TradeDate: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			Open:      100, High: 105, Low: 98, Close: 102,
			Volume: 500_000,
			Source: store.SourceAlpaca,
		},
		{
			TickerID:  0, // invalid — FK violation
			TradeDate: time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			Open:      100, High: 105, Low: 98, Close: 102,
			Volume: 500_000,
			Source: store.SourceAlpaca,
		},
	}

	err := os.UpsertBatch(ctx, candles)
	assert.Error(t, err, "batch with invalid FK must fail")

	// Entire batch must be rolled back — no rows written
	rows, err := os.GetRange(ctx, tickerID, 10)
	require.NoError(t, err)
	assert.Empty(t, rows, "rollback must leave no rows")
}

func TestOHLCVStore_GetRange_ReturnsDescendingOrder(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	os := store.NewOHLCVStore(pool)
	ctx := context.Background()

	tickerID := setupTickerForOHLCV(t, ts)

	for i := 0; i < 10; i++ {
		require.NoError(t, os.Upsert(ctx, store.OHLCVDaily{
			TickerID:  tickerID,
			TradeDate: time.Date(2025, 1, i+1, 0, 0, 0, 0, time.UTC),
			Open:      100, High: 105, Low: 98, Close: 102,
			Volume: 500_000,
			Source: store.SourceAlpaca,
		}))
	}

	rows, err := os.GetRange(ctx, tickerID, 5)
	require.NoError(t, err)
	require.Len(t, rows, 5)

	// Most recent date first
	for i := 1; i < len(rows); i++ {
		assert.True(t, rows[i-1].TradeDate.After(rows[i].TradeDate),
			"rows must be in descending date order")
	}
}
