# Trend Tracking Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement per-day slope calculation and Bullish/Bearish/Neutral/Indeterminate direction labeling across Trade (30-day), Trend (90-day), and Tail (180-day) durations, output to JSON and Excel.

**Architecture:** Two new pipeline functions — `GetSimpleSlopes` computes raw price delta per duration per day (with a validity flag to distinguish "not computed" from a true zero), then `CalculateTrendDirections` reads those slopes to label each day based on whether the three most recent consecutive slopes are all positive, all negative, or mixed. Both follow the existing one-function-per-metric pattern and are wired into `batchStocks.go` after the existing chain.

**Tech Stack:** Go 1.21+, `github.com/xuri/excelize/v2` for Excel output, standard `sort` and `time` packages, table-driven tests in `testing` package.

---

## File Map

| File | Action | Responsibility |
|------|--------|----------------|
| `pkg/portfolioToolsStructs.go` | Modify | Add direction fields + validity flags to structs; replace Slope/Trend in CondensedRangesJSON |
| `pkg/TechAnalysis.go` | Modify | Rewrite `GetSimpleSlopes`; add `CalculateTrendDirections`; update `PrepareToPrintData` |
| `pkg/TechAnalysis_test.go` | Modify | Add table-driven unit tests for both new functions |
| `pkg/Excelizer.go` | Modify | New `showTail bool` param, updated columns, direction cell coloring |
| `cmd/batchStocks/batchStocks.go` | Modify | Add `-tail` flag, wire two new pipeline calls, update CondensedRangesJSON population |

---

## Task 1: Update Data Structures

**Files:**
- Modify: `pkg/portfolioToolsStructs.go`

- [ ] **Step 1: Add validity flags and direction fields to `SingleStockCandle`**

In `portfolioToolsStructs.go`, add after the existing `SlopeLongDuration` field:

```go
// slope validity flags — json:"-" keeps them out of JSON output
SlopeShortValid bool `json:"-"`
SlopeMedValid   bool `json:"-"`
SlopeLongValid  bool `json:"-"`
// direction labels computed from 3-day slope sign check
TradeDirection string `json:"trade-direction"`
TrendDirection string `json:"trend-direction"`
TailDirection  string `json:"tail-direction"`
```

- [ ] **Step 2: Add direction fields to `condensedStockCandle`**

In `condensedStockCandle`, add after the existing `TailSlope` field:

```go
TradeDirection string `json:"trade-direction"`
TrendDirection string `json:"trend-direction"`
TailDirection  string `json:"tail-direction"`
```

- [ ] **Step 3: Replace `Slope` and `Trend` in `CondensedRangesJSON` and remove `Volume`/`Rvol`**

Replace the entire `CondensedRangesJSON` struct with:

```go
type CondensedRangesJSON struct {
	Ticker         string    `json:"ticker"`
	Close          float64   `json:"close"`
	AvgVolRatio    float64   `json:"avg_vol_ratio"`
	RVolPercent    float64   `json:"rvol_percent"`
	RiskRangeHigh  float64   `json:"rr_high"`
	RiskRangeLow   float64   `json:"rr_low"`
	TradeSlope     float64   `json:"trade-slope"`
	TrendSlope     float64   `json:"trend-slope"`
	TailSlope      float64   `json:"tail-slope"`
	TradeDirection string    `json:"trade-direction"`
	TrendDirection string    `json:"trend-direction"`
	TailDirection  string    `json:"tail-direction"`
	Timestamp      time.Time `json:"timestamp"`
}
```

- [ ] **Step 4: Verify compilation**

```bash
cd /Users/bret/github/portfoliotools && go build ./...
```

Expected: compile errors in `batchStocks.go` referencing removed fields `Volume`, `Rvol`, `Slope`, `Trend` — these are fixed in Task 8. For now confirm only those files error.

- [ ] **Step 5: Commit**

```bash
git add pkg/portfolioToolsStructs.go
git commit -m "feat: update structs for trend tracking — add direction fields, validity flags, per-duration slopes"
```

---

## Task 2: Write Failing Tests for `GetSimpleSlopes`

**Files:**
- Modify: `pkg/TechAnalysis_test.go`

