# Trend Tracking Feature Design

**Date:** 2026-06-01
**Status:** Approved

## Overview

Implements multi-duration simple slope calculation and directional trend determination (Bullish / Bearish / Neutral / Indeterminate) across three timeframes — Trade (30-day), Trend (90-day), and Tail (180-day) — inspired by Hedgeye's Price, Volume, Volatility method.

Results are stored per-day in the full data structure and surfaced in both the JSON output file and the Excel report.

## Background and Naming Note

The existing codebase has a field-naming inconsistency: struct fields are named with `30/60/90` suffixes but the actual constants are `SHORTDURATION=30`, `MEDIUMDURATION=90`, `LONGDURATION=180`. The "60-day" fields actually hold 90-day data; the "90-day" fields hold 180-day data. This is pre-existing debt, not introduced by this feature. A separate refactor pass will clean up those names.

## Architecture

Two new functions following the existing one-metric-per-function pattern:

1. `GetSimpleSlopes` — computes raw price delta over each duration window per day
2. `CalculateTrendDirections` — reads stored slopes and labels each day's direction

Both functions are called in `batchStocks.go` after the existing pipeline, before output generation.

## Data Structures

### `SingleStockCandle` (internal, full data)

Add direction fields:
```go
TradeDirection string `json:"trade-direction"`
TrendDirection string `json:"trend-direction"`
TailDirection  string `json:"tail-direction"`
```

Add internal validity flags (omitted from JSON output):
```go
SlopeShortValid bool `json:"-"`
SlopeMedValid   bool `json:"-"`
SlopeLongValid  bool `json:"-"`
```

### `condensedStockCandle` (internal, condensed view)

Add direction fields (slopes already present as `TradeSlope`, `TrendSlope`, `TailSlope`):
```go
TradeDirection string `json:"trade-direction"`
TrendDirection string `json:"trend-direction"`
TailDirection  string `json:"tail-direction"`
```

### `CondensedRangesJSON` (JSON and Excel output)

Replace placeholder `Slope float64` and `Trend string` with:
```go
TradeSlope     float64 `json:"trade-slope"`
TrendSlope     float64 `json:"trend-slope"`
TailSlope      float64 `json:"tail-slope"`
TradeDirection string  `json:"trade-direction"`
TrendDirection string  `json:"trend-direction"`
TailDirection  string  `json:"tail-direction"`
```

## `GetSimpleSlopes` (rewrite of broken stub)

**Input:** `map[string]map[int64]SingleStockCandle`, `isDebug bool`
**Output:** same map with slopes and validity flags populated

**Algorithm per ticker:**
1. Collect all date keys, sort descending (most recent first)
2. For each day at index `i`:
   - Compute target lookback timestamp: `today - N calendar days` as UnixMilli
   - Walk backward through sorted keys until finding a key `<= target`
   - If found: `slope = close_today - close_at_lookback_date`, set valid flag = `true`
   - If not found (predates available history): slope stays `0.0`, valid flag stays `false`
3. Write `SlopeShortDuration` / `SlopeShortValid`, `SlopeMedDuration` / `SlopeMedValid`, `SlopeLongDuration` / `SlopeLongValid`

**Slope value:** raw price delta (dollars), not normalized per day. Sign drives direction; magnitude reserved for future scoring.

**Lookback direction:** nearest trading day **at or before** the target calendar date (never forward-looking).

## `CalculateTrendDirections`

**Input:** `map[string]map[int64]SingleStockCandle`
**Output:** same map with direction fields populated

**Algorithm per ticker:**
1. Collect all date keys, sort **ascending** (oldest first) so prior slopes are available when needed
2. For each day at index `i`, for each duration (short/med/long):
   - If `i < 2`: direction = `"Indeterminate"` (fewer than 3 trading days in dataset)
   - Else collect slopes from `i`, `i-1`, `i-2`:
     - If any of the three has its valid flag = `false`: direction = `"Indeterminate"`
     - Else if all three slopes `> 0`: direction = `"Bullish"`
     - Else if all three slopes `< 0`: direction = `"Bearish"`
     - Else: direction = `"Neutral"`
     - Note: slope of exactly `0.0` with valid = `true` is treated as neither positive nor negative → contributes to Neutral
3. Write `TradeDirection`, `TrendDirection`, `TailDirection` onto each candle

## Pipeline Wiring (`batchStocks.go`)

Add after existing calculation chain, before the condensing loop:
```go
tickerData = pkg.GetSimpleSlopes(tickerData, debug)
tickerData = pkg.CalculateTrendDirections(tickerData)
```

Populate `CondensedRangesJSON` using latest day's slope and direction fields (replacing the hardcoded `Slope: 0.0` and `Trend: "Not Yet Implemented"`).

Add `-tail` / `--tail-cols` bool flag to the `init()` flag block. Pass it to `GenerateStockReportXLSX`.

## `PrepareToPrintData` (`TechAnalysis.go`)

Update the existing function to map the three new direction fields from `SingleStockCandle` into `condensedStockCandle`:
```go
TradeDirection: stockPrices[ticker][dateInt64].TradeDirection,
TrendDirection: stockPrices[ticker][dateInt64].TrendDirection,
TailDirection:  stockPrices[ticker][dateInt64].TailDirection,
```

## Excel Report (`Excelizer.go`)

### Default columns (11 — same count as before)

| # | Column | Notes |
|---|--------|-------|
| 1 | Ticker | |
| 2 | Close | |
| 3 | Avg Vol Ratio | |
| 4 | RVol % | |
| 5 | VAPR Low | |
| 6 | VAPR High | |
| 7 | Trade Slope | |
| 8 | Trend Slope | |
| 9 | Trade Dir | Color-coded |
| 10 | Trend Dir | Color-coded |
| 11 | Timestamp | |

Volume and raw RVol columns removed. Single Slope/Trend placeholder replaced with per-duration equivalents.

### Tail columns (optional, 13 columns)

Enabled via `-tail` / `--tail-cols` CLI flag on `batchStocks`. Adds Tail Slope and Tail Dir after Trend Dir. The `GenerateStockReportXLSX` function signature gains a `showTail bool` parameter to control this.

### Direction cell coloring

| Value | Style |
|-------|-------|
| Bullish | Green solid fill |
| Bearish | Red solid fill |
| Neutral | Grey solid fill |
| Indeterminate | Red/white diagonal stripes (excelize pattern 7) |

Applied to Trade Dir and Trend Dir cells (and Tail Dir when `-tail` enabled).

## Testing

All new functions get table-driven unit tests in `TechAnalysis_test.go` following existing patterns.

### `GetSimpleSlopes` test cases
- Correct raw delta when exact lookback date exists
- Rollback to nearest prior trading day when exact date is missing
- Valid flag set to `true` when slope is computed
- Valid flag stays `false` and slope stays `0.0` when history is insufficient

### `CalculateTrendDirections` test cases
- All three slopes positive → `"Bullish"`
- All three slopes negative → `"Bearish"`
- Mixed slopes → `"Neutral"`
- Slope exactly `0.0` with valid = `true` → `"Neutral"` (not Indeterminate)
- Fewer than 3 days in dataset → `"Indeterminate"`
- Any of the three slopes has valid flag = `false` → `"Indeterminate"`

### Coverage goals (project-wide directive)
- Unit tests: **90%** target
- Integration tests: **70%** target

New code ships with tests. Bringing existing code up to these targets is a separate backlog item.
