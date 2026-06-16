# Duration Refactor Design — Issues #5 + #6

**Date:** 2026-06-11
**Issues:** #5 (duration-parameterized functions), #6 (field naming cleanup)
**Status:** Approved, ready for implementation planning

---

## Overview

Two tech debt items combined into one pass:

- **#5** — Each function in `TechAnalysis.go` contains three near-identical code blocks, one per duration (SHORT/MEDIUM/LONG). Extract a single parameterized function body; callers call it three times.
- **#6** — `SingleStockCandle` struct fields use numeric suffixes `30/60/90` that don't match the actual duration constants (`SHORTDURATION=30`, `MEDIUMDURATION=90`, `LONGDURATION=180`), causing silent confusion. Rename to `Short/Med/Long`.

---

## Section 1: Constants

The existing duration constants are **unchanged**:

```go
const SHORTDURATION  = 30
const MEDIUMDURATION = 90
const LONGDURATION   = 180
```

These are `const` intentionally — they define what "short", "medium", and "long" mean at build time. If the thresholds need to change (e.g. short becomes 20 days), one line changes and the binary is rebuilt. No runtime override is needed or desired for durations. What is runtime-configurable is the date range passed to the data fetch, not the duration buckets.

No new `Duration` type is introduced. All parameterized functions accept `duration int` and callers pass the existing constants.

---

## Section 2: Struct Field Renames

All `30/60/90` numeric suffixes on `SingleStockCandle` are renamed to `Short/Med/Long`. JSON tags follow the concept-first, duration-qualifier-last pattern (e.g. `realized-volatility-short`).

| Go field (before) | Go field (after) | JSON tag (after) |
|--------------------|-----------------|-----------------|
| `ThirtyDaysPrices` | `ShortPrices` | `short-prices` |
| `SixtyDaysPrices` | `MedPrices` | `med-prices` |
| `NinetyDaysPrices` | `LongPrices` | `long-prices` |
| `AvgVolume30` | `AvgVolumeShort` | `short-avg-volume` |
| `AvgVolume60` | `AvgVolumeMed` | `med-avg-volume` |
| `AvgVolume90` | `AvgVolumeLong` | `long-avg-volume` |
| `AvgVolumeRatio30` | `AvgVolumeRatioShort` | `short-avg-volume-ratio` |
| `AvgVolumeRatio60` | `AvgVolumeRatioMed` | `med-avg-volume-ratio` |
| `AvgVolumeRatio90` | `AvgVolumeRatioLong` | `long-avg-volume-ratio` |
| `RealizedVolatility30` | `RealizedVolatilityShort` | `short-realized-volatility` |
| `RealizedVolatility60` | `RealizedVolatilityMed` | `med-realized-volatility` |
| `RealizedVolatility90` | `RealizedVolatilityLong` | `long-realized-volatility` |
| `VelocityRealizedVol30` | `VelocityRealizedVolShort` | `short-rvol-velocity` |
| `VelocityRealizedVol60` | `VelocityRealizedVolMed` | `med-rvol-velocity` |
| `VelocityRealizedVol90` | `VelocityRealizedVolLong` | `long-rvol-velocity` |
| `RealizedVolAccel30` | `RealizedVolAccelShort` | `short-rvol-accel` |
| `RealizedVolAccel60` | `RealizedVolAccelMed` | `med-rvol-accel` |
| `RealizedVolAccel90` | `RealizedVolAccelLong` | `long-rvol-accel` |
| `RVolHigh30` | `RVolHighShort` | `short-rvol-high` |
| `RVolHigh60` | `RVolHighMed` | `med-rvol-high` |
| `RVolHigh90` | `RVolHighLong` | `long-rvol-high` |
| `RVolLow30` | `RVolLowShort` | `short-rvol-low` |
| `RVolLow60` | `RVolLowMed` | `med-rvol-low` |
| `RVolLow90` | `RVolLowLong` | `long-rvol-low` |
| `RVolPercent30` | `RVolPercentShort` | `short-rvol-percent` |
| `RVolPercent60` | `RVolPercentMed` | `med-rvol-percent` |
| `RVolPercent90` | `RVolPercentLong` | `long-rvol-percent` |

### Existing tags that stay unchanged

Fields using `trade-/trend-/tail-` prefixes (`trade-slope`, `trade-range`, `trend-range`, `tail-slope`, `prob-adj-trade-range`, etc.) are already correctly named and are not touched. The Hedgeye-derived trade/trend/tail terminology applies to the computed output fields (slopes, ranges, directions); the renamed fields above are the underlying metric inputs.

