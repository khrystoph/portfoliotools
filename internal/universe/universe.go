package universe

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/khrystoph/portfoliotools/internal/store"
)

// Syncer orchestrates the daily and weekly ticker universe sync.
type Syncer struct {
	tickers    *store.TickerStore
	migrations *store.TickerMigrationStore
	sysconfig  *store.SystemConfigStore
	polygon    PolygonDiscoverer
	figi       FIGIResolver
	alpaca     AlpacaAssetLister
	cfg        SyncConfig
}

func NewSyncer(
	tickers *store.TickerStore,
	migrations *store.TickerMigrationStore,
	sysconfig *store.SystemConfigStore,
	polygon PolygonDiscoverer,
	figi FIGIResolver,
	alpaca AlpacaAssetLister,
	cfg SyncConfig,
) *Syncer {
	return &Syncer{
		tickers: tickers, migrations: migrations, sysconfig: sysconfig,
		polygon: polygon, figi: figi, alpaca: alpaca, cfg: cfg,
	}
}

// Daily runs the OpenFIGI sentinel to catch intraweek renames and delistings.
// Retries with exponential backoff until the next market open (9:30 AM ET).
func (s *Syncer) Daily(ctx context.Context) error {
	deadline := nextMarketOpen()
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	delay := s.cfg.BackoffConfig.InitialDelay
	for attempt := 0; ; attempt++ {
		err := s.runDaily(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return fmt.Errorf("daily sync timed out after %d attempt(s): %w", attempt+1, ctx.Err())
		}
		log.Printf("daily sync attempt %d failed: %v; retrying in %s", attempt+1, err, delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("daily sync timed out after %d attempt(s): %w", attempt+1, ctx.Err())
		}
		delay = capDuration(time.Duration(float64(delay)*s.cfg.BackoffConfig.Multiplier), s.cfg.BackoffConfig.Cap)
	}
}

// Weekly runs the full Polygon discovery sweep, threshold filtering, static asset seeding,
// then calls runDaily to catch any intraweek changes.
// Retries with exponential backoff until Sunday 17:00 ET (CME Globex open).
func (s *Syncer) Weekly(ctx context.Context) error {
	deadline := nextWeeklyDeadline()
	ctx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	delay := s.cfg.BackoffConfig.InitialDelay
	for attempt := 0; ; attempt++ {
		err := s.runWeekly(ctx)
		if err == nil {
			return nil
		}
		if ctx.Err() != nil {
			return fmt.Errorf("weekly sync timed out after %d attempt(s): %w", attempt+1, ctx.Err())
		}
		log.Printf("weekly sync attempt %d failed: %v; retrying in %s", attempt+1, err, delay)
		select {
		case <-time.After(delay):
		case <-ctx.Done():
			return fmt.Errorf("weekly sync timed out after %d attempt(s): %w", attempt+1, ctx.Err())
		}
		delay = capDuration(time.Duration(float64(delay)*s.cfg.BackoffConfig.Multiplier), s.cfg.BackoffConfig.Cap)
	}
}

func (s *Syncer) runDaily(ctx context.Context) error {
	known, err := s.tickers.ListWithFIGI(ctx)
	if err != nil {
		return fmt.Errorf("list tickers with FIGI: %w", err)
	}

	const batchSize = 100
	var results []FIGIResult
	for i := 0; i < len(known); i += batchSize {
		end := i + batchSize
		if end > len(known) {
			end = len(known)
		}
		batch := known[i:end]
		figis := make([]string, len(batch))
		for j, tk := range batch {
			figis[j] = *tk.CompositeFIGI
		}
		batchResults, err := s.figi.MapFIGIs(ctx, figis)
		if err != nil {
			return fmt.Errorf("openfigi batch %d: %w", i/batchSize, err)
		}
		results = append(results, batchResults...)
	}

	renames, delistings := DetectChanges(known, results)

	for _, change := range renames {
		if err := s.processRename(ctx, change); err != nil {
			log.Printf("rename %s→%s: %v", change.OldSymbol, change.NewSymbol, err)
		}
	}
	for _, candidate := range delistings {
		if err := s.confirmAndDeactivate(ctx, candidate); err != nil {
			log.Printf("delisting check %s: %v", candidate.Symbol, err)
		}
	}

	log.Printf("daily sync complete: %d renames, %d delisting candidates processed", len(renames), len(delistings))
	return nil
}

