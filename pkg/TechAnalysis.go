package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	alpaca "github.com/alpacahq/alpaca-trade-api-go/v3/alpaca"
	polygon "github.com/polygon-io/client-go/rest"
	"github.com/polygon-io/client-go/rest/models"
	gonum "gonum.org/v1/gonum/stat"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"
)

const DAY = 24
const YEAR = 365.24
const LONGDURATION = 90
const MEDIUMDURATION = 60
const SHORTDURATION = 30
const TRADINGDAYSPERYEAR = 252
const ALPACA_PAPER_API = "https://paper-api.alpaca.markets"
const ALPACA_LIVE_API = "https://api.alpaca.markets"

type StockDataConf struct {
	PolygonAPIToken string  `json:"polygon-api-key"`
	AlpacaAPIKey    string  `json:"alpaca-api-key"`
	AlpacaSecretKey string  `json:"alpaca-secret-key"`
	RangeAdjustment float64 `json:"probable-range-adj"`
}

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

type HistoricalBar struct {
	Close            float64 `json:"c"`
	High             float64 `json:"h"`
	Low              float64 `json:"l"`
	TransactionCount int64   `json:"n"`
	Open             float64 `json:"o"`
	Timestamp        string  `json:"t"`
	Volume           int64   `json:"v"`
	WeightedVolume   float64 `json:"vw"`
}

type SingleStockCandle struct {
	Ticker                string             `json:"ticker"`
	Close                 float64            `json:"close"`
	High                  float64            `json:"high"`
	Low                   float64            `json:"low"`
	Open                  float64            `json:"open"`
	Transactions          int64              `json:"transactions"`
	Timestamp             time.Time          `json:"timestamp"`
	Volume                float64            `json:"volume"`
	WeightedVolume        float64            `json:"weighted-volume"`
	AvgVolume30           float64            `json:"avg-volume-30"`
	AvgVolume60           float64            `json:"avg-volume-60"`
	AvgVolume90           float64            `json:"avg-volume-90"`
	AvgVolumeRatio30      float64            `json:"avg-volume-ratio-30"`
	AvgVolumeRatio60      float64            `json:"avg-volume-ratio-60"`
	AvgVolumeRatio90      float64            `json:"avg-volume-ratio-90"`
	ThirtyDaysPrices      map[string]float64 `json:"30-days-prices"`
	SixtyDaysPrices       map[string]float64 `json:"60-days-prices"`
	NinetyDaysPrices      map[string]float64 `json:"90-days-prices"`
	RVolHigh30            float64            `json:"30-day-rvol-high"`
	RVolLow30             float64            `json:"30-day-rvol-low"`
	RVolHigh60            float64            `json:"60-day-rvol-high"`
	RVolLow60             float64            `json:"60-day-rvol-low"`
	RVolHigh90            float64            `json:"90-day-rvol-high"`
	RVolLow90             float64            `json:"90-day-rvol-low"`
	RVolPercent30         float64            `json:"30-day-rvol-range-percent"`
	RVolPercent60         float64            `json:"60-day-rvol-range-percent"`
	RVolPercent90         float64            `json:"90-day-rvol-range-percent"`
	RealizedVolatility30  float64            `json:"30-day-realized-volatility"`
	RealizedVolatility60  float64            `json:"60-day-realized-volatility"`
	RealizedVolatility90  float64            `json:"90-day-realized-volatility"`
	VelocityRealizedVol30 float64            `json:"30-day-realized-volatility-velocity"`
	VelocityRealizedVol60 float64            `json:"60-day-realized-volatility-velocity"`
	VelocityRealizedVol90 float64            `json:"90-day-realized-volatility-velocity"`
	RealizedVolAccel30    float64            `json:"30-day-realized-volatility-accel"`
	RealizedVolAccel60    float64            `json:"60-day-realized-volatility-accel"`
	RealizedVolAccel90    float64            `json:"90-day-realized-volatility-accel"`
	TradeRange            map[string]float64 `json:"trade-range"`
	TrendRange            map[string]float64 `json:"trend-range"`
	TailRange             map[string]float64 `json:"tail-range"`
	TradeRangeAdj         map[string]float64 `json:"trade-range-vadj"`
	TrendRangeAdj         map[string]float64 `json:"trend-range-vadj"`
	TailRangeAdj          map[string]float64 `json:"tail-range-vadj"`
	PTradeRange           map[string]float64 `json:"prob-adj-trade-range"`
	PTrendRange           map[string]float64 `json:"prob-adj-trend-range"`
	PTailRange            map[string]float64 `json:"prob-adj-tail-range"`
	PTradeRangeAdj        map[string]float64 `json:"prob-trade-range-vadj"`
	PTrendRangeAdj        map[string]float64 `json:"prob-trend-range-vadj"`
	PTailRangeAdj         map[string]float64 `json:"prob-tail-range-vadj"`
}

