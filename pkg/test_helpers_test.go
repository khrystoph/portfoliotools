package pkg

import "time"

// makeTestData builds a map of numDays daily candles for the given ticker,
// with close prices incrementing by 0.5 per day from 100.0.
func makeTestData(ticker string, numDays int) map[string]map[int64]SingleStockCandle {
	data := map[string]map[int64]SingleStockCandle{}
	data[ticker] = map[int64]SingleStockCandle{}
	now := time.Now()
	for i := 0; i < numDays; i++ {
		ts := now.AddDate(0, 0, -i)
		data[ticker][ts.UnixMilli()] = SingleStockCandle{
			Ticker:         ticker,
			Close:          100.0 + float64(i)*0.5,
			High:           102.0 + float64(i)*0.5,
			Low:            98.0 + float64(i)*0.5,
			Open:           99.0 + float64(i)*0.5,
			Volume:         1_000_000.0,
			WeightedVolume: 100.5 + float64(i)*0.5,
			Timestamp:      ts,
		}
	}
	return data
}
