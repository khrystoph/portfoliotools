package testutil

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
)

// NewTestDBWithFixtures returns a test DB seeded with a 12-ticker representative universe
// covering all asset classes, both active/inactive states, pinned/unpinned, with/without FIGI,
// and a recorded OSTK→BBBY migration.
func NewTestDBWithFixtures(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := NewTestDB(t)
	ctx := context.Background()
	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)

	aaplFIGI := "BBG000B9XRY4"
	spyFIGI := "BBG000BDTBL9"
	btcFIGI := "BBG00H8NHC57"
	tinyFIGI := "BBG0000TINY1"
	smolFIGI := "BBG0000SMOL2"
	deadFIGI := "BBG0000DEAD3"
	ostkFIGI := "BBG000BL1Q17"
	bbbyFIGI := "BBG00BBBY001"

	rows := []store.Ticker{
		{Symbol: "AAPL", Name: "Apple Inc.", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &aaplFIGI},
		{Symbol: "SPY", Name: "SPDR S&P 500 ETF Trust", AssetClass: store.AssetClassETF, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &spyFIGI},
		{Symbol: "BTC", Name: "Bitcoin", AssetClass: store.AssetClassCrypto, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &btcFIGI},
		{Symbol: "GC=F", Name: "Gold", AssetClass: store.AssetClassCommodity, PrimarySource: store.SourceYahoo, Currency: "usd", Active: true},
		{Symbol: "^GSPC", Name: "S&P 500", AssetClass: store.AssetClassIndex, PrimarySource: store.SourceYahoo, Currency: "usd", Active: true},
		{Symbol: "EURUSD=X", Name: "EUR/USD", AssetClass: store.AssetClassForex, PrimarySource: store.SourceYahoo, Currency: "usd", Active: true},
		{Symbol: "^TNX", Name: "10-Year Treasury Yield", AssetClass: store.AssetClassBond, PrimarySource: store.SourceYahoo, Currency: "usd", Active: true},
		{Symbol: "TINY", Name: "Tiny Corp", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &tinyFIGI},
		{Symbol: "SMOL", Name: "Small Co", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &smolFIGI},
		{Symbol: "DEAD", Name: "Dead Corp", AssetClass: store.AssetClassEquity, PrimarySource: store.SourcePolygon, Currency: "usd", Active: false, CompositeFIGI: &deadFIGI},
		{Symbol: "OSTK", Name: "Overstock.com Inc.", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: false, CompositeFIGI: &ostkFIGI},
		{Symbol: "BBBY", Name: "Beyond Inc.", AssetClass: store.AssetClassEquity, PrimarySource: store.SourceAlpaca, Currency: "usd", Active: true, CompositeFIGI: &bbbyFIGI},
	}

	ids := make(map[string]int64, len(rows))
	for _, r := range rows {
		id, err := ts.Upsert(ctx, r)
		require.NoError(t, err, "fixture upsert %s", r.Symbol)
		ids[r.Symbol] = id
	}

	require.NoError(t, ts.SetPinned(ctx, ids["SMOL"], true))

	_, err := ms.Create(ctx, store.TickerMigration{
		FromTickerID:  ids["OSTK"],
		ToTickerID:    ids["BBBY"],
		EffectiveDate: time.Date(2023, 8, 1, 0, 0, 0, 0, time.UTC),
		Reason:        store.MigrationReasonRebranding,
		Source:        store.MigrationSourceOpenFIGI,
	})
	require.NoError(t, err)

	return pool
}