- [ ] **Step 1: Add the test function**

Append to `pkg/TechAnalysis_test.go`:

```go
func TestGetSimpleSlopes(t *testing.T) {
	today := time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC)
	thirtyDaysAgo := today.AddDate(0, 0, -30)    // exact target for SHORT
	thirtyTwoDaysAgo := today.AddDate(0, 0, -32) // nearest prior when exact missing
	tenDaysAgo := today.AddDate(0, 0, -10)       // too recent for 30-day lookback

	tests := []struct {
		name           string
		stockPrices    map[string]map[int64]SingleStockCandle
		ticker         string
		checkDate      int64
		wantShortSlope float64
		wantShortValid bool
	}{
		{
			name:      "exact 30-day date exists — correct delta and valid=true",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():        {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					thirtyDaysAgo.UnixMilli(): {Ticker: "AAPL", Close: 90.0, Timestamp: thirtyDaysAgo},
				},
			},
			wantShortSlope: 10.0,
			wantShortValid: true,
		},
		{
			name:      "exact 30-day date missing — rolls back to nearest prior day",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():          {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					thirtyTwoDaysAgo.UnixMilli(): {Ticker: "AAPL", Close: 85.0, Timestamp: thirtyTwoDaysAgo},
				},
			},
			wantShortSlope: 15.0,
			wantShortValid: true,
		},
		{
			name:      "no data old enough for 30-day lookback — slope=0, valid=false",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():      {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					tenDaysAgo.UnixMilli(): {Ticker: "AAPL", Close: 95.0, Timestamp: tenDaysAgo},
				},
			},
			wantShortSlope: 0.0,
			wantShortValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetSimpleSlopes(tt.stockPrices, false)
			got := result[tt.ticker][tt.checkDate]
			if got.SlopeShortDuration != tt.wantShortSlope {
				t.Errorf("SlopeShortDuration = %v, want %v", got.SlopeShortDuration, tt.wantShortSlope)
			}
			if got.SlopeShortValid != tt.wantShortValid {
				t.Errorf("SlopeShortValid = %v, want %v", got.SlopeShortValid, tt.wantShortValid)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/bret/github/portfoliotools && go test ./pkg/... -v -run TestGetSimpleSlopes
```

Expected: FAIL — the existing stub overwrites its input with an empty map so all assertions fail.

---

## Task 3: Implement `GetSimpleSlopes`

**Files:**
- Modify: `pkg/TechAnalysis.go`

- [ ] **Step 1: Replace the broken stub with a correct implementation**

Replace the entire existing `GetSimpleSlopes` function (lines 770–794) with:

```go
// GetSimpleSlopes computes the raw price delta for each duration per day.
// For each day it looks back N calendar days, rolling back one day at a time
// until finding a trading day at or before the target, then computes:
//   slope = close_today - close_at_lookback_date
//
// SlopeXxxValid is set true only when a lookback date was found; false means
// insufficient history and the slope value of 0.0 is meaningless.
func GetSimpleSlopes(stockPrices map[string]map[int64]SingleStockCandle, isDebug bool) map[string]map[int64]SingleStockCandle {
	for ticker := range stockPrices {
		var dateKeys []int64
		for dateKey := range stockPrices[ticker] {
			dateKeys = append(dateKeys, dateKey)
		}
		// Sort descending so dateKeys[0] is most recent; the inner loop's
		// first match at or before a target is the nearest-prior trading day.
		sort.Slice(dateKeys, func(i, j int) bool {
			return dateKeys[i] > dateKeys[j]
		})

		for _, currentDate := range dateKeys {
			stockCandle := stockPrices[ticker][currentDate]
			currentClose := stockPrices[ticker][currentDate].Close

			shortTarget := time.UnixMilli(currentDate).AddDate(0, 0, -SHORTDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= shortTarget {
					stockCandle.SlopeShortDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeShortValid = true
					break
				}
			}

			medTarget := time.UnixMilli(currentDate).AddDate(0, 0, -MEDIUMDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= medTarget {
					stockCandle.SlopeMedDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeMedValid = true
					break
				}
			}

			longTarget := time.UnixMilli(currentDate).AddDate(0, 0, -LONGDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= longTarget {
					stockCandle.SlopeLongDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeLongValid = true
					break
				}
			}

			stockPrices[ticker][currentDate] = stockCandle
		}
	}
	return stockPrices
}
```