### condensedStockCandle

Go field names in `condensedStockCandle` already use `Short/Med/Long` and do not change. JSON tags in that struct that still use `short-/med-/long-` (e.g. `short-avg-volume`) already match the target convention and require no changes.

### Breaking change

This is a breaking change to the JSON wire format. Any downstream consumer of the JSON output (scripts, Excel imports, other tooling) must be updated to use the new tag names.

### Presentation layer (future)

The Excel/user-facing output layer will map `Short→Pulse`, `Med→Wave`, `Long→Tide` for human-readable column headers. This mapping lives exclusively in `pkg/Excelizer.go` and is out of scope for this PR. The terminology change is deliberately decoupled so it can be updated independently.

---

## Section 3: Accessor Helpers

A new file `pkg/duration_accessors.go` holds one unexported `get`/`set` pair per field family. Each pair contains the `duration int → struct field` mapping exactly once. This is the single source of truth for the duration-to-field dispatch — no other code should contain a raw `switch d { case SHORTDURATION: ... }` over these fields.

```go
func getRVol(c SingleStockCandle, d int) float64 {
    switch d {
    case SHORTDURATION:  return c.RealizedVolatilityShort
    case MEDIUMDURATION: return c.RealizedVolatilityMed
    case LONGDURATION:   return c.RealizedVolatilityLong
    }
    return 0
}

func setRVol(c *SingleStockCandle, d int, v float64) {
    switch d {
    case SHORTDURATION:  c.RealizedVolatilityShort = v
    case MEDIUMDURATION: c.RealizedVolatilityMed = v
    case LONGDURATION:   c.RealizedVolatilityLong = v
    }
}
```

Full set of helper pairs:

| Helper pair | Field family |
|-------------|-------------|
| `getPrices` / `setPrices` | `ShortPrices` / `MedPrices` / `LongPrices` |
| `getRVol` / `setRVol` | `RealizedVolatilityShort` / `Med` / `Long` |
| `getAvgVol` / `setAvgVol` | `AvgVolumeShort` / `Med` / `Long` |
| `getAvgVolRatio` / `setAvgVolRatio` | `AvgVolumeRatioShort` / `Med` / `Long` |
| `getRVolVel` / `setRVolVel` | `VelocityRealizedVolShort` / `Med` / `Long` |
| `getRVolAccel` / `setRVolAccel` | `RealizedVolAccelShort` / `Med` / `Long` |
| `getRVolHigh` / `setRVolHigh` | `RVolHighShort` / `Med` / `Long` |
| `getRVolLow` / `setRVolLow` | `RVolLowShort` / `Med` / `Long` |
| `getRVolPercent` / `setRVolPercent` | `RVolPercentShort` / `Med` / `Long` |
| `getRiskRange` / `setRiskRange` | `TradeRange` / `TrendRange` / `TailRange` |
| `getAdjRiskRange` / `setAdjRiskRange` | `TradeRangeAdj` / `TrendRangeAdj` / `TailRangeAdj` |

The risk range helpers dispatch to the existing `Trade/Trend/Tail` named fields: `SHORTDURATION → TradeRange`, `MEDIUMDURATION → TrendRange`, `LONGDURATION → TailRange`.

---

## Section 4: Function Signatures

All functions that currently process all three durations internally gain a `duration int` parameter and process only that duration per call.

```go
func StoreRealizedVols(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func GetAvgVolume(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func GetRelHighLowVol(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func CalculateRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func CalculateVolumeAdjustedRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func CalculateAvgVolumeRatios(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func CalculateVelocities(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func CalculateAccelerations(stockPrices map[string]map[int64]SingleStockCandle, duration int) map[string]map[int64]SingleStockCandle
func GetLinearRegressionSlope(stockPrices map[string]map[int64]SingleStockCandle, duration int, isDebug bool) map[string]map[int64]SingleStockCandle
func GetProbAdjRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int, probabilityAdjustment float64) map[string]map[int64]SingleStockCandle
```

`GetSimpleSlopes` and `CalculateTrendDirections` are **not** changed — they already operate cleanly across all durations in a single pass and their internal structure does not duplicate per-duration blocks.

---

## Section 5: Internal Structure

### Group 1 — Date-windowed: `StoreRealizedVols`, `GetAvgVolume`, `GetRelHighLowVol`

These three functions share identical boilerplate: collect date keys, sort descending, compute window start, validate window bounds, collect window dates. A private helper extracts this scaffolding:

```go
// collectWindowDates returns the slice of date keys within the duration window
// starting at reverseDateKeys[index], and a bool indicating whether the window
// was valid (enough history, within range). reverseDateKeys must be sorted descending.
func collectWindowDates(reverseDateKeys []int64, index int, duration int) ([]int64, bool) {
    durationStartMilli := time.UnixMilli(reverseDateKeys[index]).AddDate(0, 0, -1*duration).UnixMilli()
    if index+duration >= len(reverseDateKeys)-1 || reverseDateKeys[index] < durationStartMilli {
        return nil, false
    }
    var windowDates []int64
    for i := index; i < len(reverseDateKeys) && reverseDateKeys[i] >= durationStartMilli; i++ {
        windowDates = append(windowDates, reverseDateKeys[i])
    }
    return windowDates, true
}
```

Each of the three functions then: sorts dates once, iterates, calls `collectWindowDates`, passes the result to its specific computation, writes back via accessor helpers.

### Group 2 — Per-day: `CalculateRiskRanges`, `CalculateVolumeAdjustedRiskRanges`, `CalculateAvgVolumeRatios`, `GetProbAdjRiskRanges`, `GetLinearRegressionSlope`

These iterate `for day := range stockPrices[ticker]` with no date sorting. Each becomes a read-via-accessor / compute / write-via-accessor loop. No shared helper needed.

`GetLinearRegressionSlope` fits here because it reads from the stored price maps (`ShortPrices`, `MedPrices`, `LongPrices`) already attached to each candle — it does not compare to a previous day or build a new date window.

`GetProbAdjRiskRanges` processes both the plain and adjusted risk range fields for one duration per call, using `getRiskRange`/`setRiskRange` and `getAdjRiskRange`/`setAdjRiskRange` accessor helpers.

### Group 3 — Sequential: `CalculateVelocities`, `CalculateAccelerations`

These walk dates in sorted ascending order and compute the delta from the previous trading day. Accessor helpers handle field dispatch; loop structure is unchanged.

---

## Section 6: Caller Changes

Both `cmd/stockClient/stockClient.go` and `cmd/batchStocks/batchStocks.go` expand each single function call to a range loop over the three duration constants. A `durations` slice is defined once and reused:

```go
durations := []int{pkg.SHORTDURATION, pkg.MEDIUMDURATION, pkg.LONGDURATION}

for _, d := range durations {
    tickerData = pkg.StoreRealizedVols(tickerData, d)
}
for _, d := range durations {
    tickerData = pkg.GetRelHighLowVol(tickerData, d)
}
// ... repeated for all parameterized functions
```

Call order within each caller is preserved from the current implementation.

`GetProbAdjRiskRanges` call sites update `probabilityAdjustment` from second argument to third:
```go
// before
pkg.GetProbAdjRiskRanges(tickerData, stockDataConfig.RangeAdjustment)
// after
pkg.GetProbAdjRiskRanges(tickerData, d, stockDataConfig.RangeAdjustment)
```

### Future: concurrency

Each duration is independent within a function call — `StoreRealizedVols(data, SHORTDURATION)` writes only to `Short` fields and has no dependency on the `Med` or `Long` calls. The sequential range loop is correct and sufficient at current scale (single ticker, interactive use).

When processing thousands of tickers regularly, the natural upgrade path is to replace the range loop with goroutines + `sync.WaitGroup` or `errgroup`. The key constraint at that point: all three duration calls for a single function mutate the same shared map, writing to non-overlapping fields — so the upgrade will require either a per-ticker mutex or a copy-and-merge pattern to avoid map write races.

---

## Section 7: Testing

This refactor is behavior-preserving. The existing unit tests for `GetSimpleSlopes` and `CalculateTrendDirections` pass unchanged.

The primary testing value delivered by this refactor is to issue #8 (TDD coverage pass): each parameterized function can now be tested per-duration independently. A test can call `StoreRealizedVols(data, SHORTDURATION)` and assert only the `Short` fields were populated, without interference from other durations. Writing those tests is out of scope for this PR.

---

## Slope Pipeline Status

`GetSimpleSlopes` and `CalculateTrendDirections` are already wired into both `stockClient` and `batchStocks` (landed in PR #18). All three duration slopes (`SlopeShortDuration`, `SlopeMedDuration`, `SlopeLongDuration`) and all three trend directions (`TradeDirection`, `TrendDirection`, `TailDirection`) are present in both the full `SingleStockCandle` JSON output and the condensed output via `PrepareToPrintData`.

`GetLinearRegressionSlope` remains commented out in both callers — its activation is a separate decision not in scope here.