func (s *Syncer) runWeekly(ctx context.Context) error {
	alpacaSymbols, err := s.alpaca.ListActiveSymbols(ctx)
	if err != nil {
		return fmt.Errorf("list alpaca symbols: %w", err)
	}

	polygonAssets, err := s.polygon.FetchActiveAssets(ctx, s.cfg.McapThresholdUSD)
	if err != nil {
		return fmt.Errorf("fetch polygon assets: %w", err)
	}

	polygonByFIGI := make(map[string]PolygonAsset, len(polygonAssets))
	for _, a := range polygonAssets {
		if a.CompositeFIGI != "" {
			polygonByFIGI[a.CompositeFIGI] = a
		}
	}

	for _, a := range polygonAssets {
		src := store.SourcePolygon
		if _, ok := alpacaSymbols[a.Ticker]; ok {
			src = store.SourceAlpaca
		}
		var cf, sf *string
		if a.CompositeFIGI != "" {
			cf = &a.CompositeFIGI
		}
		if a.ShareClassFIGI != "" {
			sf = &a.ShareClassFIGI
		}
		if _, err := s.tickers.Upsert(ctx, store.Ticker{
			Symbol:         a.Ticker,
			Name:           a.Name,
			AssetClass:     resolveAssetClass(a),
			PrimarySource:  src,
			Currency:       a.CurrencyName,
			Active:         true,
			CompositeFIGI:  cf,
			ShareClassFIGI: sf,
		}); err != nil {
			log.Printf("upsert %s: %v", a.Ticker, err)
		}
	}

	// Deactivate tickers absent from Polygon with two-signal confirmation
	active, err := s.tickers.ListActive(ctx, "")
	if err != nil {
		return fmt.Errorf("list active tickers: %w", err)
	}
	for _, tk := range active {
		if tk.PrimarySource == store.SourceYahoo {
			continue // static assets are never swept
		}
		if tk.CompositeFIGI == nil {
			continue
		}
		if _, inPolygon := polygonByFIGI[*tk.CompositeFIGI]; inPolygon {
			continue
		}
		if tk.IsPinned {
			log.Printf("ticker %s absent from Polygon but is_pinned; preserving", tk.Symbol)
			continue
		}
		if err := s.confirmAndDeactivate(ctx, tk); err != nil {
			log.Printf("deactivation check %s: %v", tk.Symbol, err)
		}
	}

	// Seed static assets — always upsert, never deactivate
	for _, sa := range AllStaticAssets() {
		if _, err := s.tickers.Upsert(ctx, store.Ticker{
			Symbol:        sa.Symbol,
			Name:          sa.Name,
			AssetClass:    sa.Class,
			PrimarySource: store.SourceYahoo,
			Currency:      "usd",
			Active:        true,
		}); err != nil {
			log.Printf("seed static %s: %v", sa.Symbol, err)
		}
	}

	return s.runDaily(ctx)
}

func (s *Syncer) processRename(ctx context.Context, change FIGIChange) error {
	details, err := s.polygon.FetchTickerDetails(ctx, change.NewSymbol)
	if err != nil {
		return fmt.Errorf("fetch details for %s: %w", change.NewSymbol, err)
	}

	old, err := s.tickers.GetByFIGI(ctx, change.CompositeFIGI)
	if err != nil {
		return fmt.Errorf("get old ticker by FIGI: %w", err)
	}

	var cf, sf *string
	if details.CompositeFIGI != "" {
		cf = &details.CompositeFIGI
	}
	if details.ShareClassFIGI != "" {
		sf = &details.ShareClassFIGI
	}
	newID, err := s.tickers.Upsert(ctx, store.Ticker{
		Symbol:         details.Ticker,
		Name:           details.Name,
		AssetClass:     resolveAssetClass(details),
		PrimarySource:  store.SourcePolygon,
		Currency:       details.CurrencyName,
		Active:         true,
		CompositeFIGI:  cf,
		ShareClassFIGI: sf,
	})
	if err != nil {
		return fmt.Errorf("upsert renamed ticker: %w", err)
	}

	_, err = s.migrations.Create(ctx, store.TickerMigration{
		FromTickerID:  old.ID,
		ToTickerID:    newID,
		EffectiveDate: time.Now().UTC(),
		Reason:        store.MigrationReasonRename,
		Source:        store.MigrationSourceOpenFIGI,
	})
	if err != nil {
		return fmt.Errorf("create migration record: %w", err)
	}

	return s.tickers.DeactivateByID(ctx, old.ID)
}

func (s *Syncer) confirmAndDeactivate(ctx context.Context, candidate store.Ticker) error {
	if candidate.IsPinned {
		log.Printf("delisting skipped for %s: is_pinned=true", candidate.Symbol)
		return nil
	}
	details, err := s.polygon.FetchTickerDetails(ctx, candidate.Symbol)
	if err != nil {
		// Polygon unreachable or 404 — cannot confirm with two signals; skip
		log.Printf("delisting confirmation skipped for %s: %v", candidate.Symbol, err)
		return nil
	}
	if details.Active {
		log.Printf("delisting skipped for %s: Polygon reports still active", candidate.Symbol)
		return nil
	}
	return s.tickers.DeactivateByID(ctx, candidate.ID)
}

func resolveAssetClass(a PolygonAsset) store.AssetClass {
	switch a.Market {
	case "crypto":
		return store.AssetClassCrypto
	default:
		if a.Type == "ETF" || a.Type == "ETV" {
			return store.AssetClassETF
		}
		return store.AssetClassEquity
	}
}

func nextMarketOpen() time.Time {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	next := time.Date(now.Year(), now.Month(), now.Day(), 9, 30, 0, 0, loc)
	if now.Before(next) {
		return next
	}
	return next.Add(24 * time.Hour)
}

func nextWeeklyDeadline() time.Time {
	loc, _ := time.LoadLocation("America/New_York")
	now := time.Now().In(loc)
	daysUntilSunday := (7 - int(now.Weekday())) % 7
	if daysUntilSunday == 0 {
		target := time.Date(now.Year(), now.Month(), now.Day(), 17, 0, 0, 0, loc)
		if now.Before(target) {
			return target
		}
		daysUntilSunday = 7
	}
	d := now.AddDate(0, 0, daysUntilSunday)
	return time.Date(d.Year(), d.Month(), d.Day(), 17, 0, 0, 0, loc)
}

func capDuration(d, max time.Duration) time.Duration {
	if d > max {
		return max
	}
	return d
}
