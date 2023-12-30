package pkg

import (
	"context"
	"fmt"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	"log"
	"math"
	"time"
)

const DAY = 24
const YEAR = 365

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
	Ticker         string    `json:"ticker"`
	Close          float64   `json:"close"`
	High           float64   `json:"high"`
	Low            float64   `json:"low"`
	Open           float64   `json:"open"`
	Transactions   int64     `json:"transactions"`
	Timestamp      time.Time `json:"timestamp"`
	Volume         float64   `json:"volume"`
	WeightedVolume float64   `json:"weightedvolume"`
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

/* GetTargetAnnual Return uses the formula: AP = ((P + G) / P) ^ (365 / n) - 1 solving for G given
 * AP, P, and n. N is days owned, P is costBasis, and AP is the risk-free rate/target rate
 * The new formula should look like this: G = (365/n)√(AP + 1) * P - P
 * We can drop the "-P" at the end because we don't just want the gains, we want the new price as the price target.
 * Additionally, since a root of a number is an inverse power, we can flip the 365/n to be n/365 and use math.Pow.
 */
func GetTargetAnnualReturn(costBasis, riskFreeRate float64, purchaseDate time.Time) (targetAnnualReturnPrice float64, err error) {
	hoursOwned := truncateToDay(time.Now()).Sub(truncateToDay(purchaseDate)).Hours()
	daysOwned := hoursOwned / DAY

	targetAnnualReturnPrice = math.Pow(riskFreeRate+1, daysOwned/YEAR) * costBasis
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

/*
RealizedVariance calculates the variability in price of the current day as compared with the previous day. This is
foundational to calculating realized volatility.
*/
func RealizedVariance(currPrice, prevPrice float64) (periodVariance float64) {
	periodVariance = math.Log10(currPrice) - math.Log10(prevPrice)
	return
}

/*
RealizedVolatility calculates the volatility of prices on varying timelines. It's used to calculate the volatility
that has impacted price and compare with Implied Volatility to evaluate the risk-reward ratio and if it is favorable to
enter a trade.
*/
func RealizedVolatility(prices []float64) (realizedVol float64) {
	var sumRealizedVar float64
	for index, price := range prices {
		if index+1 > len(prices) {
			break
		}
		priceVariance := RealizedVariance(price, prices[index+1])
		sumRealizedVar += math.Pow(priceVariance, 2)
	}
	realizedVol = math.Sqrt(sumRealizedVar)
	return
}
