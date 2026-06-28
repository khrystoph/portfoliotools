# Track A2 — Ticker Universe Population: Design Spec

> **Read first:** `docs/superpowers/plans/2026-06-17-master-architecture.md`
> **Status:** Approved for implementation planning
> **Date:** 2026-06-28

---

## 1. Goal

Populate and maintain the `tickers` table with the full asset universe the platform will track: US equities and ETFs above a configurable market cap threshold, crypto, and a curated static list of commodities, indices, forex pairs, and bond yields. Run a nightly sentinel to catch intraweek renames and delistings, and a weekly full discovery sweep to catch new listings and threshold changes.

---

## 2. Asset Universe

### Dynamic (Polygon-discovered, market cap filtered)

| Asset Class | Source | Filter |
|---|---|---|
| US Equities | Polygon `/v3/reference/tickers` | market_cap ≥ `mcap_threshold_usd` (default 100M USD) |
| US ETFs | Polygon `/v3/reference/tickers` | market_cap ≥ `mcap_threshold_usd` |
| Crypto (top by mcap/volume) | Polygon `/v3/reference/tickers` | market = crypto |

### Static (seeded from Go constants, Yahoo Finance as OHLCV source)

**Commodities**

| Name | Yahoo Symbol |
|---|---|
| Gold | GC=F |
| Silver | SI=F |
| Crude Oil (WTI) | CL=F |
| Natural Gas | NG=F |
| Wheat | ZW=F |
| Corn | ZC=F |
| Soybeans | ZS=F |
| Copper | HG=F |
| Aluminum | ALI=F |

> **Note:** Steel has no reliable single futures ticker on Yahoo Finance. SLX (VanEck Steel ETF) covers the exposure and will enter via the normal Polygon equity/ETF discovery path. Verify ALI=F availability at implementation time — LME Aluminum futures coverage on Yahoo Finance varies.

**Major Indices**

| Name | Yahoo Symbol |
|---|---|
| S&P 500 | ^GSPC |
| NASDAQ Composite | ^IXIC |
| NASDAQ 100 | ^NDX |
| Dow Jones Industrial Average | ^DJI |
| Russell 2000 | ^RUT |
| VIX | ^VIX |
| KOSPI | ^KS11 |
| FTSE 100 | ^FTSE |
| DAX | ^GDAXI |
| Nikkei 225 | ^N225 |
| Hang Seng | ^HSI |

**Bond Yields**

| Name | Yahoo Symbol |
|---|---|
| 13-week T-Bill | ^IRX |
| 5-year Treasury | ^FVX |
| 10-year Treasury | ^TNX |
| 30-year Treasury | ^TYX |

> **Note:** The 2-year Treasury yield Yahoo ticker needs verification at implementation time (^TWO or similar). The 10-2 spread is a derived signal calculated in Track B from individual yield tickers — it is not a directly fetchable instrument.

**Forex (top 20 pairs by volume)**

| Pair | Yahoo Symbol |
|---|---|
| EUR/USD | EURUSD=X |
| USD/JPY | JPY=X |
| GBP/USD | GBPUSD=X |
| USD/CHF | CHF=X |
| AUD/USD | AUDUSD=X |
| USD/CAD | CAD=X |
| NZD/USD | NZDUSD=X |
| EUR/GBP | EURGBP=X |
| EUR/JPY | EURJPY=X |
| EUR/CHF | EURCHF=X |
| USD/HKD | HKD=X |
| USD/SGD | SGD=X |
| USD/SEK | SEK=X |
| USD/NOK | NOK=X |
| USD/MXN | MXN=X |
| USD/CNY | CNY=X |
| USD/INR | INR=X |
| USD/ZAR: | ZAR=X |
| USD/BRL | BRL=X |
| USD/TRY | TRY=X |

Static assets are never deactivated by the sync job. They are managed manually or via the admin UI.

---

## 3. Data Model

Three new migrations on top of A1's schema.

### Migration 007 — Extend `tickers`

```sql
ALTER TABLE tickers
  ADD COLUMN composite_figi    VARCHAR(20) UNIQUE,
  ADD COLUMN share_class_figi  VARCHAR(20),
  ADD COLUMN is_pinned         BOOLEAN NOT NULL DEFAULT FALSE;

CREATE INDEX idx_tickers_figi ON tickers(composite_figi)
  WHERE composite_figi IS NOT NULL;
```

`composite_figi` is nullable — static Yahoo-only assets have no FIGI. `is_pinned = true` blocks mcap-threshold deactivation but does NOT block confirmed-delisting deactivation.

### Migration 008 — `ticker_migrations` table