- [ ] **Step 2: Run tests to confirm they pass**

```bash
cd /Users/bret/github/portfoliotools && go test ./pkg/... -v -run TestGetSimpleSlopes
```

Expected: PASS — all three cases.

- [ ] **Step 3: Commit**

```bash
git add pkg/TechAnalysis.go pkg/TechAnalysis_test.go
git commit -m "feat: implement GetSimpleSlopes with calendar-day lookback and validity flags"
```

---

## Task 4: Write Failing Tests for `CalculateTrendDirections`

**Files:**
- Modify: `pkg/TechAnalysis_test.go`

- [ ] **Step 1: Add the test function**

Append to `pkg/TechAnalysis_test.go`:

```go
func TestCalculateTrendDirections(t *testing.T) {
	day1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	day2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	day3 := time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC)

	// allValid builds a candle with all three slopes valid and set to the given value.
	allValid := func(slope float64) SingleStockCandle {
		return SingleStockCandle{
			SlopeShortDuration: slope, SlopeShortValid: true,
			SlopeMedDuration: slope, SlopeMedValid: true,
			SlopeLongDuration: slope, SlopeLongValid: true,
		}
	}

	tests := []struct {
		name               string
		stockPrices        map[string]map[int64]SingleStockCandle
		ticker             string
		checkDate          int64
		wantTradeDirection string
		wantTrendDirection string
		wantTailDirection  string
	}{
		{
			name:      "all three slopes positive → Bullish",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
					day2.UnixMilli(): allValid(2.0),
					day3.UnixMilli(): allValid(3.0),
				},
			},
			wantTradeDirection: "Bullish",
			wantTrendDirection: "Bullish",
			wantTailDirection:  "Bullish",
		},
		{
			name:      "all three slopes negative → Bearish",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(-1.0),
					day2.UnixMilli(): allValid(-2.0),
					day3.UnixMilli(): allValid(-3.0),
				},
			},
			wantTradeDirection: "Bearish",
			wantTrendDirection: "Bearish",
			wantTailDirection:  "Bearish",
		},
		{
			name:      "mixed slopes → Neutral",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
					day2.UnixMilli(): allValid(-1.0),
					day3.UnixMilli(): allValid(1.0),
				},
			},
			wantTradeDirection: "Neutral",
			wantTrendDirection: "Neutral",
			wantTailDirection:  "Neutral",
		},
		{
			name:      "slope exactly 0.0 with valid=true → Neutral, not Indeterminate",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(0.0),
					day2.UnixMilli(): allValid(0.0),
					day3.UnixMilli(): allValid(0.0),
				},
			},
			wantTradeDirection: "Neutral",
			wantTrendDirection: "Neutral",
			wantTailDirection:  "Neutral",
		},
		{
			name:      "fewer than 3 days in dataset → Indeterminate",
			ticker:    "AAPL",
			checkDate: day1.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
				},
			},
			wantTradeDirection: "Indeterminate",
			wantTrendDirection: "Indeterminate",
			wantTailDirection:  "Indeterminate",
		},
		{
			name:      "valid=false on oldest of the three days → Indeterminate",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): {
						SlopeShortDuration: 1.0, SlopeShortValid: false,
						SlopeMedDuration: 1.0, SlopeMedValid: false,
						SlopeLongDuration: 1.0, SlopeLongValid: false,
					},
					day2.UnixMilli(): allValid(2.0),
					day3.UnixMilli(): allValid(3.0),
				},
			},
			wantTradeDirection: "Indeterminate",
			wantTrendDirection: "Indeterminate",
			wantTailDirection:  "Indeterminate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTrendDirections(tt.stockPrices)
			got := result[tt.ticker][tt.checkDate]
			if got.TradeDirection != tt.wantTradeDirection {
				t.Errorf("TradeDirection = %q, want %q", got.TradeDirection, tt.wantTradeDirection)
			}
			if got.TrendDirection != tt.wantTrendDirection {
				t.Errorf("TrendDirection = %q, want %q", got.TrendDirection, tt.wantTrendDirection)
			}
			if got.TailDirection != tt.wantTailDirection {
				t.Errorf("TailDirection = %q, want %q", got.TailDirection, tt.wantTailDirection)
			}
		})
	}
}
```