type condensedStockCandle struct {
	Ticker              string             `json:"ticker"`
	Close               float64            `json:"close"`
	Volume              float64            `json:"volume"`
	Timestamp           time.Time          `json:"timestamp"`
	AvgVolumeShort      float64            `json:"short-avg-volume"`
	AvgVolumeRatioShort float64            `json:"short-avg-volume-ratio"`
	RVolShort           float64            `json:"rvol-short"`
	RVolShortVel        float64            `json:"rvol-short-vel"`
	RVolShortAccel      float64            `json:"rvol-short-accel"`
	RVolPercentShort    float64            `json:"short-day-rvol-range-percent"`
	RVolHighShort       float64            `json:"short-day-rvol-high"`
	RVolLowShort        float64            `json:"short-day-rvol-low"`
	TradeRangeAdj       map[string]float64 `json:"trade-range-vadj"`
	PtradeRangeAdj      map[string]float64 `json:"prob-trade-range-vadj"`
	AvgVolumeMed        float64            `json:"med-avg-volume"`
	AvgVolumeRatioMed   float64            `json:"med-avg-volume-ratio"`
	RVolMed             float64            `json:"rvol-med"`
	RVolMedVel          float64            `json:"rvol-med-vel"`
	RVolMedAccel        float64            `json:"rvol-med-accel"`
	RVolPercentMed      float64            `json:"med-day-rvol-range-percent"`
	RVolHighMed         float64            `json:"med-day-rvol-high"`
	RVolLowMed          float64            `json:"med-day-rvol-low"`
	TrendRangeAdj       map[string]float64 `json:"trend-range-vadj"`
	PTrendRangeAdj      map[string]float64 `json:"prob-trend-range-vadj"`
	AvgVolumeLong       float64            `json:"long-avg-volume"`
	AvgVolumeRatioLong  float64            `json:"long-avg-volume-ratio"`
	RVolLong            float64            `json:"rvol-long"`
	RVolLongVel         float64            `json:"rvol-long-vel"`
	RVolLongAccel       float64            `json:"rvol-long-accel"`
	RVolPercentLong     float64            `json:"long-day-rvol-range-percent"`
	RVolHighLong        float64            `json:"long-day-rvol-high"`
	RVolLowLong         float64            `json:"long-day-rvol-low"`
	TailRangeAdj        map[string]float64 `json:"tail-range-vadj"`
	PTailRangeAdj       map[string]float64 `json:"prob-tail-range-vadj"`
}

func PrintData(stockPrices map[string]map[int64]SingleStockCandle, debug bool) {
	var jsonTickerData []byte
	var err error

	// Prepare the condensed data
	condensedPrices := PrepareToPrintData(stockPrices)

	// Print the data
	if debug {
		jsonTickerData, err = json.MarshalIndent(stockPrices, "", "  ")
	} else {
		jsonTickerData, err = json.MarshalIndent(condensedPrices, "", "  ")
	}
	if err != nil {
		_ = fmt.Errorf("error marshalling data into JSON string")
		os.Exit(1)
	}
	fmt.Printf("%v\n", string(jsonTickerData))
}