```sql
CREATE TABLE ticker_migrations (
  id             BIGSERIAL    PRIMARY KEY,
  from_ticker_id BIGINT       NOT NULL REFERENCES tickers(id),
  to_ticker_id   BIGINT       NOT NULL REFERENCES tickers(id),
  effective_date DATE         NOT NULL,
  reason         VARCHAR(50)  NOT NULL,
  source         VARCHAR(20)  NOT NULL,
  notes          TEXT,
  detected_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  created_at     TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_ticker_migrations_from ON ticker_migrations(from_ticker_id);
CREATE INDEX idx_ticker_migrations_to   ON ticker_migrations(to_ticker_id);
```

`reason` values: `rename`, `rebranding`, `acquisition`, `reverse_merger`
`source` values: `openfigi`, `polygon_events`, `manual`

### Migration 009 — `system_config` table

```sql
CREATE TABLE system_config (
  key         VARCHAR(100) PRIMARY KEY,
  value       TEXT         NOT NULL,
  description TEXT,
  updated_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

INSERT INTO system_config (key, value, description) VALUES
  ('mcap_threshold_usd', '100000000',
   'Minimum market cap in USD for automatic ticker inclusion. Pinned tickers are exempt from this threshold.');
```

The sync binary reads `mcap_threshold_usd` at runtime. Env var `SYNC_MCAP_THRESHOLD` provides a bootstrap fallback for the first run before the DB row exists; DB value takes precedence thereafter. The admin UI (Track C) updates this table directly.

---

## 4. Package Structure

```
migrations/
  000007_extend_tickers_figi_pinned.{up,down}.sql
  000008_create_ticker_migrations.{up,down}.sql
  000009_create_system_config.{up,down}.sql

internal/
  store/
    types.go          ← add TickerMigration type; extend Ticker with FIGI + IsPinned fields
    ticker.go         ← add GetByFIGI, ListWithFIGI, SetPinned, DeactivateByID
    ticker_test.go    ← tests for new methods
    migration.go      ← TickerMigrationStore: Create, GetByFromTicker, GetByToTicker
    migration_test.go
    config.go         ← SystemConfigStore: Get, Set, GetInt64
    config_test.go

  universe/
    universe.go       ← Syncer struct: Daily(), Weekly()
    polygon.go        ← PolygonClient: FetchActiveAssets, FetchTickerDetails
    openfigi.go       ← OpenFIGIClient: MapFIGIs; DetectChanges() pure function
    static.go         ← StaticAsset type; Commodities, Indices, ForexPairs, BondYields slices
    config.go         ← SyncConfig: load threshold from SystemConfigStore (env fallback)
    universe_test.go
    polygon_test.go
    openfigi_test.go

  testutil/
    db.go             ← existing NewTestDB()
    fixtures.go       ← NewTestDBWithFixtures() — schema + representative universe

cmd/
  syncTickers/
    main.go           ← --mode daily|weekly; wires config + Syncer

testdata/
  openfigi/
    map_figis_success.json
    map_figis_rate_limit.json
  polygon/
    reference_tickers_p1.json
    reference_tickers_p2.json
    ticker_details_found.json
    ticker_details_404.json
  alpaca/
    assets_active.json
```

`internal/universe` avoids shadowing the stdlib `sync` package.

---

## 5. Interfaces

```go
// universe/universe.go
type PolygonDiscoverer interface {
    FetchActiveAssets(ctx context.Context, mcapThreshold int64) ([]PolygonAsset, error)
    FetchTickerDetails(ctx context.Context, symbol string) (PolygonAsset, error)
}

type FIGIResolver interface {
    MapFIGIs(ctx context.Context, figis []string) ([]FIGIResult, error)
}

type AlpacaAssetLister interface {
    ListActiveSymbols(ctx context.Context) (map[string]struct{}, error)
}
```

Concrete implementations (`PolygonClient`, `OpenFIGIClient`) satisfy these interfaces. Mocks satisfy them in orchestration unit tests. This is the only place interfaces are used — the store layer does not get interfaces.

---

## 6. Sync Flows

### Daily (OpenFIGI sentinel)

1. Load `mcap_threshold_usd` from `system_config`
2. Fetch all tickers `WHERE composite_figi IS NOT NULL` from DB
3. Batch FIGIs into groups of 100 → call OpenFIGI `/v3/mapping`
4. Run `DetectChanges(known, results)` → renames + delisting candidates
5. **Renames:** call `PolygonClient.FetchTickerDetails(newSymbol)` → upsert new ticker, create `ticker_migrations` record (`source=openfigi`), mark old ticker `active=false`
6. **Delisting candidates:** call `PolygonClient.FetchTickerDetails(symbol)` to confirm → if confirmed inactive: `active=false`; if ambiguous (Polygon unreachable or still active): log warning, skip
7. Static assets: skip entirely (no FIGI, never swept)
8. Log summary

