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

func TestTickerMigrationStore_Create(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)
	ctx := context.Background()

	fromID, err := ts.Upsert(ctx, store.Ticker{Symbol: "OSTK", Name: "Overstock.com",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: false})
	require.NoError(t, err)

	bbbyFIGI := "BBG00BBBY001"
	toID, err := ts.Upsert(ctx, store.Ticker{Symbol: "BBBY", Name: "Beyond Inc.",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: true, CompositeFIGI: &bbbyFIGI})
	require.NoError(t, err)

	mig := store.TickerMigration{
		FromTickerID:  fromID,
		ToTickerID:    toID,
		EffectiveDate: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
		Reason:        store.MigrationReasonRebranding,
		Source:        store.MigrationSourceOpenFIGI,
	}
	id, err := ms.Create(ctx, mig)
	require.NoError(t, err)
	assert.Positive(t, id)
}

func TestTickerMigrationStore_GetByFromTicker(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)
	ctx := context.Background()

	fromID, _ := ts.Upsert(ctx, store.Ticker{Symbol: "OSTK", Name: "Overstock.com",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: false})
	bbbyFIGI := "BBG00BBBY001"
	toID, _ := ts.Upsert(ctx, store.Ticker{Symbol: "BBBY", Name: "Beyond Inc.",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: true, CompositeFIGI: &bbbyFIGI})

	_, err := ms.Create(ctx, store.TickerMigration{
		FromTickerID: fromID, ToTickerID: toID,
		EffectiveDate: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
		Reason: store.MigrationReasonRebranding, Source: store.MigrationSourceOpenFIGI,
	})
	require.NoError(t, err)

	results, err := ms.GetByFromTicker(ctx, fromID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, toID, results[0].ToTickerID)
	assert.Equal(t, store.MigrationReasonRebranding, results[0].Reason)
}

func TestTickerMigrationStore_GetByToTicker(t *testing.T) {
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)
	ctx := context.Background()

	fromID, _ := ts.Upsert(ctx, store.Ticker{Symbol: "OSTK", Name: "Overstock.com",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: false})
	bbbyFIGI := "BBG00BBBY001"
	toID, _ := ts.Upsert(ctx, store.Ticker{Symbol: "BBBY", Name: "Beyond Inc.",
		AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca,
		Currency: "usd", Active: true, CompositeFIGI: &bbbyFIGI})

	_, err := ms.Create(ctx, store.TickerMigration{
		FromTickerID: fromID, ToTickerID: toID,
		EffectiveDate: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
		Reason: store.MigrationReasonRebranding, Source: store.MigrationSourceOpenFIGI,
	})
	require.NoError(t, err)

	results, err := ms.GetByToTicker(ctx, toID)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, fromID, results[0].FromTickerID)
}