func PrepareToPrintData(stockPrices map[string]map[int64]SingleStockCandle) (condensedPrices map[string]map[int64]condensedStockCandle) {
	condensedPrices = map[string]map[int64]condensedStockCandle{}

	// prepare the condensed data
	for ticker := range stockPrices {
		if _, ok := condensedPrices[ticker]; !ok {
			condensedPrices[ticker] = make(map[int64]condensedStockCandle)
		}
		for dateInt64 := range stockPrices[ticker] {
			condensedPrices[ticker][dateInt64] = condensedStockCandle{
				// Base data from pulling stock candles from data provider
				Ticker:    stockPrices[ticker][dateInt64].Ticker,
				Close:     stockPrices[ticker][dateInt64].Close,
				Volume:    stockPrices[ticker][dateInt64].Volume,
				Timestamp: stockPrices[ticker][dateInt64].Timestamp,
				// short duration
				AvgVolumeShort:      stockPrices[ticker][dateInt64].AvgVolume30,
				AvgVolumeRatioShort: stockPrices[ticker][dateInt64].AvgVolumeRatio30,
				RVolShort:           stockPrices[ticker][dateInt64].RealizedVolatility30,
				RVolShortVel:        stockPrices[ticker][dateInt64].VelocityRealizedVol30,
				RVolShortAccel:      stockPrices[ticker][dateInt64].RealizedVolAccel30,
				RVolPercentShort:    stockPrices[ticker][dateInt64].RVolPercent30,
				RVolHighShort:       stockPrices[ticker][dateInt64].RVolHigh30,
				RVolLowShort:        stockPrices[ticker][dateInt64].RVolLow30,
				TradeRangeAdj:       stockPrices[ticker][dateInt64].TradeRangeAdj,
				PtradeRangeAdj:      stockPrices[ticker][dateInt64].PTradeRangeAdj,
				// medium duration
				AvgVolumeMed:      stockPrices[ticker][dateInt64].AvgVolume60,
				AvgVolumeRatioMed: stockPrices[ticker][dateInt64].AvgVolumeRatio60,
				RVolMed:           stockPrices[ticker][dateInt64].RealizedVolatility60,
				RVolMedVel:        stockPrices[ticker][dateInt64].VelocityRealizedVol60,
				RVolMedAccel:      stockPrices[ticker][dateInt64].RealizedVolAccel60,
				RVolPercentMed:    stockPrices[ticker][dateInt64].RVolPercent60,
				RVolHighMed:       stockPrices[ticker][dateInt64].RVolHigh60,
				RVolLowMed:        stockPrices[ticker][dateInt64].RVolLow60,
				TrendRangeAdj:     stockPrices[ticker][dateInt64].TrendRangeAdj,
				PTrendRangeAdj:    stockPrices[ticker][dateInt64].PTrendRangeAdj,
				// long duration
				AvgVolumeLong:      stockPrices[ticker][dateInt64].AvgVolume90,
				AvgVolumeRatioLong: stockPrices[ticker][dateInt64].AvgVolumeRatio90,
				RVolLong:           stockPrices[ticker][dateInt64].RealizedVolatility90,
				RVolLongVel:        stockPrices[ticker][dateInt64].VelocityRealizedVol90,
				RVolLongAccel:      stockPrices[ticker][dateInt64].RealizedVolAccel90,
				RVolPercentLong:    stockPrices[ticker][dateInt64].RVolPercent90,
				RVolHighLong:       stockPrices[ticker][dateInt64].RVolHigh90,
				RVolLowLong:        stockPrices[ticker][dateInt64].RVolLow90,
				TailRangeAdj:       stockPrices[ticker][dateInt64].TailRangeAdj,
				PTailRangeAdj:      stockPrices[ticker][dateInt64].PTailRangeAdj,
			}
		}
	}
	return condensedPrices
}

func truncateToDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.UTC)
}