- [ ] **Step 2: Run tests to confirm they fail**

```bash
cd /Users/bret/github/portfoliotools && go test ./pkg/... -v -run TestCalculateTrendDirections
```

Expected: FAIL — `CalculateTrendDirections` does not exist yet.

---

## Task 5: Implement `CalculateTrendDirections`

**Files:**
- Modify: `pkg/TechAnalysis.go`

- [ ] **Step 1: Add the function after `GetSimpleSlopes`**

```go
// CalculateTrendDirections assigns TradeDirection, TrendDirection, and TailDirection
// to each day based on whether the three most recent consecutive slopes (today,
// yesterday, day-before) are all positive (Bullish), all negative (Bearish),
// mixed (Neutral), or unavailable (Indeterminate).
//
// Must be called after GetSimpleSlopes so that validity flags are set.
func CalculateTrendDirections(stockPrices map[string]map[int64]SingleStockCandle) map[string]map[int64]SingleStockCandle {
	for ticker := range stockPrices {
		var dateKeys []int64
		for dateKey := range stockPrices[ticker] {
			dateKeys = append(dateKeys, dateKey)
		}
		// Sort ascending so index i-1 and i-2 are the prior trading days.
		sort.Slice(dateKeys, func(i, j int) bool {
			return dateKeys[i] < dateKeys[j]
		})

		for i, currentDate := range dateKeys {
			stockCandle := stockPrices[ticker][currentDate]
			if i < 2 {
				stockCandle.TradeDirection = "Indeterminate"
				stockCandle.TrendDirection = "Indeterminate"
				stockCandle.TailDirection = "Indeterminate"
			} else {
				prev1 := stockPrices[ticker][dateKeys[i-1]]
				prev2 := stockPrices[ticker][dateKeys[i-2]]

				stockCandle.TradeDirection = trendLabel(
					stockCandle.SlopeShortDuration, stockCandle.SlopeShortValid,
					prev1.SlopeShortDuration, prev1.SlopeShortValid,
					prev2.SlopeShortDuration, prev2.SlopeShortValid,
				)
				stockCandle.TrendDirection = trendLabel(
					stockCandle.SlopeMedDuration, stockCandle.SlopeMedValid,
					prev1.SlopeMedDuration, prev1.SlopeMedValid,
					prev2.SlopeMedDuration, prev2.SlopeMedValid,
				)
				stockCandle.TailDirection = trendLabel(
					stockCandle.SlopeLongDuration, stockCandle.SlopeLongValid,
					prev1.SlopeLongDuration, prev1.SlopeLongValid,
					prev2.SlopeLongDuration, prev2.SlopeLongValid,
				)
			}
			stockPrices[ticker][currentDate] = stockCandle
		}
	}
	return stockPrices
}

// trendLabel returns the direction label for one duration given three consecutive
// slope values and their validity flags.
func trendLabel(s0 float64, v0 bool, s1 float64, v1 bool, s2 float64, v2 bool) string {
	if !v0 || !v1 || !v2 {
		return "Indeterminate"
	}
	if s0 > 0 && s1 > 0 && s2 > 0 {
		return "Bullish"
	}
	if s0 < 0 && s1 < 0 && s2 < 0 {
		return "Bearish"
	}
	return "Neutral"
}
```

- [ ] **Step 2: Run all pkg tests**

```bash
cd /Users/bret/github/portfoliotools && go test ./pkg/... -v
```

Expected: all tests PASS including the two new suites.

- [ ] **Step 3: Commit**

```bash
git add pkg/TechAnalysis.go pkg/TechAnalysis_test.go
git commit -m "feat: add CalculateTrendDirections with Bullish/Bearish/Neutral/Indeterminate labeling"
```

---

## Task 6: Update `PrepareToPrintData`

**Files:**
- Modify: `pkg/TechAnalysis.go`

- [ ] **Step 1: Add direction fields to the `condensedStockCandle` literal in `PrepareToPrintData`**

