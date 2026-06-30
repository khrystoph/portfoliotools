package universe_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/testutil"
	"github.com/khrystoph/portfoliotools/internal/universe"
)

type mockFIGI struct {
	results []universe.FIGIResult
	err     error
}

func (m *mockFIGI) MapFIGIs(_ context.Context, figis []string) ([]universe.FIGIResult, error) {
	return m.results, m.err
}

type mockPolygon struct {
	bySymbol map[string]universe.PolygonAsset
	listErr  error
}

func (m *mockPolygon) FetchActiveAssets(_ context.Context, _ int64) ([]universe.PolygonAsset, error) {
	var all []universe.PolygonAsset
	for _, a := range m.bySymbol {
		all = append(all, a)
	}
	return all, m.listErr
}

func (m *mockPolygon) FetchTickerDetails(_ context.Context, sym string) (universe.PolygonAsset, error) {
	if m.listErr != nil {
		return universe.PolygonAsset{}, m.listErr
	}
	a, ok := m.bySymbol[sym]
	if !ok {
		return universe.PolygonAsset{}, fmt.Errorf("ticker %s: %w", sym, universe.ErrNotFound)
	}
	return a, nil
}

type mockAlpaca struct {
	symbols map[string]struct{}
	err     error
}

func (m *mockAlpaca) ListActiveSymbols(_ context.Context) (map[string]struct{}, error) {
	return m.symbols, m.err
}

func makeSyncer(t *testing.T, pg universe.PolygonDiscoverer, fi universe.FIGIResolver, al universe.AlpacaAssetLister) (*universe.Syncer, *store.TickerStore, *store.TickerMigrationStore) {
	t.Helper()
	pool := testutil.NewTestDB(t)
	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)
	sc := store.NewSystemConfigStore(pool)
	cfg := universe.SyncConfig{
		McapThresholdUSD: 100_000_000,
		BackoffConfig: universe.BackoffConfig{
			InitialDelay: time.Millisecond,
			Multiplier:   2.0,
			Cap:          5 * time.Millisecond,
		},
	}
	return universe.NewSyncer(ts, ms, sc, pg, fi, al, cfg), ts, ms
}

func TestSyncer_Daily_Rename(t *testing.T) {
	ctx := context.Background()

	ostkFIGI := "BBG000BL1Q17"
	bbbyFIGI := "BBG00BBBY001"

	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"BBBY": {Ticker: "BBBY", Name: "Beyond Inc.", Market: "stocks", Type: "CS",
			CurrencyName: "usd", MarketCap: 200_000_000, CompositeFIGI: bbbyFIGI, Active: true},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{
		{CompositeFIGI: ostkFIGI, Ticker: "BBBY", Found: true},
	}}

	syncer, ts, ms := makeSyncer(t, polygon, figi, &mockAlpaca{symbols: map[string]struct{}{}})

	ostkID, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "OSTK", Name: "Overstock.com", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true,
		CompositeFIGI: &ostkFIGI,
	})
	require.NoError(t, err)

	require.NoError(t, syncer.Daily(ctx))

	ostk, err := ts.GetBySymbol(ctx, "OSTK", store.AssetClassEquity)
	require.NoError(t, err)
	assert.False(t, ostk.Active, "OSTK should be deactivated after rename")

	bbby, err := ts.GetBySymbol(ctx, "BBBY", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, bbby.Active)

	migs, err := ms.GetByFromTicker(ctx, ostkID)
	require.NoError(t, err)
	require.Len(t, migs, 1)
	assert.Equal(t, store.MigrationReasonRename, migs[0].Reason)
}

func TestSyncer_Daily_DelistingConfirmed(t *testing.T) {
	ctx := context.Background()
	deadFIGI := "BBG0000DEAD3"

	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"DEAD": {Ticker: "DEAD", Active: false, Market: "stocks", Type: "CS", CurrencyName: "usd"},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{
		{CompositeFIGI: deadFIGI, Found: false},
	}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, &mockAlpaca{symbols: map[string]struct{}{}})

	_, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "DEAD", Name: "Dead Corp", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourcePolygon, Currency: "usd", Active: true,
		CompositeFIGI: &deadFIGI,
	})
	require.NoError(t, err)

	require.NoError(t, syncer.Daily(ctx))

	dead, err := ts.GetBySymbol(ctx, "DEAD", store.AssetClassEquity)
	require.NoError(t, err)
	assert.False(t, dead.Active)
}

func TestSyncer_Daily_DelistingAmbiguous(t *testing.T) {
	ctx := context.Background()
	aaplFIGI := "BBG000B9XRY4"

	// OpenFIGI says FIGI missing, but Polygon says still active — ambiguous
	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"AAPL": {Ticker: "AAPL", Active: true, Market: "stocks", Type: "CS", CurrencyName: "usd"},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{
		{CompositeFIGI: aaplFIGI, Found: false},
	}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, &mockAlpaca{symbols: map[string]struct{}{}})

	_, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "AAPL", Name: "Apple Inc.", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true,
		CompositeFIGI: &aaplFIGI,
	})
	require.NoError(t, err)

	require.NoError(t, syncer.Daily(ctx))

	aapl, err := ts.GetBySymbol(ctx, "AAPL", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, aapl.Active, "AAPL should remain active when deactivation is ambiguous")
}