func annualization(ticker string) (daysInYear float64) {
	if strings.HasPrefix(strings.ToUpper(ticker), "X:") {
		return YEAR
	} else {
		return float64(TRADINGDAYSPERYEAR)
	}
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
	//TODO: Implement changes in the function to accept optional "endTimeMilli"
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

func createAlpacaClient(APIKey, APISecretKey string, live bool) (client *alpaca.Client) {
	if live {
		return alpaca.NewClient(alpaca.ClientOpts{
			APIKey:    APIKey,
			APISecret: APISecretKey,
			BaseURL:   ALPACA_LIVE_API,
		})
	} else {
		return alpaca.NewClient(alpaca.ClientOpts{
			APIKey:    APIKey,
			APISecret: APISecretKey,
			BaseURL:   ALPACA_PAPER_API,
		})
	}
}

// GetStockPricesAlpaca retrieves stock prices using Alpaca's stock API. It does NOT gather crypto data using the stock
// api, which is counter to polygon's behavior.
func GetStockPricesAlpaca(clientConfs StockDataConf, ticker, resolution string, startTimeMilli,
	endTimeMilli time.Time, isDebug bool) (stockData map[string]map[int64]SingleStockCandle, err error) {
	var (
		result             map[string]any
		url                string
		feed               = "sip"
		originalTicker     = ticker
		originalResolution = resolution
	)
	if endTimeMilli.Format(time.DateOnly) == time.Now().Format(time.DateOnly) {
		endTimeMilli = endTimeMilli.AddDate(0, 0, -1)
	}
	var startTime = startTimeMilli.Format(time.DateOnly)
	var endTime = endTimeMilli.Format(time.DateOnly)
	stockData = map[string]map[int64]SingleStockCandle{}
	switch resolution {
	case "1T", "1H", "1D", "1W", "1M":
		break
	default:
		fmt.Errorf("invalid resolution format")
		err = errors.New("invalid time resolution format error")
		return nil, err
	}
	//TODO: implement logic to ascertain if the lookup is for crypto or stocks.
	if strings.HasPrefix(ticker, "X:") {
		ticker = strings.Split(ticker, ":")[1]
		cryptoTicker := strings.Replace(ticker, "/", "%2F", 1)
		fmt.Printf("Adjusted ticker is: %s\n", cryptoTicker)
		url = "https://data.alpaca.markets/v1beta3/crypto/us/bars?symbols=" + cryptoTicker + "&timeframe=" +
			resolution + "&start=" + startTime + "&end=" + endTime + "&limit=1000&sort=asc"
	} else {
		url = "https://data.alpaca.markets/v2/stocks/bars?symbols=" + ticker + "&timeframe=" + resolution +
			"&start=" + startTime + "&end=" + endTime + "&limit=1000&adjustment=split&feed=" + feed + "&sort=asc"
	}
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("accept", "application/json")
	req.Header.Add("APCA-API-KEY-ID", clientConfs.AlpacaAPIKey)
	req.Header.Add("APCA-API-SECRET-KEY", clientConfs.AlpacaSecretKey)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Errorf("error retrieving historical stock bars: %e", err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(body, &result)
	if err != nil {
		fmt.Errorf("error unmarshalling stock data: %e", err)
	}

	stockPrices, ok := result["bars"].(map[string]any)
	if !ok || len(stockPrices) == 0 {
		if isDebug {
			fmt.Printf("results do not exist.\n")
		}
		if strings.HasPrefix(originalTicker, "X:") {
			originalTicker = strings.Replace(originalTicker, "/", "", -1)
		}
		switch originalResolution {
		case "1T":
			originalResolution = "minute"
		case "1H":
			originalResolution = "hour"
		case "1W":
			originalResolution = "week"
		case "1M":
			originalResolution = "month"
		case "1Q":
			originalResolution = "quarter"
		case "1Y":
			originalResolution = "year"
		case "1D":
			fallthrough
		default:
			originalResolution = "day"
		}
		stockData, err = GetStockPrices(originalTicker, clientConfs.PolygonAPIToken, originalResolution, startTimeMilli, endTimeMilli)
		if err != nil {
			fmt.Errorf("error pulling stock data from polygon\n")
			return nil, err
		}
	} else {
		for stockSymbol := range stockPrices {
			hb := stockPrices[stockSymbol].([]any)
			if _, ok := stockData[stockSymbol]; !ok {
				stockData[stockSymbol] = map[int64]SingleStockCandle{}
			}
			for _, val := range hb {
				bar := val.(map[string]any)
				timeStamp := bar["t"].(string)
				ts, tsErr := time.Parse(time.RFC3339, timeStamp)
				if tsErr != nil {
					fmt.Errorf("error converting timestamp to time: %w\n", tsErr)
				}
				tsUnixMilli := ts.UnixMilli()
				stockData[stockSymbol][tsUnixMilli] = SingleStockCandle{
					Ticker:         stockSymbol,
					Close:          bar["c"].(float64),
					High:           bar["h"].(float64),
					Low:            bar["l"].(float64),
					Open:           bar["o"].(float64),
					Transactions:   int64(bar["n"].(float64)),
					Timestamp:      ts,
					Volume:         bar["v"].(float64),
					WeightedVolume: bar["vw"].(float64),
				}
			}
		}
	}
	return stockData, nil
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
			Ticker:         ticker,
			Close:          iter.Item().Close,
			High:           iter.Item().High,
			Low:            iter.Item().Low,
			Open:           iter.Item().Open,
			Transactions:   iter.Item().Transactions,
			Timestamp:      time.Time(iter.Item().Timestamp),
			Volume:         iter.Item().Volume,
			WeightedVolume: iter.Item().VWAP,
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
func RealizedVolatility(prices []float64, ticker string) (realizedVol float64) {
	returns := calculateDailyReturn(prices)
	variance := calculateVariance(returns)
	daysInYear := annualization(ticker)
	return math.Sqrt(variance * daysInYear)
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

		if index+SHORTDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= shortDurationStartMilli {
			var volDatesShort []int64
			for shortIndex := index; reverseDateKeys[shortIndex] >= shortDurationStartMilli; shortIndex++ {
				volDatesShort = append(volDatesShort, reverseDateKeys[shortIndex])
			}
			stockCandle.ThirtyDaysPrices, stockCandle.RealizedVolatility30 = calculateVolatility(volDatesShort, stockPrices, ticker)
		}
		if index+MEDIUMDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= medDurationStartMilli {
			var volDatesMed []int64
			for medIndex := index; reverseDateKeys[medIndex] >= medDurationStartMilli; medIndex++ {
				volDatesMed = append(volDatesMed, reverseDateKeys[medIndex])
			}
			stockCandle.SixtyDaysPrices, stockCandle.RealizedVolatility60 = calculateVolatility(volDatesMed, stockPrices, ticker)
		}
		if index+LONGDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= longDurationStartMilli {
			var volDatesLong []int64
			for longIndex := index; reverseDateKeys[longIndex] >= longDurationStartMilli; longIndex++ {
				volDatesLong = append(volDatesLong, reverseDateKeys[longIndex])
			}
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
	realizedVolPeriod := RealizedVolatility(prices, ticker)
	return priceMap, realizedVolPeriod
}

func CalculateRiskRanges(stockPrices map[string]map[int64]SingleStockCandle) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for day := range stockPrices[ticker] {
			dailyTicker := stockPrices[ticker][day]
			if stockPrices[ticker][day].RealizedVolatility30 != 0.0 {
				dailyTicker.TradeRange = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, stockPrices[ticker][day].RealizedVolatility30, SHORTDURATION, ticker)
			}
			if stockPrices[ticker][day].RealizedVolatility60 != 0.0 {
				dailyTicker.TrendRange = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, stockPrices[ticker][day].RealizedVolatility60, MEDIUMDURATION, ticker)
			}
			if stockPrices[ticker][day].RealizedVolatility90 != 0.0 {
				dailyTicker.TailRange = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, stockPrices[ticker][day].RealizedVolatility90, LONGDURATION, ticker)
			}
			stockPrices[ticker][day] = dailyTicker
		}
	}
	return stockPrices
}