In `PrepareToPrintData` in `pkg/TechAnalysis.go`, the `condensedStockCandle` struct literal ends with `PTailRangeAdj`. Add the three direction fields immediately before the closing `}`:

```go
		TradeDirection: stockPrices[ticker][dateInt64].TradeDirection,
		TrendDirection: stockPrices[ticker][dateInt64].TrendDirection,
		TailDirection:  stockPrices[ticker][dateInt64].TailDirection,
	}
```

- [ ] **Step 2: Verify compilation and run tests**

```bash
cd /Users/bret/github/portfoliotools && go build ./... && go test ./pkg/... -v
```

Expected: clean build and all tests pass.

- [ ] **Step 3: Commit**

```bash
git add pkg/TechAnalysis.go
git commit -m "feat: propagate direction fields through PrepareToPrintData"
```

---

## Task 7: Update `Excelizer.go`

**Files:**
- Modify: `pkg/Excelizer.go`

- [ ] **Step 1: Replace `GenerateStockReportXLSX` with the new version**

Replace the entire contents of `pkg/Excelizer.go` with:

```go
package pkg

import (
	"fmt"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
)

// GenerateStockReportXLSX writes the stock report to an Excel file.
// showTail adds Tail Slope and Tail Dir columns (13 cols total vs default 11).
func GenerateStockReportXLSX(data map[string]CondensedRangesJSON, outputPath string, showTail bool) error {
	tickers := make([]string, 0, len(data))
	for t := range data {
		tickers = append(tickers, t)
	}
	sort.Strings(tickers)

	f := excelize.NewFile()
	sheet := "Stock Report"
	index, err := f.NewSheet(sheet)
	if err != nil {
		fmt.Println(err, "Excelizer.go: Could not create new excel sheet")
	}
	f.SetActiveSheet(index)

	// Build header list and determine last column letter for merges/footer.
	headers := []string{
		"Ticker", "Close", "Avg Vol Ratio", "RVol %",
		"VAPR Low", "VAPR High", "Trade Slope", "Trend Slope",
	}
	if showTail {
		headers = append(headers, "Tail Slope")
	}
	headers = append(headers, "Trade Dir", "Trend Dir")
	if showTail {
		headers = append(headers, "Tail Dir")
	}
	headers = append(headers, "Timestamp")

	lastCol, _ := excelize.ColumnNumberToName(len(headers))

	// Title and subtitle
	title := "📊 Stock Range Report"
	subtitle := fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006"))
	f.MergeCell(sheet, "A1", lastCol+"1")
	f.MergeCell(sheet, "A2", lastCol+"2")

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "#004B87"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "#333333"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellValue(sheet, "A1", title)
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellValue(sheet, "A2", subtitle)
	f.SetCellStyle(sheet, "A2", "A2", subtitleStyle)

	// Header row
	headerRow := 4
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1F4E79"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "808080", Style: 1},
			{Type: "top", Color: "808080", Style: 1},
			{Type: "right", Color: "808080", Style: 1},
			{Type: "bottom", Color: "808080", Style: 1},
		},
	})
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Shared border for body and direction cells.
	border := []excelize.Border{
		{Type: "left", Color: "808080", Style: 1},
		{Type: "top", Color: "808080", Style: 1},
		{Type: "right", Color: "808080", Style: 1},
		{Type: "bottom", Color: "808080", Style: 1},
	}

	// Direction cell styles.
	bullishStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#70AD47"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	bearishStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FF0000"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	neutralStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#A5A5A5"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	indeterminateStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 7, Color: []string{"#FF0000", "#FFFFFF"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})

	// Alternating body styles.
	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	altBodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
		Border:    border,
	})

	// Direction column indices (0-based). Layout (default, no tail):
	//   0:Ticker 1:Close 2:AvgVolRatio 3:RVol% 4:VAPRLow 5:VAPRHigh
	//   6:TradeSlope 7:TrendSlope 8:TradeDir 9:TrendDir 10:Timestamp
	// With tail:
	//   0–7 same, 8:TailSlope 9:TradeDir 10:TrendDir 11:TailDir 12:Timestamp
	dirColTrade := 8
	dirColTrend := 9
	dirColTail := -1
	if showTail {
		dirColTrade = 9
		dirColTrend = 10
		dirColTail = 11
	}

	rowStart := headerRow + 1
	rowIndex := rowStart

	for _, ticker := range tickers {
		s := data[ticker]

		row := []interface{}{
			ticker,
			fmt.Sprintf("%.2f", s.Close),
			fmt.Sprintf("%.2f", s.AvgVolRatio),
			fmt.Sprintf("%.2f", s.RVolPercent),
			fmt.Sprintf("%.2f", s.RiskRangeLow),
			fmt.Sprintf("%.2f", s.RiskRangeHigh),
			fmt.Sprintf("%.4f", s.TradeSlope),
			fmt.Sprintf("%.4f", s.TrendSlope),
		}
		if showTail {
			row = append(row, fmt.Sprintf("%.4f", s.TailSlope))
		}
		row = append(row, s.TradeDirection, s.TrendDirection)
		if showTail {
			row = append(row, s.TailDirection)
		}
		row = append(row, s.Timestamp.Format("2006-01-02"))

		for i, val := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			f.SetCellValue(sheet, cell, val)

			var style int
			switch {
			case i == dirColTrade:
				style = directionCellStyle(s.TradeDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case i == dirColTrend:
				style = directionCellStyle(s.TrendDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case showTail && i == dirColTail:
				style = directionCellStyle(s.TailDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case (rowIndex-rowStart)%2 == 0:
				style = altBodyStyle
			default:
				style = bodyStyle
			}
			f.SetCellStyle(sheet, cell, cell, style)
		}
		rowIndex++
	}

	// Freeze header rows.
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      headerRow,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	// Footer.
	footerRow := rowIndex + 1
	f.MergeCell(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("%s%d", lastCol, footerRow))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), "Report generated automatically from stock range data.")
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Font:      &excelize.Font{Italic: true, Size: 10},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("%s%d", lastCol, footerRow), footerStyle)

	// Column widths.
	var widths []float64
	if showTail {
		widths = []float64{12, 10, 14, 12, 12, 12, 12, 12, 12, 12, 12, 12, 18}
	} else {
		widths = []float64{12, 10, 14, 12, 12, 12, 12, 12, 12, 12, 18}
	}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save XLSX: %v", err)
	}
	fmt.Printf("✅ XLSX generated successfully at: %s\n", outputPath)
	return nil
}

// directionCellStyle returns the excelize style ID for a given direction label.
func directionCellStyle(direction string, bullish, bearish, neutral, indeterminate int) int {
	switch direction {
	case "Bullish":
		return bullish
	case "Bearish":
		return bearish
	case "Neutral":
		return neutral
	default:
		return indeterminate
	}
}
```

