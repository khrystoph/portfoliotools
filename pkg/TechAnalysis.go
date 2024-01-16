package pkg

import (
	"context"
	"errors"
	"fmt"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	gonum "gonum.org/v1/gonum/stat"
	"log"
	"math"
	"sort"
	"time"
)

const DAY = 24
const YEAR = 365
const LONGDURATION = 90
const MEDIUMDURATION = 60
const SHORTDURATION = 30
const TRADINGDAYSPERYEAR = 252

// OHLC is a struct that contains the Open, High, Low, and Close values from a range of times for a specific ticker
type OHLC struct {
	Close            []float64 `json:"c"`
	High             []float64 `json:"h"`
	Low              []float64 `json:"l"`
	Status           string    `json:"s"`
	Timestamp        []int64   `json:"t"`
	TransactionCount []int64   `json:"n"`
	Volume           []int64   `json:"v"`
}

type SingleStockCandle struct {
	Ticker               string             `json:"ticker"`
	Close                float64            `json:"close"`
	High                 float64            `json:"high"`
	Low                  float64            `json:"low"`
	Open                 float64            `json:"open"`
	Transactions         int64              `json:"transactions"`
	Timestamp            time.Time          `json:"timestamp"`
	Volume               float64            `json:"volume"`
	WeightedVolume       float64            `json:"weighted-volume"`
	ThirtyDaysPrices     map[string]float64 `json:"30-days-prices"`
	SixtyDaysPrices      map[string]float64 `json:"60-days-prices"`
	NinetyDaysPrices     map[string]float64 `json:"90-days-prices"`
	RealizedVolatility30 float64            `json:"30-day-realized-volatility"`
	RealizedVolatility60 float64            `json:"60-day-realized-volatility"`
	RealizedVolatility90 float64            `json:"90-day-realized-volatility"`
	TradeRange           map[string]float64 `json:"trade-range"`
	TrendRange           map[string]float64 `json:"trend-range"`
	TailRange            map[string]float64 `json:"tail-range"`
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func GetCurrAnnualReturn(currentPrice, costBasis float64, purchaseDate time.Time, isShort bool) (currentAnnualizedReturn float64, err error) {
	hoursOwned := truncateToDay(time.Now()).Sub(truncateToDay(purchaseDate)).Hours()
	daysOwned := hoursOwned / DAY
	var totReturn float64
	if !isShort {
		totReturn = currentPrice - costBasis
	} else {
		totReturn = costBasis - currentPrice
	}

	fmt.Printf("Hours owned: %f.\n", hoursOwned)
	currentAnnualizedReturn = math.Pow((totReturn+costBasis)/costBasis, YEAR/daysOwned) - 1

	return currentAnnualizedReturn, nil
}

/* GetTargetAnnualReturn uses the formula: AP = ((P + G) / P) ^ (365 / n) - 1 solving for G given
 * AP, P, and n. N is days owned, P is costBasis, and AP is the risk-free rate/target rate
 * The new formula should look like this: G = (365/n)√(AP + 1) * P - P
 * We can drop the "-P" at the end because we don't just want the gains, we want the new price as the price target.
 * Additionally, since a root of a number is an inverse power, we can flip the 365/n to be n/365 and use math.Pow.
 */
func GetTargetAnnualReturn(costBasis, riskFreeRate float64, purchaseDate time.Time, isShort bool) (targetAnnualReturnPrice float64, err error) {
	hoursOwned := truncateToDay(time.Now()).Sub(truncateToDay(purchaseDate)).Hours()
	daysOwned := hoursOwned / DAY

	baseReturn := math.Pow(riskFreeRate+1, daysOwned/YEAR) * costBasis

	if math.IsNaN(baseReturn) {
		err = errors.New("result was NaN")
		return math.NaN(), err
	}

	if !isShort {
		targetAnnualReturnPrice = baseReturn
	} else {
		targetAnnualReturnPrice = 2*costBasis - baseReturn
	}
	return targetAnnualReturnPrice, nil
}

// GetStockPrices grabs a set of prices for a ticker over a duration and returns the set
func GetStockPrices(ticker, apiToken, resolution string, startTimeMilli, endTimeMilli time.Time) (stockPrices map[string]map[int64]SingleStockCandle, err error) {
	polygonClient := polygon.New(apiToken)
	if _, ok := stockPrices[ticker]; !ok {
		stockPrices = map[string]map[int64]SingleStockCandle{}
		stockPrices[ticker] = map[int64]SingleStockCandle{}
	}

	// set params
	params := models.ListAggsParams{
		Ticker:     ticker,
		Multiplier: 1,
		Timespan:   models.Timespan(resolution),
		From:       models.Millis(startTimeMilli),
		To:         models.Millis(endTimeMilli),
	}.WithOrder(models.Desc).WithLimit(50000).WithAdjusted(true)

	// make request
	iter := polygonClient.ListAggs(context.Background(), params)

	// do something with the result
	for iter.Next() {
		ts := time.Time(iter.Item().Timestamp).UnixMilli()
		stockPrices[ticker][ts] = SingleStockCandle{
			ticker,
			iter.Item().Close,
			iter.Item().High,
			iter.Item().Low,
			iter.Item().Open,
			iter.Item().Transactions,
			time.Time(iter.Item().Timestamp),
			iter.Item().Volume,
			iter.Item().VWAP,
			// initialize containers for last n days of prices for various durations
			make(map[string]float64),
			make(map[string]float64),
			make(map[string]float64),
			// values filled below here are calculated later in the process and filled with zeros for now
			// RealizedVolatility values over various durations
			0.0,
			0.0,
			0.0,
			// Trade, Trend, and Tail range values
			make(map[string]float64),
			make(map[string]float64),
			make(map[string]float64),
		}
	}
	if iter.Err() != nil {
		log.Fatal(iter.Err())
	}
	return stockPrices, nil
}

/*
ImpliedVolatility calculates the implied volatility of prices on varying timelines. It's used to calculate whether
there is a discount on volatility compared to what is realized. This can be used to determine if options risk-reward
ratio is favorable and help time normal position entries.
*/
func ImpliedVolatility() (impliedVol float64, err error) {
	return
}

// This section of functions deals specifically with calculating volatility of an asset

// calculateDailyReturn generates a slice of float64 values that represents the % return day over day
func calculateDailyReturn(prices []float64) []float64 {
	var returns []float64

	for i := 1; i < len(prices); i++ {
		dailyReturn := math.Log(prices[i] / prices[i-1])
		returns = append(returns, dailyReturn)
	}

	return returns
}

// calculateVariance gives the average variance
func calculateVariance(returns []float64) float64 {
	return gonum.Variance(returns, nil)
}

// calculateMean calculates the arithmetic mean of a float64 slice of inputs and returns the resulting mean
func calculateMean(returns []float64) float64 {
	var sum float64

	for _, r := range returns {
		sum += r
	}

	return sum / float64(len(returns))
}

/*
RealizedVolatility calculates the volatility of prices on varying timelines. It's used to calculate the volatility
that has impacted price and compare with Implied Volatility to evaluate the risk-reward ratio and if it is favorable to
enter a trade.
*/
func RealizedVolatility(prices []float64) (realizedVol float64) {
	returns := calculateDailyReturn(prices)
	variance := calculateVariance(returns)
	return math.Sqrt(variance * TRADINGDAYSPERYEAR)
}

func StoreRealizedVols(stockPrices map[string]map[int64]SingleStockCandle, ticker string) (stockPriceData map[string]map[int64]SingleStockCandle) {
	var dateKeys []int64
	for dateKey := range stockPrices[ticker] {
		dateKeys = append(dateKeys, dateKey)
	}
	reverseDateKeys := dateKeys
	// Sort our date keys in reverse order such that the most recent date is first and the oldest date is last
	sort.Slice(reverseDateKeys, func(i, j int) bool {
		return reverseDateKeys[i] > reverseDateKeys[j]
	})
	for index, date := range reverseDateKeys {
		stockCandle := stockPrices[ticker][date]
		shortDurationStartMilli := time.UnixMilli(date).AddDate(0, 0, -1*SHORTDURATION).UnixMilli()
		medDurationStartMilli := time.UnixMilli(date).AddDate(0, 0, -1*MEDIUMDURATION).UnixMilli()
		longDurationStartMilli := time.UnixMilli(date).AddDate(0, 0, -1*LONGDURATION).UnixMilli()

		if index+SHORTDURATION < len(reverseDateKeys) && reverseDateKeys[index] >= shortDurationStartMilli {
			var volDatesShort []int64
			for shortIndex := index; reverseDateKeys[shortIndex] >= shortDurationStartMilli; shortIndex++ {
				volDatesShort = append(volDatesShort, reverseDateKeys[shortIndex])
			}
			fmt.Printf("number of days in short duration vol dates slice: %d\n", len(volDatesShort))
			stockCandle.ThirtyDaysPrices, stockCandle.RealizedVolatility30 = calculateVolatility(volDatesShort, stockPrices, ticker)
		}
		if index+MEDIUMDURATION < len(reverseDateKeys) && reverseDateKeys[index] >= medDurationStartMilli {
			var volDatesMed []int64
			for medIndex := index; reverseDateKeys[medIndex] >= medDurationStartMilli; medIndex++ {
				volDatesMed = append(volDatesMed, reverseDateKeys[medIndex])
			}
			fmt.Printf("number of days in medium duration vol dates slice: %d\n", len(volDatesMed))
			stockCandle.SixtyDaysPrices, stockCandle.RealizedVolatility60 = calculateVolatility(volDatesMed, stockPrices, ticker)
		}
		if index+LONGDURATION < len(reverseDateKeys) && reverseDateKeys[index] >= longDurationStartMilli {
			var volDatesLong []int64
			for longIndex := index; reverseDateKeys[longIndex] >= longDurationStartMilli; longIndex++ {
				volDatesLong = append(volDatesLong, reverseDateKeys[longIndex])
			}
			fmt.Printf("number of days in long duration vol dates slice: %d\n\n", len(volDatesLong))
			stockCandle.NinetyDaysPrices, stockCandle.RealizedVolatility90 = calculateVolatility(volDatesLong, stockPrices, ticker)
		}
		stockPrices[ticker][date] = stockCandle
	}

	return stockPrices
}

// calculateVolatility calculates the realized volatility for various timeframes
func calculateVolatility(volDatesArray []int64,
	stockPrices map[string]map[int64]SingleStockCandle, ticker string) (stockData map[string]float64, periodVol float64) {
	var prices []float64
	var priceMap = make(map[string]float64)
	for _, dateMilli := range volDatesArray {
		dateOnly := time.UnixMilli(dateMilli).Format(time.DateOnly)
		priceMap[dateOnly] = stockPrices[ticker][dateMilli].Close
		prices = append(prices, stockPrices[ticker][dateMilli].Close)
	}
	realizedVolPeriod := RealizedVolatility(prices)
	return priceMap, realizedVolPeriod
}

func CalculateRiskRanges(stockPrices map[string]map[int64]SingleStockCandle) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for day := range stockPrices[ticker] {
			dailyTicker := stockPrices[ticker][day]
			if stockPrices[ticker][day].RealizedVolatility30 != 0.0 {
				dailyTicker.TradeRange = calculateRiskRange(stockPrices[ticker][day].Close, stockPrices[ticker][day].RealizedVolatility30, SHORTDURATION)
			}
			if stockPrices[ticker][day].RealizedVolatility60 != 0.0 {
				dailyTicker.TrendRange = calculateRiskRange(stockPrices[ticker][day].Close, stockPrices[ticker][day].RealizedVolatility60, MEDIUMDURATION)
			}
			if stockPrices[ticker][day].RealizedVolatility90 != 0.0 {
				dailyTicker.TailRange = calculateRiskRange(stockPrices[ticker][day].Close, stockPrices[ticker][day].RealizedVolatility90, LONGDURATION)
			}
			stockPrices[ticker][day] = dailyTicker
		}
	}
	return stockPrices
}

func calculateRiskRange(price, volatility, riskRangeDuration float64) (riskRange map[string]float64) {
	riskRange = make(map[string]float64)
	riskRange["high"] = (1 + (volatility / TRADINGDAYSPERYEAR * riskRangeDuration)) * price
	riskRange["low"] = (1 - (volatility / TRADINGDAYSPERYEAR * riskRangeDuration)) * price
	return
}