func calculateRiskRange(price, volatility, riskRangeDuration float64, ticker string) (riskRange map[string]float64) {
	riskRange = make(map[string]float64)
	daysInYear := annualization(ticker)
	riskRange["high"] = (1 + (volatility / daysInYear * riskRangeDuration)) * price
	riskRange["low"] = (1 - (volatility / daysInYear * riskRangeDuration)) * price
	return
}

func CalculateVelocityOfVolatility(stockPrices map[string]map[int64]SingleStockCandle) (stockPriceMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		var prevDate = int64(0)
		var int64DateArray []int64
		for int64Date := range stockPrices[ticker] {
			int64DateArray = append(int64DateArray, int64Date)
		}

		sort.Slice(int64DateArray, func(i, j int) bool {
			return int64DateArray[i] < int64DateArray[j]
		})

		for _, int64Date := range int64DateArray {
			singleStock := stockPrices[ticker][int64Date]
			if prevDate != 0 {
				singleStock.VelocityRealizedVol30 = stockPrices[ticker][int64Date].RealizedVolatility30 - stockPrices[ticker][prevDate].RealizedVolatility30
				singleStock.VelocityRealizedVol60 = stockPrices[ticker][int64Date].RealizedVolatility60 - stockPrices[ticker][prevDate].RealizedVolatility60
				singleStock.VelocityRealizedVol90 = stockPrices[ticker][int64Date].RealizedVolatility90 - stockPrices[ticker][prevDate].RealizedVolatility90
			}
			stockPrices[ticker][int64Date] = singleStock
			prevDate = int64Date
		}
	}
	stockPriceMap = stockPrices
	return stockPriceMap
}