- [ ] **Step 2: Verify compilation**

```bash
cd /Users/bret/github/portfoliotools && go build ./pkg/...
```

Expected: clean compile (batchStocks.go still broken from Task 1 — fixed next).

- [ ] **Step 3: Commit**

```bash
git add pkg/Excelizer.go
git commit -m "feat: update Excelizer with direction columns, color coding, and optional tail flag"
```

---

## Task 8: Wire `batchStocks.go`

**Files:**
- Modify: `cmd/batchStocks/batchStocks.go`

- [ ] **Step 1: Add `showTail` to the flag var block**

Replace the existing var block at the top of `batchStocks.go`:

```go
var (
	csvFile, outFile, tickerConfig, batchStockRangesFile, timeDuration string
	debug, excelOut, noEmail, showTail                                  bool
)
```

- [ ] **Step 2: Register `-tail` flags in `init()`**

Add to the `init()` function, after the existing flag declarations:

```go
flag.BoolVar(&showTail, "tail", false, "Include Tail Slope and Tail Dir columns in Excel output")
flag.BoolVar(&showTail, "tail-cols", false, "Include Tail Slope and Tail Dir columns in Excel output")
```

- [ ] **Step 3: Add the two new pipeline calls**

After the existing `tickerData = pkg.GetProbAdjRiskRanges(...)` call and before the commented-out `GetLinearRegressionSlope` line, add:

