package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
)

func TestTickerStore_Upsert(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewTickerStore(pool)
	ctx := context.Background()

	ticker := store.Ticker{
		Symbol:        "AAPL",
		Name:          "Apple Inc.",
		AssetClass:    store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca,
		Currency:      "USD",
		Active:        true,
	}

	id, err := s.Upsert(ctx, ticker)
	require.NoError(t, err)
	assert.Greater(t, id, int64(0))

	// Upsert again — should update, not duplicate
	ticker.Name = "Apple Inc. (updated)"
	id2, err := s.Upsert(ctx, ticker)
	require.NoError(t, err)
	assert.Equal(t, id, id2, "upsert of same symbol+class must return same ID")
}

func TestTickerStore_ListActive(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewTickerStore(pool)
	ctx := context.Background()

	tickers := []store.Ticker{
		{Symbol: "AAPL", Name: "Apple Inc.", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "USD", Active: true},
		{Symbol: "BTC",  Name: "Bitcoin",    AssetClass: store.AssetClassCrypto,  PrimarySource: store.SourceAlpaca, Currency: "USD", Active: true},
		{Symbol: "DEAD", Name: "Inactive",   AssetClass: store.AssetClassEquity,  PrimarySource: store.SourceAlpaca, Currency: "USD", Active: false},
	}
	for _, tk := range tickers {
		_, err := s.Upsert(ctx, tk)
		require.NoError(t, err)
	}

	// All active
	all, err := s.ListActive(ctx, "")
	require.NoError(t, err)
	assert.Len(t, all, 2)

	// Filtered by class
	equities, err := s.ListActive(ctx, store.AssetClassEquity)
	require.NoError(t, err)
	assert.Len(t, equities, 1)
	assert.Equal(t, "AAPL", equities[0].Symbol)
}

func TestTickerStore_GetBySymbol(t *testing.T) {
	pool := testutil.NewTestDB(t)
	s := store.NewTickerStore(pool)
	ctx := context.Background()

	_, err := s.Upsert(ctx, store.Ticker{
		Symbol:        "MSFT",
		Name:          "Microsoft Corp.",
		AssetClass:    store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca,
		Currency:      "USD",
		Active:        true,
	})
	require.NoError(t, err)

	got, err := s.GetBySymbol(ctx, "MSFT", store.AssetClassEquity)
	require.NoError(t, err)
	assert.Equal(t, "MSFT", got.Symbol)
	assert.Equal(t, "Microsoft Corp.", got.Name)

	_, err = s.GetBySymbol(ctx, "NOPE", store.AssetClassEquity)
	assert.Error(t, err, "missing ticker must return error")
	assert.True(t, store.IsNotFound(err), "missing ticker must satisfy IsNotFound")
}