func CalculateRealizedVolatilityAccel(stockPrices map[string]map[int64]SingleStockCandle) (stockPriceMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		var prevDate = int64(0)
		var int64DateArray []int64
		for int64Date := range stockPrices[ticker] {
			int64DateArray = append(int64DateArray, int64Date)
		}

		sort.Slice(int64DateArray, func(i, j int) bool {
			return int64DateArray[i] < int64DateArray[j]
		})

		for _, int64Date := range int64DateArray {
			singleStock := stockPrices[ticker][int64Date]
			if prevDate != 0 {
				singleStock.RealizedVolAccel30 = stockPrices[ticker][int64Date].VelocityRealizedVol30 - stockPrices[ticker][prevDate].VelocityRealizedVol30
				singleStock.RealizedVolAccel60 = stockPrices[ticker][int64Date].VelocityRealizedVol60 - stockPrices[ticker][prevDate].VelocityRealizedVol60
				singleStock.RealizedVolAccel90 = stockPrices[ticker][int64Date].VelocityRealizedVol90 - stockPrices[ticker][prevDate].VelocityRealizedVol90
			}
			stockPrices[ticker][int64Date] = singleStock
			prevDate = int64Date
		}
	}
	stockPriceMap = stockPrices
	return stockPriceMap
}

func GetAvgVolume(stockPrices map[string]map[int64]SingleStockCandle) (stockData map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
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

			if index+SHORTDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= shortDurationStartMilli {
				var volumesShort []float64
				for shortIndex := index; reverseDateKeys[shortIndex] >= shortDurationStartMilli; shortIndex++ {
					volumesShort = append(volumesShort, stockPrices[ticker][reverseDateKeys[shortIndex]].Volume)
				}
				stockCandle.AvgVolume30 = CalculateAvgVolume(volumesShort)
			}
			if index+MEDIUMDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= medDurationStartMilli {
				var volumesMed []float64
				for medIndex := index; reverseDateKeys[medIndex] >= medDurationStartMilli; medIndex++ {
					volumesMed = append(volumesMed, stockPrices[ticker][reverseDateKeys[medIndex]].Volume)
				}
				stockCandle.AvgVolume60 = CalculateAvgVolume(volumesMed)
			}
			if index+LONGDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= longDurationStartMilli {
				var volumesLong []float64
				for longIndex := index; reverseDateKeys[longIndex] >= longDurationStartMilli; longIndex++ {
					volumesLong = append(volumesLong, stockPrices[ticker][reverseDateKeys[longIndex]].Volume)
				}
				stockCandle.AvgVolume90 = CalculateAvgVolume(volumesLong)
			}
			stockPrices[ticker][date] = stockCandle
		}
	}
	return stockPrices
}

// CalculateAvgVolume computes and returns the average volume across any duration
func CalculateAvgVolume(periodicVolumes []float64) (avgVolume float64) {
	var sum float64
	for _, volume := range periodicVolumes {
		sum += volume
	}
	return sum / float64(len(periodicVolumes))
}