func TestSyncer_Daily_OpenFIGIUnreachable(t *testing.T) {
	ctx := context.Background()
	aaplFIGI := "BBG000B9XRY4"

	figi := &mockFIGI{err: errors.New("connection refused")}
	syncer, ts, _ := makeSyncer(t, &mockPolygon{}, figi, &mockAlpaca{symbols: map[string]struct{}{}})

	_, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "AAPL", Name: "Apple Inc.", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true,
		CompositeFIGI: &aaplFIGI,
	})
	require.NoError(t, err)

	ctx2, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	err = syncer.Daily(ctx2)
	require.Error(t, err)

	aapl, err := ts.GetBySymbol(ctx, "AAPL", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, aapl.Active, "no deactivations when OpenFIGI is unreachable")
}

func TestSyncer_Weekly_NetNewAsset(t *testing.T) {
	ctx := context.Background()
	aaplFIGI := "BBG000B9XRY4"

	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"AAPL": {Ticker: "AAPL", Name: "Apple Inc.", Market: "stocks", Type: "CS",
			CurrencyName: "usd", MarketCap: 3_000_000_000_000, CompositeFIGI: aaplFIGI, Active: true},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{}}
	alpaca := &mockAlpaca{symbols: map[string]struct{}{"AAPL": {}}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, alpaca)
	require.NoError(t, syncer.Weekly(ctx))

	aapl, err := ts.GetBySymbol(ctx, "AAPL", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, aapl.Active)
	assert.Equal(t, store.SourceAlpaca, aapl.PrimarySource)
	require.NotNil(t, aapl.CompositeFIGI)
	assert.Equal(t, aaplFIGI, *aapl.CompositeFIGI)
}

func TestSyncer_Weekly_BelowThresholdNotPinned(t *testing.T) {
	ctx := context.Background()
	tinyFIGI := "BBG0000TINY1"

	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"TINY": {Ticker: "TINY", Active: false, Market: "stocks", Type: "CS", CurrencyName: "usd"},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{}}
	alpaca := &mockAlpaca{symbols: map[string]struct{}{}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, alpaca)

	_, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "TINY", Name: "Tiny Corp", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true,
		CompositeFIGI: &tinyFIGI,
	})
	require.NoError(t, err)

	require.NoError(t, syncer.Weekly(ctx))

	tiny, err := ts.GetBySymbol(ctx, "TINY", store.AssetClassEquity)
	require.NoError(t, err)
	assert.False(t, tiny.Active, "TINY should be deactivated: below threshold, not pinned, Polygon confirms inactive")
}

func TestSyncer_Weekly_BelowThresholdPinned(t *testing.T) {
	ctx := context.Background()
	smolFIGI := "BBG0000SMOL2"

	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{
		"SMOL": {Ticker: "SMOL", Active: false, Market: "stocks", Type: "CS", CurrencyName: "usd"},
	}}
	figi := &mockFIGI{results: []universe.FIGIResult{}}
	alpaca := &mockAlpaca{symbols: map[string]struct{}{}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, alpaca)

	id, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "SMOL", Name: "Small Co", AssetClass: store.AssetClassEquity,
		PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true,
		CompositeFIGI: &smolFIGI,
	})
	require.NoError(t, err)
	require.NoError(t, ts.SetPinned(ctx, id, true))

	require.NoError(t, syncer.Weekly(ctx))

	smol, err := ts.GetBySymbol(ctx, "SMOL", store.AssetClassEquity)
	require.NoError(t, err)
	assert.True(t, smol.Active, "SMOL should be preserved: is_pinned=true")
}

func TestSyncer_Weekly_StaticAssetsSeeded(t *testing.T) {
	ctx := context.Background()
	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{}}
	figi := &mockFIGI{results: []universe.FIGIResult{}}
	alpaca := &mockAlpaca{symbols: map[string]struct{}{}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, alpaca)
	require.NoError(t, syncer.Weekly(ctx))

	gold, err := ts.GetBySymbol(ctx, "GC=F", store.AssetClassCommodity)
	require.NoError(t, err)
	assert.True(t, gold.Active)
	assert.Equal(t, store.SourceYahoo, gold.PrimarySource)

	tnx, err := ts.GetBySymbol(ctx, "^TNX", store.AssetClassBond)
	require.NoError(t, err)
	assert.True(t, tnx.Active)
}

func TestSyncer_Weekly_StaticAssetsNotDeactivated(t *testing.T) {
	ctx := context.Background()
	polygon := &mockPolygon{bySymbol: map[string]universe.PolygonAsset{}}
	figi := &mockFIGI{results: []universe.FIGIResult{}}
	alpaca := &mockAlpaca{symbols: map[string]struct{}{}}

	syncer, ts, _ := makeSyncer(t, polygon, figi, alpaca)

	_, err := ts.Upsert(ctx, store.Ticker{
		Symbol: "^GSPC", Name: "S&P 500", AssetClass: store.AssetClassIndex,
		PrimarySource: store.SourceYahoo, Currency: "usd", Active: true,
	})
	require.NoError(t, err)

	require.NoError(t, syncer.Weekly(ctx))

	gspc, err := ts.GetBySymbol(ctx, "^GSPC", store.AssetClassIndex)
	require.NoError(t, err)
	assert.True(t, gspc.Active, "static asset ^GSPC must not be deactivated by weekly sweep")
}
