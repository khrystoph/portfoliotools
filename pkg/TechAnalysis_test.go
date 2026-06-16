package pkg

import (
	"reflect"
	"testing"
	"time"
)

func TestRealizedVolatility(t *testing.T) {
	type args struct {
		prices []float64
		ticker string
	}
	tests := []struct {
		name            string
		args            args
		wantRealizedVol float64
	}{
		// TODO: Add test cases.
		{
			name:            "Realized Volatility of 17.646% for AAPL in last 30 calendar days from 2023-12-06 to 2024-01-05",
			args:            args{ticker: "abcd", prices: []float64{181.18, 181.91, 184.25, 185.64, 192.53, 193.58, 193.15, 193.05, 193.60, 194.68, 194.83, 196.94, 195.89, 197.57, 198.11, 197.96, 194.71, 193.18, 195.71, 194.27, 192.32}},
			wantRealizedVol: 0.17646791566477701,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRealizedVol := RealizedVolatility(tt.args.prices, tt.args.ticker); gotRealizedVol != tt.wantRealizedVol {
				t.Errorf("RealizedVolatility() = %v, want %v", gotRealizedVol, tt.wantRealizedVol)
			}
		})
	}
}

func TestCalculateDailyReturn(t *testing.T) {
	type args struct {
		prices []float64
	}
	tests := []struct {
		name string
		args args
		want []float64
	}{
		// TODO: Add test cases.
		{
			name: "default test case",
			args: args{prices: []float64{95, 101, 101, 102, 98, 101, 95, 97, 97, 104, 97, 95, 99, 100, 101, 101, 94, 95, 99, 103}},
			want: []float64{0.061243625240718594, 0, 0.00985229644301164, -0.04000533461369913, 0.030153038170687457, -0.06124362524071867, 0.020834086902842053, 0, 0.0696799206379898, -0.06967992063798982, -0.020834086902842025, 0.041242958534049, 0.010050335853501506, 0.009950330853168092, 0, -0.0718257345712555, 0.010582109330537008, 0.041242958534049, 0.03960913809504588},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateDailyReturn(tt.args.prices); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("calculateDailyReturn() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeTicker(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"all lowercase", "envx", "ENVX"},
		{"all uppercase", "ENVX", "ENVX"},
		{"mixed case", "eNvX", "ENVX"},
		{"crypto lowercase prefix and symbol", "x:eth", "X:ETH"},
		{"crypto uppercase", "X:ETH", "X:ETH"},
		{"crypto mixed case", "X:eTh", "X:ETH"},
		{"leading/trailing whitespace", "  aapl  ", "AAPL"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NormalizeTicker(tt.input); got != tt.want {
				t.Errorf("NormalizeTicker(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

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
		wantMedSlope   float64
		wantMedValid   bool
		wantLongSlope  float64
		wantLongValid  bool
	}{
		{
			name:      "exact 30-day date exists — correct delta and valid=true",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():         {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					thirtyDaysAgo.UnixMilli(): {Ticker: "AAPL", Close: 90.0, Timestamp: thirtyDaysAgo},
				},
			},
			wantShortSlope: 10.0,
			wantShortValid: true,
			wantMedSlope:   0.0,
			wantMedValid:   false,
			wantLongSlope:  0.0,
			wantLongValid:  false,
		},
		{
			name:      "exact 30-day date missing — rolls back to nearest prior day",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():            {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					thirtyTwoDaysAgo.UnixMilli(): {Ticker: "AAPL", Close: 85.0, Timestamp: thirtyTwoDaysAgo},
				},
			},
			wantShortSlope: 15.0,
			wantShortValid: true,
			wantMedSlope:   0.0,
			wantMedValid:   false,
			wantLongSlope:  0.0,
			wantLongValid:  false,
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
			wantMedSlope:   0.0,
			wantMedValid:   false,
			wantLongSlope:  0.0,
			wantLongValid:  false,
		},
		{
			name:      "medium 90-day lookback found — slope and valid set",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():                    {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					today.AddDate(0, 0, -95).UnixMilli(): {Ticker: "AAPL", Close: 70.0, Timestamp: today.AddDate(0, 0, -95)},
				},
			},
			wantShortSlope: 30.0,
			wantShortValid: true,
			wantMedSlope:   30.0,
			wantMedValid:   true,
			wantLongSlope:  0.0,
			wantLongValid:  false,
		},
		{
			name:      "long 180-day lookback found — slope and valid set",
			ticker:    "AAPL",
			checkDate: today.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					today.UnixMilli():                     {Ticker: "AAPL", Close: 100.0, Timestamp: today},
					today.AddDate(0, 0, -185).UnixMilli(): {Ticker: "AAPL", Close: 60.0, Timestamp: today.AddDate(0, 0, -185)},
				},
			},
			wantShortSlope: 40.0,
			wantShortValid: true,
			wantMedSlope:   40.0,
			wantMedValid:   true,
			wantLongSlope:  40.0,
			wantLongValid:  true,
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
			if got.SlopeMedDuration != tt.wantMedSlope {
				t.Errorf("SlopeMedDuration = %v, want %v", got.SlopeMedDuration, tt.wantMedSlope)
			}
			if got.SlopeMedValid != tt.wantMedValid {
				t.Errorf("SlopeMedValid = %v, want %v", got.SlopeMedValid, tt.wantMedValid)
			}
			if got.SlopeLongDuration != tt.wantLongSlope {
				t.Errorf("SlopeLongDuration = %v, want %v", got.SlopeLongDuration, tt.wantLongSlope)
			}
			if got.SlopeLongValid != tt.wantLongValid {
				t.Errorf("SlopeLongValid = %v, want %v", got.SlopeLongValid, tt.wantLongValid)
			}
		})
	}
}

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
			name:      "exactly 2 days in dataset — second day still Indeterminate",
			ticker:    "AAPL",
			checkDate: day2.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
					day2.UnixMilli(): allValid(2.0),
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
		{
			name:      "valid=false on middle day → Indeterminate",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
					day2.UnixMilli(): {
						SlopeShortDuration: 2.0, SlopeShortValid: false,
						SlopeMedDuration: 2.0, SlopeMedValid: false,
						SlopeLongDuration: 2.0, SlopeLongValid: false,
					},
					day3.UnixMilli(): allValid(3.0),
				},
			},
			wantTradeDirection: "Indeterminate",
			wantTrendDirection: "Indeterminate",
			wantTailDirection:  "Indeterminate",
		},
		{
			name:      "valid=false on current day → Indeterminate",
			ticker:    "AAPL",
			checkDate: day3.UnixMilli(),
			stockPrices: map[string]map[int64]SingleStockCandle{
				"AAPL": {
					day1.UnixMilli(): allValid(1.0),
					day2.UnixMilli(): allValid(2.0),
					day3.UnixMilli(): {
						SlopeShortDuration: 3.0, SlopeShortValid: false,
						SlopeMedDuration: 3.0, SlopeMedValid: false,
						SlopeLongDuration: 3.0, SlopeLongValid: false,
					},
				},
			},
			wantTradeDirection: "Indeterminate",
			wantTrendDirection: "Indeterminate",
			wantTailDirection:  "Indeterminate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateTrendDirections(tt.stockPrices, false)
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

func TestCollectWindowDates(t *testing.T) {
	day := int64(24 * 60 * 60 * 1000)
	now := (time.Now().UnixMilli() / day) * day

	dates := make([]int64, 60)
	for i := 0; i < 60; i++ {
		dates[i] = now - int64(i)*day
	}

	window, ok := collectWindowDates(dates, 0, SHORTDURATION)
	if !ok {
		t.Fatal("expected valid window at index 0 with 60 dates and SHORTDURATION=30")
	}
	if len(window) == 0 {
		t.Fatal("expected non-empty window")
	}
	cutoff := dates[0] - int64(SHORTDURATION)*day
	for _, d := range window {
		if d < cutoff {
			t.Errorf("date %v is outside SHORTDURATION window (cutoff %v)", d, cutoff)
		}
	}

	_, ok = collectWindowDates(dates, 58, SHORTDURATION)
	if ok {
		t.Error("expected invalid window near end of slice")
	}

	shortSlice := dates[:10]
	_, ok = collectWindowDates(shortSlice, 0, LONGDURATION)
	if ok {
		t.Error("expected invalid window: slice shorter than LONGDURATION")
	}
}

func TestStoreRealizedVols_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	result := StoreRealizedVols(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.RealizedVolatilityShort != 0 {
				hasShort = true
			}
			if c.RealizedVolatilityMed != 0 {
				t.Errorf("Med should not be set by SHORTDURATION call: got %v", c.RealizedVolatilityMed)
			}
			if c.RealizedVolatilityLong != 0 {
				t.Errorf("Long should not be set by SHORTDURATION call: got %v", c.RealizedVolatilityLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with RealizedVolatilityShort populated")
	}
}

func TestGetAvgVolume_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	result := GetAvgVolume(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.AvgVolumeShort != 0 {
				hasShort = true
			}
			if c.AvgVolumeMed != 0 {
				t.Errorf("AvgVolumeMed should not be set by SHORTDURATION call: got %v", c.AvgVolumeMed)
			}
			if c.AvgVolumeLong != 0 {
				t.Errorf("AvgVolumeLong should not be set by SHORTDURATION call: got %v", c.AvgVolumeLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with AvgVolumeShort populated")
	}
}

func TestGetRelHighLowVol_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	data = StoreRealizedVols(data, SHORTDURATION)
	result := GetRelHighLowVol(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.RVolHighShort != 0 {
				hasShort = true
			}
			if c.RVolHighMed != 0 {
				t.Errorf("RVolHighMed should not be set by SHORTDURATION call: got %v", c.RVolHighMed)
			}
			if c.RVolHighLong != 0 {
				t.Errorf("RVolHighLong should not be set by SHORTDURATION call: got %v", c.RVolHighLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with RVolHighShort populated")
	}
}

func TestCalculateRiskRanges_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	data = StoreRealizedVols(data, SHORTDURATION)
	result := CalculateRiskRanges(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if len(c.TradeRange) > 0 {
				hasShort = true
			}
			if len(c.TrendRange) > 0 {
				t.Errorf("TrendRange should not be set by SHORTDURATION call")
			}
			if len(c.TailRange) > 0 {
				t.Errorf("TailRange should not be set by SHORTDURATION call")
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with TradeRange populated")
	}
}

func TestCalculateAvgVolumeRatios_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	data = GetAvgVolume(data, SHORTDURATION)
	result := CalculateAvgVolumeRatios(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.AvgVolumeRatioShort != 0 {
				hasShort = true
			}
			if c.AvgVolumeRatioMed != 0 {
				t.Errorf("AvgVolumeRatioMed should not be set by SHORTDURATION call: got %v", c.AvgVolumeRatioMed)
			}
			if c.AvgVolumeRatioLong != 0 {
				t.Errorf("AvgVolumeRatioLong should not be set by SHORTDURATION call: got %v", c.AvgVolumeRatioLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with AvgVolumeRatioShort populated")
	}
}

func TestCalculateVolumeAdjustedRiskRanges_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	data = StoreRealizedVols(data, SHORTDURATION)
	data = GetAvgVolume(data, SHORTDURATION)
	data = CalculateAvgVolumeRatios(data, SHORTDURATION)
	result := CalculateVolumeAdjustedRiskRanges(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if len(c.TradeRangeAdj) > 0 {
				hasShort = true
			}
			if len(c.TrendRangeAdj) > 0 {
				t.Errorf("TrendRangeAdj should not be set by SHORTDURATION call")
			}
			if len(c.TailRangeAdj) > 0 {
				t.Errorf("TailRangeAdj should not be set by SHORTDURATION call")
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with TradeRangeAdj populated")
	}
}

func TestGetProbAdjRiskRanges_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 90)
	data = StoreRealizedVols(data, SHORTDURATION)
	data = GetAvgVolume(data, SHORTDURATION)
	data = CalculateAvgVolumeRatios(data, SHORTDURATION)
	data = CalculateRiskRanges(data, SHORTDURATION)
	data = CalculateVolumeAdjustedRiskRanges(data, SHORTDURATION)
	result := GetProbAdjRiskRanges(data, SHORTDURATION, 0.1)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if len(c.PTradeRange) > 0 {
				hasShort = true
			}
			if len(c.PTrendRange) > 0 {
				t.Errorf("PTrendRange should not be set by SHORTDURATION call")
			}
			if len(c.PTailRange) > 0 {
				t.Errorf("PTailRange should not be set by SHORTDURATION call")
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with PTradeRange populated")
	}
}

func TestCalculateVelocities_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 60)
	data = StoreRealizedVols(data, SHORTDURATION)
	result := CalculateVelocities(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.VelocityRealizedVolShort != 0 {
				hasShort = true
			}
			if c.VelocityRealizedVolMed != 0 {
				t.Errorf("VelocityRealizedVolMed should not be set by SHORTDURATION call: got %v", c.VelocityRealizedVolMed)
			}
			if c.VelocityRealizedVolLong != 0 {
				t.Errorf("VelocityRealizedVolLong should not be set by SHORTDURATION call: got %v", c.VelocityRealizedVolLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with VelocityRealizedVolShort populated")
	}
}

func TestCalculateAccelerations_PopulatesOnlyTargetDuration(t *testing.T) {
	data := makeTestData("AAPL", 60)
	data = StoreRealizedVols(data, SHORTDURATION)
	data = CalculateVelocities(data, SHORTDURATION)
	result := CalculateAccelerations(data, SHORTDURATION)
	hasShort := false
	for _, candles := range result {
		for _, c := range candles {
			if c.RealizedVolAccelShort != 0 {
				hasShort = true
			}
			if c.RealizedVolAccelMed != 0 {
				t.Errorf("RealizedVolAccelMed should not be set by SHORTDURATION call: got %v", c.RealizedVolAccelMed)
			}
			if c.RealizedVolAccelLong != 0 {
				t.Errorf("RealizedVolAccelLong should not be set by SHORTDURATION call: got %v", c.RealizedVolAccelLong)
			}
		}
	}
	if !hasShort {
		t.Error("expected at least one candle with RealizedVolAccelShort populated")
	}
}