// CalculateAvgVolumeRatios takes the current day's short, medium, and long duration volume averages and compares them
// to the current day's volume to get a ratio for calculating volume adjusted risk ranges
func CalculateAvgVolumeRatios(stockPrices map[string]map[int64]SingleStockCandle) (stockData map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for int64Date := range stockPrices[ticker] {
			stockPriceData := stockPrices[ticker][int64Date]
			if stockPrices[ticker][int64Date].AvgVolume30 != 0.0 {
				stockPriceData.AvgVolumeRatio30 = stockPrices[ticker][int64Date].Volume / stockPrices[ticker][int64Date].AvgVolume30
			}
			if stockPrices[ticker][int64Date].AvgVolume60 != 0.0 {
				stockPriceData.AvgVolumeRatio60 = stockPrices[ticker][int64Date].Volume / stockPrices[ticker][int64Date].AvgVolume60
			}
			if stockPrices[ticker][int64Date].AvgVolume90 != 0.0 {
				stockPriceData.AvgVolumeRatio90 = stockPrices[ticker][int64Date].Volume / stockPrices[ticker][int64Date].AvgVolume90
			}
			stockPrices[ticker][int64Date] = stockPriceData
		}
	}
	return stockPrices
}

func CalculateVolumeAdjustedRiskRanges(stockPrices map[string]map[int64]SingleStockCandle) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for day := range stockPrices[ticker] {
			dailyTicker := stockPrices[ticker][day]
			if stockPrices[ticker][day].RealizedVolatility30 != 0.0 {
				adjVolatility30 := stockPrices[ticker][day].RealizedVolatility30 / stockPrices[ticker][day].AvgVolumeRatio30
				dailyTicker.TradeRangeAdj = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, adjVolatility30, SHORTDURATION, ticker)
			}
			if stockPrices[ticker][day].RealizedVolatility60 != 0.0 {
				adjVolatility60 := stockPrices[ticker][day].RealizedVolatility60 / stockPrices[ticker][day].AvgVolumeRatio60
				dailyTicker.TrendRangeAdj = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, adjVolatility60, MEDIUMDURATION, ticker)
			}
			if stockPrices[ticker][day].RealizedVolatility90 != 0.0 {
				adjVolatility90 := stockPrices[ticker][day].RealizedVolatility90 / stockPrices[ticker][day].AvgVolumeRatio90
				dailyTicker.TailRangeAdj = calculateRiskRange(stockPrices[ticker][day].WeightedVolume, adjVolatility90, LONGDURATION, ticker)
			}
			stockPrices[ticker][day] = dailyTicker
		}
	}
	return stockPrices
}

func CalculateProbabilityAdjRiskRange(riskRange map[string]float64, probabilityAdjustment float64) (probAdjRiskRange map[string]float64) {
	tempRiskRange := map[string]float64{"high": 0.0, "low": 0.0}
	tempRiskRange["high"] = riskRange["high"] - probabilityAdjustment*(riskRange["high"]-riskRange["low"])
	tempRiskRange["low"] = riskRange["low"] + probabilityAdjustment*(riskRange["high"]-riskRange["low"])
	return tempRiskRange
}

func GetProbAdjRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, probabilityAdjustment float64) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	if probabilityAdjustment == 0.0 {
		probabilityAdjustment = .1
	}
	for ticker := range stockPrices {
		for int64Date := range stockPrices[ticker] {
			stockPrice := stockPrices[ticker][int64Date]
			stockPrice.PTradeRange = CalculateProbabilityAdjRiskRange(stockPrice.TradeRange, probabilityAdjustment)
			stockPrice.PTrendRange = CalculateProbabilityAdjRiskRange(stockPrice.TrendRange, probabilityAdjustment)
			stockPrice.PTailRange = CalculateProbabilityAdjRiskRange(stockPrice.TailRange, probabilityAdjustment)
			stockPrice.PTradeRangeAdj = CalculateProbabilityAdjRiskRange(stockPrice.TradeRangeAdj, probabilityAdjustment)
			stockPrice.PTrendRangeAdj = CalculateProbabilityAdjRiskRange(stockPrice.TrendRangeAdj, probabilityAdjustment)
			stockPrice.PTailRangeAdj = CalculateProbabilityAdjRiskRange(stockPrice.TailRangeAdj, probabilityAdjustment)

			stockPrices[ticker][int64Date] = stockPrice
		}
	}
	return stockPrices
}

