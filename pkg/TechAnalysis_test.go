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
					today.UnixMilli():         {Ticker: "AAPL", Close: 100.0, Timestamp: today},
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
					today.UnixMilli():            {Ticker: "AAPL", Close: 100.0, Timestamp: today},
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