**Safety rule:** deactivation requires confirmation from two independent signals (OpenFIGI missing + Polygon inactive). One signal alone never deactivates.

**`is_pinned` rule:** `is_pinned=true` exempts a ticker from mcap-threshold deactivation. It does NOT exempt from confirmed-delisting deactivation — a confirmed dead ticker goes inactive regardless of pin.

### Weekly (Polygon full discovery)

1. Load `mcap_threshold_usd` from `system_config`
2. Fetch full Alpaca active asset list once → build `map[symbol]struct{}` for `primary_source` resolution
3. Paginate Polygon `/v3/reference/tickers` with `market_cap >= threshold`
4. For each Polygon asset:
   - **Not in DB (by FIGI):** batch OpenFIGI lookup to get FIGI, resolve `primary_source`, upsert
   - **In DB, same symbol, above threshold:** update name/exchange metadata, ensure `active=true`
   - **In DB, below threshold, `is_pinned=false`:** `active=false`
   - **In DB, below threshold, `is_pinned=true`:** skip, log
5. Any ticker in DB (`active=true`, not pinned, `primary_source != 'yahoo'`) absent from full Polygon result set → delisting candidate → same two-signal confirmation as daily flow
6. Seed/update static assets — upsert all, `primary_source=yahoo`, never deactivate
7. Run `Daily()` as sub-step to catch intraweek renames/delistings

### `primary_source` resolution

```
symbol in Alpaca active set? → primary_source = "alpaca"
else                         → primary_source = "polygon"
```

Static assets always seed with `primary_source = "yahoo"`. This field is the only thing the A4 OHLCV backfill needs to decide where to fetch data — no per-fetch logic.

### `DetectChanges` (pure function)

```go
func DetectChanges(known []store.Ticker, results []FIGIResult) (renames []FIGIChange, delistings []store.Ticker)
```

Input: known tickers with FIGIs, OpenFIGI mapping results.
Output: renames (FIGI now under different symbol), delisting candidates (FIGI absent from results).
No DB access. No HTTP calls. Fully deterministic. Separately unit-tested.

---

## 7. Error Handling and Retry Strategy

### Two-tier model

**Tier 1 — per-call local retry:** 3 attempts with 1s/2s/4s backoff for transient network errors. If exhausted, escalates to job-level.

**Tier 2 — job-level exponential backoff with deadline:**

```
Initial delay:  1 minute
Multiplier:     2×
Cap:            30 minutes
Schedule:       1m → 2m → 4m → 8m → 16m → 30m → 30m → ...

Deadlines:
  Daily  → next market open (9:30 AM ET, next trading day via Alpaca calendar)
  Weekly → Sunday 17:00 ET (CME Globex / futures market open)
```

`context.WithDeadline()` carries the deadline into all downstream API calls. When deadline is reached, in-flight calls cancel, job exits with logged timeout.

All sync operations are idempotent (upserts, `ON CONFLICT`). Retry from the beginning of a partially-completed run is always safe.

k8s CronJob uses `backoffLimit: 0` — the application owns its retry lifecycle.

### Failure table

| Failure | Behaviour |
|---|---|
| Any API unreachable | Tier 1 local retry, then tier 2 exponential backoff to deadline |
| Deadline reached | Log timeout with step context and retry count; alert |
| Partial OpenFIGI batch failure | Process successful batches; failed FIGIs retry next job cycle |
| OpenFIGI missing FIGI, Polygon also unreachable | No deactivations; both signals required |
| Polygon cross-check ambiguous (delisting) | Skip deactivation; log warning; retry next cycle |
| Rename + delisting same week | Rename processed first; missing new symbol caught next daily cycle |
| `is_pinned` below threshold | Preserved; logged for admin review |
| `is_pinned` confirmed delisted | Deactivated; pin does not override confirmed death |

---

## 8. Testing

### Layer 1 — Pure unit tests

`DetectChanges()` exhaustively:
- FIGI present, same symbol → no change
- FIGI present, different symbol → rename
- FIGI absent → delisting candidate
- Mixed batch covering all three
- Empty inputs → no-op

Also: `primary_source` resolution, backoff deadline calculation, static asset sanity (no duplicate symbols, all fields populated).

Retry/deadline tests use injected backoff config (1ms initial delay, 5ms deadline) — no real sleeps.

### Layer 2 — HTTP client tests (`httptest.NewServer`)

Recorded responses in `testdata/` (real API shapes, stable, no live network in CI). Covers: pagination, market cap filter parsing, 404 handling, rate limit response triggering local retry, partial batch parsing.

### Layer 3 — Orchestration unit tests (interface mocks)

`Syncer.Daily()` and `Syncer.Weekly()` with mock `PolygonDiscoverer`, `FIGIResolver`, `AlpacaAssetLister`.