func GetRelHighLowVol(stockPrices map[string]map[int64]SingleStockCandle) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
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

			if index+SHORTDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= shortDurationStartMilli {
				var (
					shortVolHigh = 0.0
					shortVolLow  = 0.0
				)
				for shortIndex := index; reverseDateKeys[shortIndex] >= shortDurationStartMilli; shortIndex++ {
					if stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30 > shortVolHigh {
						shortVolHigh = stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30
					}
					if stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30 < shortVolLow &&
						stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30 > 0.0 {
						shortVolLow = stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30
					} else if shortVolLow == 0.0 &&
						stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30 > 0.0 {
						shortVolLow = stockPrices[ticker][reverseDateKeys[shortIndex]].RealizedVolatility30
					}
				}
				stockCandle.RVolHigh30 = shortVolHigh
				stockCandle.RVolLow30 = shortVolLow
				shortRVolPercent, err := calculateRVolPercentRange(stockCandle.RVolHigh30, stockCandle.RVolLow30,
					stockCandle.RealizedVolatility30)
				if err != nil {
					fmt.Printf("rVol percent would result in an error. msg:%e\n", err)
				}
				stockCandle.RVolPercent30 = shortRVolPercent
			}
			if index+MEDIUMDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= medDurationStartMilli {
				var (
					medVolHigh = 0.0
					medVolLow  = 0.0
				)
				for medIndex := index; reverseDateKeys[medIndex] >= medDurationStartMilli; medIndex++ {
					if stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60 > medVolHigh {
						medVolHigh = stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60
					}
					if stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60 < medVolLow &&
						stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60 > 0.0 {
						medVolLow = stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60
					} else if medVolLow == 0.0 &&
						stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60 > 0.0 {
						medVolLow = stockPrices[ticker][reverseDateKeys[medIndex]].RealizedVolatility60
					}
				}
				stockCandle.RVolHigh60 = medVolHigh
				stockCandle.RVolLow60 = medVolLow
				medRVolPercent, err := calculateRVolPercentRange(stockCandle.RVolHigh60, stockCandle.RVolLow60,
					stockCandle.RealizedVolatility60)
				if err != nil {
					fmt.Printf("rVol percent would result in an error. msg:%e\n", err)
				}
				stockCandle.RVolPercent60 = medRVolPercent
			}
			if index+LONGDURATION < len(reverseDateKeys)-1 && reverseDateKeys[index] >= longDurationStartMilli {
				var (
					longVolHigh = 0.0
					longVolLow  = 0.0
				)
				for longIndex := index; reverseDateKeys[longIndex] >= longDurationStartMilli; longIndex++ {
					if stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90 > longVolHigh {
						longVolHigh = stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90
					}
					if stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90 < longVolLow &&
						stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90 > 0.0 {
						longVolLow = stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90
					} else if longVolLow == 0.0 &&
						stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90 > 0.0 {
						longVolLow = stockPrices[ticker][reverseDateKeys[longIndex]].RealizedVolatility90
					}
				}
				stockCandle.RVolHigh90 = longVolHigh
				stockCandle.RVolLow90 = longVolLow
				longRVolPercent, err := calculateRVolPercentRange(stockCandle.RVolHigh90, stockCandle.RVolLow90,
					stockCandle.RealizedVolatility90)
				if err != nil {
					fmt.Printf("rVol percent would result in an error. msg:%e\n", err)
				}
				stockCandle.RVolPercent90 = longRVolPercent
			}
			stockPrices[ticker][date] = stockCandle
		}
	}
	return stockPrices
}

func calculateRVolPercentRange(rVolHigh, rVolLow, rVol float64) (rvolPercent float64, err error) {
	if rVolHigh-rVolLow == 0.0 {
		err = errors.New("divide by zero error")
		return 0.0, err
	}
	return (rVol - rVolLow) / (rVolHigh - rVolLow), nil
}