```go
tickerData = pkg.GetSimpleSlopes(tickerData, debug)
tickerData = pkg.CalculateTrendDirections(tickerData)
```

- [ ] **Step 4: Remove `durationRVol` and clean up the switch statement**

Replace the inner `for ticker, stock := range tickerData` variable declaration:

```go
var rrHigh, rrLow, rvolpct, avgvolratio float64
```

Replace the switch statement body to remove `durationRVol` assignments:

```go
switch timeDuration {
case "MEDIUM":
	if isCrypto {
		rrHigh = stock[latestDate].TrendRangeAdj["high"]
		rrLow = stock[latestDate].TrendRangeAdj["low"]
	} else {
		rrHigh = stock[latestDate].PTrendRangeAdj["high"]
		rrLow = stock[latestDate].PTrendRangeAdj["low"]
	}
	rvolpct = stock[latestDate].RVolPercent60
	avgvolratio = stock[latestDate].AvgVolumeRatio60
case "LONG":
	if isCrypto {
		rrHigh = stock[latestDate].TailRangeAdj["high"]
		rrLow = stock[latestDate].TailRangeAdj["low"]
	} else {
		rrHigh = stock[latestDate].PTailRangeAdj["high"]
		rrLow = stock[latestDate].PTailRangeAdj["low"]
	}
	rvolpct = stock[latestDate].RVolPercent90
	avgvolratio = stock[latestDate].AvgVolumeRatio90
case "SHORT":
	fallthrough
default:
	if isCrypto {
		rrHigh = stock[latestDate].TradeRangeAdj["high"]
		rrLow = stock[latestDate].TradeRangeAdj["low"]
	} else {
		rrHigh = stock[latestDate].PTradeRangeAdj["high"]
		rrLow = stock[latestDate].PTradeRangeAdj["low"]
	}
	rvolpct = stock[latestDate].RVolPercent30
	avgvolratio = stock[latestDate].AvgVolumeRatio30
}
```

- [ ] **Step 5: Replace the `CondensedRangesJSON` struct literal**

Replace the existing `batchStockRanges[tickerStripped] = pkg.CondensedRangesJSON{...}` block with:

```go
batchStockRanges[tickerStripped] = pkg.CondensedRangesJSON{
	Ticker:         tickerStripped,
	Close:          stock[latestDate].Close,
	AvgVolRatio:    avgvolratio,
	RVolPercent:    rvolpct,
	RiskRangeHigh:  rrHigh,
	RiskRangeLow:   rrLow,
	TradeSlope:     stock[latestDate].SlopeShortDuration,
	TrendSlope:     stock[latestDate].SlopeMedDuration,
	TailSlope:      stock[latestDate].SlopeLongDuration,
	TradeDirection: stock[latestDate].TradeDirection,
	TrendDirection: stock[latestDate].TrendDirection,
	TailDirection:  stock[latestDate].TailDirection,
	Timestamp:      stock[latestDate].Timestamp,
}
```

- [ ] **Step 6: Pass `showTail` to `GenerateStockReportXLSX`**

Replace:
```go
pkg.GenerateStockReportXLSX(batchStockRanges, excelOutFile)
```
With:
```go
pkg.GenerateStockReportXLSX(batchStockRanges, excelOutFile, showTail)
```

- [ ] **Step 7: Full build**

```bash
cd /Users/bret/github/portfoliotools && go build ./...
```

Expected: clean compile with no errors.

- [ ] **Step 8: Run all tests**

```bash
cd /Users/bret/github/portfoliotools && go test ./pkg/... -v
```

Expected: all tests pass.

- [ ] **Step 9: Commit**

```bash
git add cmd/batchStocks/batchStocks.go
git commit -m "feat: wire slope and direction pipeline into batchStocks, add -tail flag"
```

---

## Done

At this point:
- `stockRanges.json` will contain `trade-slope`, `trend-slope`, `tail-slope`, `trade-direction`, `trend-direction`, `tail-direction` for each ticker.
- `stockRanges.xlsx` will show 11 columns by default (Trade/Trend slope and direction, color-coded) or 13 with `-tail`.
- All new functions have unit tests following the existing table-driven pattern.
- Coverage goal: 90% unit / 70% integration (tracked as a project-wide backlog item).