Key `Daily()` cases:
- Rename → migration record created, old inactive, new active
- Polygon says still active → no deactivation (ambiguous rule enforced)
- OpenFIGI unreachable → entire FIGI check skipped, zero deactivations
- `is_pinned=true`, below threshold → preserved
- `is_pinned=true`, confirmed delisted → deactivated

Key `Weekly()` cases:
- Net-new asset → upserted with FIGI and correct `primary_source`
- Existing, below threshold, not pinned → deactivated
- Existing, below threshold, pinned → preserved, logged
- Absent from Polygon result set → delisting candidate → confirmed → deactivated
- Static assets → always upserted, never deactivated

### Layer 4 — Integration tests (real DB + httptest + fixtures)

`NewTestDBWithFixtures()` seeds:

| Symbol | Class | Source | Status | FIGI | Notes |
|---|---|---|---|---|---|
| AAPL | equity | alpaca | active | yes | above threshold |
| SPY | etf | alpaca | active | yes | above threshold |
| BTC | crypto | alpaca | active | yes | above threshold |
| GC=F | commodity | yahoo | active | no | static; never swept |
| ^GSPC | index | yahoo | active | no | static; never swept |
| EURUSD=X | forex | yahoo | active | no | static; never swept |
| ^TNX | bond | yahoo | active | no | static; never swept |
| TINY | equity | alpaca | active | yes | BELOW threshold, not pinned → deactivated by weekly sync |
| SMOL | equity | alpaca | active | yes | BELOW threshold, `is_pinned=true` → preserved |
| DEAD | equity | polygon | inactive | yes | confirmed delisted, no successor |
| OSTK | equity | alpaca | inactive | yes | migration → BBBY (Beyond Inc.) |
| BBBY | equity | alpaca | active | yes | successor of OSTK |

Integration tests verify full flow from httptest API response through to DB state: migration records created correctly, active flags flipped as expected, pinned tickers preserved, static assets untouched.

---

## 9. Configuration

| Setting | Source | Default | Notes |
|---|---|---|---|
| `mcap_threshold_usd` | `system_config` table | 100,000,000 | Admin UI writable; env var `SYNC_MCAP_THRESHOLD` as bootstrap fallback |
| Pinned tickers | `tickers.is_pinned` column | false | Admin UI writable per-ticker |
| Polygon API key | `.stockclientconfig.json` / k8s Secret | — | Existing credential pattern |
| Alpaca API key/secret | `.stockclientconfig.json` / k8s Secret | — | Existing credential pattern |
| OpenFIGI API key | k8s Secret | — | Optional; raises rate limit from 25 to 250 req/min |

---

## 10. k8s Deployment

Two CronJobs, same image, different `--mode` flag:

```yaml
# Daily sentinel — runs at 01:00 UTC (after market close, before backfill)
schedule: "0 1 * * *"
args: ["--mode", "daily"]

# Weekly discovery — runs at 06:00 UTC Saturday (after Friday close)
schedule: "0 6 * * 6"
args: ["--mode", "weekly"]
```

Both: `backoffLimit: 0` (app handles retry), resource requests TBD at Track C infra planning.

---

## 11. Out of Scope (deferred to later tracks)

- OHLCV data fetching (Track A3)
- Yahoo Finance HTTP client implementation (Track A3)
- Backfill orchestration (Track A4)
- 10-2 spread signal calculation (Track B)
- Admin UI for `is_pinned` and `mcap_threshold_usd` (Track C)
- Alpaca corporate actions as secondary migration signal (can be added if Polygon event coverage proves insufficient)

---

## 12. Key Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Ticker discovery source | Polygon `/v3/reference/tickers` | Only source with bulk active-ticker list + market cap in one call |
| Identity and migration detection | OpenFIGI `composite_figi` | Persistent through symbol changes; industry standard; free API |
| Daily change detection | OpenFIGI daily FIGI batch check | Catches intraweek renames/delistings before nightly OHLCV backfill |
| Targeted vs full Polygon on rename | Targeted single-ticker lookup | 1-2 API calls per detected change vs full discovery run |
| Deactivation safety rule | Two-signal confirmation required | OpenFIGI alone insufficient; prevents false delistings on API gaps |
| `is_pinned` scope | Blocks mcap deactivation only | Confirmed dead tickers must go inactive regardless of pin |
| mcap threshold storage | `system_config` DB table | Admin UI writable; no ConfigMap change required |
| Static assets | Go constants in `static.go` | Small list, rarely changes, visible in git history, no runtime file dep |
| Retry strategy | App-level exponential backoff with market-deadline | Respects trading day boundaries; doesn't hammer down APIs |
| Interface mocking | Only for HTTP API clients in orchestration tests | DB divergence risk absent for narrow, documented external APIs |
