package pkg

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

// NormalizeTicker upper-cases and trims a ticker string so the CLI accepts
// any mix of upper, lower, or mixed case input.
func NormalizeTicker(s string) string {
	return strings.ToUpper(strings.TrimSpace(s))
}

const DAY = 24
const YEAR = 365.24
const LONGDURATION = 180
const MEDIUMDURATION = 90
const SHORTDURATION = 30
const TRADINGDAYSPERYEAR = 252
const ALPACA_PAPER_API = "https://paper-api.alpaca.markets"
const ALPACA_LIVE_API = "https://api.alpaca.markets"

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
		log.Printf("error marshalling data into JSON string: %v", err)
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
				Ticker:            stockPrices[ticker][dateInt64].Ticker,
				Close:             stockPrices[ticker][dateInt64].Close,
				Volume:            stockPrices[ticker][dateInt64].Volume,
				PriceVelocity:     stockPrices[ticker][dateInt64].PriceVelocity,
				PriceAcceleration: stockPrices[ticker][dateInt64].PriceAccel,

				Timestamp: stockPrices[ticker][dateInt64].Timestamp,
				// short duration
				AvgVolumeShort:      stockPrices[ticker][dateInt64].AvgVolumeShort,
				AvgVolumeRatioShort: stockPrices[ticker][dateInt64].AvgVolumeRatioShort,
				TradeSlope:          stockPrices[ticker][dateInt64].SlopeShortDuration,
				RVolShort:           stockPrices[ticker][dateInt64].RealizedVolatilityShort,
				RVolShortVel:        stockPrices[ticker][dateInt64].VelocityRealizedVolShort,
				RVolShortAccel:      stockPrices[ticker][dateInt64].RealizedVolAccelShort,
				RVolPercentShort:    stockPrices[ticker][dateInt64].RVolPercentShort,
				RVolHighShort:       stockPrices[ticker][dateInt64].RVolHighShort,
				RVolLowShort:        stockPrices[ticker][dateInt64].RVolLowShort,
				TradeRangeAdj:       stockPrices[ticker][dateInt64].TradeRangeAdj,
				PtradeRangeAdj:      stockPrices[ticker][dateInt64].PTradeRangeAdj,
				// medium duration
				AvgVolumeMed:      stockPrices[ticker][dateInt64].AvgVolumeMed,
				AvgVolumeRatioMed: stockPrices[ticker][dateInt64].AvgVolumeRatioMed,
				TrendSlope:        stockPrices[ticker][dateInt64].SlopeMedDuration,
				RVolMed:           stockPrices[ticker][dateInt64].RealizedVolatilityMed,
				RVolMedVel:        stockPrices[ticker][dateInt64].VelocityRealizedVolMed,
				RVolMedAccel:      stockPrices[ticker][dateInt64].RealizedVolAccelMed,
				RVolPercentMed:    stockPrices[ticker][dateInt64].RVolPercentMed,
				RVolHighMed:       stockPrices[ticker][dateInt64].RVolHighMed,
				RVolLowMed:        stockPrices[ticker][dateInt64].RVolLowMed,
				TrendRangeAdj:     stockPrices[ticker][dateInt64].TrendRangeAdj,
				PTrendRangeAdj:    stockPrices[ticker][dateInt64].PTrendRangeAdj,
				// long duration
				AvgVolumeLong:      stockPrices[ticker][dateInt64].AvgVolumeLong,
				AvgVolumeRatioLong: stockPrices[ticker][dateInt64].AvgVolumeRatioLong,
				TailSlope:          stockPrices[ticker][dateInt64].SlopeLongDuration,
				RVolLong:           stockPrices[ticker][dateInt64].RealizedVolatilityLong,
				RVolLongVel:        stockPrices[ticker][dateInt64].VelocityRealizedVolLong,
				RVolLongAccel:      stockPrices[ticker][dateInt64].RealizedVolAccelLong,
				RVolPercentLong:    stockPrices[ticker][dateInt64].RVolPercentLong,
				RVolHighLong:       stockPrices[ticker][dateInt64].RVolHighLong,
				RVolLowLong:        stockPrices[ticker][dateInt64].RVolLowLong,
				TailRangeAdj:       stockPrices[ticker][dateInt64].TailRangeAdj,
				PTailRangeAdj:      stockPrices[ticker][dateInt64].PTailRangeAdj,
				TradeDirection:     stockPrices[ticker][dateInt64].TradeDirection,
				TrendDirection:     stockPrices[ticker][dateInt64].TrendDirection,
				TailDirection:      stockPrices[ticker][dateInt64].TailDirection,
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
	if endTimeMilli.Format(time.DateOnly) == time.Now().Format(time.DateOnly) && !strings.HasPrefix(ticker, "X:") {
		endTimeMilli = endTimeMilli.AddDate(0, 0, -1)
	}
	var startTime = startTimeMilli.Format(time.DateOnly)
	var endTime = endTimeMilli.Format(time.DateOnly)
	stockData = map[string]map[int64]SingleStockCandle{}
	switch resolution {
	case "1T", "1H", "1D", "1W", "1M":
		break
	default:
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
		log.Printf("error retrieving historical stock bars: %v", err)
	}

	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)

	err = json.Unmarshal(body, &result)
	if err != nil {
		log.Printf("error unmarshalling stock data: %v", err)
	}

	stockPrices, ok := result["bars"].(map[string]any)
	if !ok || len(stockPrices) == 0 {
		if isDebug {
			fmt.Printf("results do not exist.\n")
		}
		if strings.HasPrefix(originalTicker, "X:") {
			originalTicker = strings.ReplaceAll(originalTicker, "/", "")
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
			log.Printf("error pulling stock data from polygon: %v", err)
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
					log.Printf("error converting timestamp to time: %v", tsErr)
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
			Ticker:         strings.ReplaceAll(ticker, "X:", ""),
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

// calculateMean calculates the arithmetic mean of a float64 slice of inputs and returns the resulting mean

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

func StoreRealizedVols(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPriceData map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		var dateKeys []int64
		for dateKey := range stockPrices[ticker] {
			dateKeys = append(dateKeys, dateKey)
		}
		reverseDateKeys := dateKeys
		sort.Slice(reverseDateKeys, func(i, j int) bool {
			return reverseDateKeys[i] > reverseDateKeys[j]
		})
		for index, date := range reverseDateKeys {
			stockCandle := stockPrices[ticker][date]
			windowDates, ok := collectWindowDates(reverseDateKeys, index, duration)
			if ok {
				prices, vol := calculateVolatility(windowDates, stockPrices, ticker)
				setPrices(&stockCandle, duration, prices)
				setRVol(&stockCandle, duration, vol)
			}
			stockPrices[ticker][date] = stockCandle
		}
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

func CalculateRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for day := range stockPrices[ticker] {
			dailyTicker := stockPrices[ticker][day]
			if rv := getRVol(stockPrices[ticker][day], duration); rv != 0.0 {
				setRiskRange(&dailyTicker, duration,
					calculateRiskRange(stockPrices[ticker][day].WeightedVolume, rv, float64(duration), ticker))
			}
			stockPrices[ticker][day] = dailyTicker
		}
	}
	return stockPrices
}

// todo: add test for riskRange["low"]
func calculateRiskRange(price, volatility, riskRangeDuration float64, ticker string) (riskRange map[string]float64) {
	riskRange = make(map[string]float64)
	daysInYear := annualization(ticker)
	riskRange["high"] = (1 + (volatility / daysInYear * riskRangeDuration)) * price
	rrlow := (1 - (volatility / daysInYear * riskRangeDuration)) * price
	if rrlow < 0 {
		rrlow = 0
	}
	riskRange["low"] = rrlow
	return
}

func CalculateVelocities(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPriceMap map[string]map[int64]SingleStockCandle) {
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
				setRVolVel(&singleStock, duration,
					getRVol(stockPrices[ticker][int64Date], duration)-getRVol(stockPrices[ticker][prevDate], duration))
				singleStock.PriceVelocity = stockPrices[ticker][int64Date].Close - stockPrices[ticker][prevDate].Close
			}
			stockPrices[ticker][int64Date] = singleStock
			prevDate = int64Date
		}
	}
	stockPriceMap = stockPrices
	return stockPriceMap
}

func CalculateAccelerations(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPriceMap map[string]map[int64]SingleStockCandle) {
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
				setRVolAccel(&singleStock, duration,
					getRVolVel(stockPrices[ticker][int64Date], duration)-getRVolVel(stockPrices[ticker][prevDate], duration))
				singleStock.PriceAccel = stockPrices[ticker][int64Date].PriceVelocity - stockPrices[ticker][prevDate].PriceVelocity
			}
			stockPrices[ticker][int64Date] = singleStock
			prevDate = int64Date
		}
	}
	stockPriceMap = stockPrices
	return stockPriceMap
}

func GetAvgVolume(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockData map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		var dateKeys []int64
		for dateKey := range stockPrices[ticker] {
			dateKeys = append(dateKeys, dateKey)
		}
		reverseDateKeys := dateKeys
		sort.Slice(reverseDateKeys, func(i, j int) bool {
			return reverseDateKeys[i] > reverseDateKeys[j]
		})
		for index, date := range reverseDateKeys {
			stockCandle := stockPrices[ticker][date]
			windowDates, ok := collectWindowDates(reverseDateKeys, index, duration)
			if ok {
				var volumes []float64
				for _, wd := range windowDates {
					volumes = append(volumes, stockPrices[ticker][wd].Volume)
				}
				setAvgVol(&stockCandle, duration, CalculateAvgVolume(volumes))
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
func CalculateAvgVolumeRatios(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockData map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for int64Date := range stockPrices[ticker] {
			stockPriceData := stockPrices[ticker][int64Date]
			if getAvgVol(stockPrices[ticker][int64Date], duration) != 0.0 {
				setAvgVolRatio(&stockPriceData, duration,
					stockPrices[ticker][int64Date].Volume/getAvgVol(stockPrices[ticker][int64Date], duration))
			}
			stockPrices[ticker][int64Date] = stockPriceData
		}
	}
	return stockPrices
}

func CalculateVolumeAdjustedRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for day := range stockPrices[ticker] {
			dailyTicker := stockPrices[ticker][day]
			rv := getRVol(stockPrices[ticker][day], duration)
			ratio := getAvgVolRatio(stockPrices[ticker][day], duration)
			if rv != 0.0 {
				adjVol := rv / ratio
				setAdjRiskRange(&dailyTicker, duration,
					calculateRiskRange(stockPrices[ticker][day].WeightedVolume, adjVol, float64(duration), ticker))
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

func GetProbAdjRiskRanges(stockPrices map[string]map[int64]SingleStockCandle, duration int, probabilityAdjustment float64) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	if probabilityAdjustment == 0.0 {
		probabilityAdjustment = .1
	}
	for ticker := range stockPrices {
		for int64Date := range stockPrices[ticker] {
			stockPrice := stockPrices[ticker][int64Date]
			setProbRiskRange(&stockPrice, duration,
				CalculateProbabilityAdjRiskRange(getRiskRange(stockPrice, duration), probabilityAdjustment))
			setProbAdjRiskRange(&stockPrice, duration,
				CalculateProbabilityAdjRiskRange(getAdjRiskRange(stockPrice, duration), probabilityAdjustment))
			stockPrices[ticker][int64Date] = stockPrice
		}
	}
	return stockPrices
}

func GetRelHighLowVol(stockPrices map[string]map[int64]SingleStockCandle, duration int) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		var dateKeys []int64
		for dateKey := range stockPrices[ticker] {
			dateKeys = append(dateKeys, dateKey)
		}
		reverseDateKeys := dateKeys
		sort.Slice(reverseDateKeys, func(i, j int) bool {
			return reverseDateKeys[i] > reverseDateKeys[j]
		})
		for index, date := range reverseDateKeys {
			stockCandle := stockPrices[ticker][date]
			windowDates, ok := collectWindowDates(reverseDateKeys, index, duration)
			if ok {
				high := 0.0
				low := 0.0
				for _, wd := range windowDates {
					rv := getRVol(stockPrices[ticker][wd], duration)
					if rv > high {
						high = rv
					}
					if rv > 0.0 && (low == 0.0 || rv < low) {
						low = rv
					}
				}
				setRVolHigh(&stockCandle, duration, high)
				setRVolLow(&stockCandle, duration, low)
				pct, err := calculateRVolPercentRange(high, low, getRVol(stockCandle, duration))
				if err != nil {
					fmt.Printf("rVol percent would result in an error. msg:%e\n", err)
				}
				setRVolPercent(&stockCandle, duration, pct)
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

func GetLinearRegressionSlope(stockPrices map[string]map[int64]SingleStockCandle, duration int, isDebug bool) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	for ticker := range stockPrices {
		for dateInt64 := range stockPrices[ticker] {
			singleTickerData := stockPrices[ticker][dateInt64]
			prices := getPrices(singleTickerData, duration)
			if len(prices) > 0 {
				var xVals []float64
				var yVals []float64
				i := 0
				for dateString := range prices {
					yVals = append(yVals, prices[dateString])
					i++
					xVals = append(xVals, float64(i))
				}
				slope, intercept, err := calcLinearRegression(xVals, yVals)
				if err != nil {
					log.Printf("error getting linear regression: %v", err)
				} else {
					if isDebug {
						fmt.Printf("Date: %s duration: %d intercept: %f\n",
							stockPrices[ticker][dateInt64].Timestamp, duration, intercept)
					}
					setSlope(&singleTickerData, duration, slope)
				}
			} else {
				setSlope(&singleTickerData, duration, 0.0)
			}
			stockPrices[ticker][dateInt64] = singleTickerData
		}
	}
	return stockPrices
}

func calcLinearRegression(xValues, yValues []float64) (slope, intercept float64, err error) {
	if len(xValues) != len(yValues) || len(xValues) < 2 {
		return 0, 0, errors.New("invalid input: x and y slices must have the same length and at least 2 data points")
	}

	n := float64(len(xValues))
	var sumX, sumY, sumXY, sumX2 float64

	// Calculate sums
	for i := 0; i < len(xValues); i++ {
		sumX += xValues[i]
		sumY += yValues[i]
		sumXY += xValues[i] * yValues[i]
		sumX2 += xValues[i] * xValues[i]
	}

	// Calculate slope (m)
	slope = (n*sumXY - sumX*sumY) / (n*sumX2 - sumX*sumX)

	// Calculate y-intercept (b)
	intercept = (sumY - slope*sumX) / n

	return slope, intercept, nil
}

// GetSimpleSlopes computes the raw price delta for each duration per day.
// For each day it looks back N calendar days, rolling back one day at a time
// until finding a trading day at or before the target, then computes:
//
//	slope = close_today - close_at_lookback_date
//
// SlopeXxxValid is set true only when a lookback date was found; false means
// insufficient history and the slope value of 0.0 is meaningless.
func GetSimpleSlopes(stockPrices map[string]map[int64]SingleStockCandle, isDebug bool) (stockPricesMap map[string]map[int64]SingleStockCandle) {
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
			currentClose := stockCandle.Close

			shortTarget := time.UnixMilli(currentDate).AddDate(0, 0, -SHORTDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= shortTarget {
					stockCandle.SlopeShortDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeShortValid = true
					if isDebug {
						fmt.Printf("ticker=%s date=%s shortSlope=%.4f\n",
							ticker, time.UnixMilli(currentDate).Format(time.DateOnly), stockCandle.SlopeShortDuration)
					}
					break
				}
			}

			medTarget := time.UnixMilli(currentDate).AddDate(0, 0, -MEDIUMDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= medTarget {
					stockCandle.SlopeMedDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeMedValid = true
					if isDebug {
						fmt.Printf("ticker=%s date=%s medSlope=%.4f\n",
							ticker, time.UnixMilli(currentDate).Format(time.DateOnly), stockCandle.SlopeMedDuration)
					}
					break
				}
			}

			longTarget := time.UnixMilli(currentDate).AddDate(0, 0, -LONGDURATION).UnixMilli()
			for _, pastDate := range dateKeys {
				if pastDate <= longTarget {
					stockCandle.SlopeLongDuration = currentClose - stockPrices[ticker][pastDate].Close
					stockCandle.SlopeLongValid = true
					if isDebug {
						fmt.Printf("ticker=%s date=%s longSlope=%.4f\n",
							ticker, time.UnixMilli(currentDate).Format(time.DateOnly), stockCandle.SlopeLongDuration)
					}
					break
				}
			}

			stockPrices[ticker][currentDate] = stockCandle
		}
	}
	return stockPrices
}

// CalculateTrendDirections assigns TradeDirection, TrendDirection, and TailDirection
// to each day based on whether the three most recent consecutive slopes (today,
// yesterday, day-before) are all positive (Bullish), all negative (Bearish),
// mixed (Neutral), or unavailable (Indeterminate).
//
// Must be called after GetSimpleSlopes so that validity flags are set.
func CalculateTrendDirections(stockPrices map[string]map[int64]SingleStockCandle, isDebug bool) (stockPricesMap map[string]map[int64]SingleStockCandle) {
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
				if isDebug {
					fmt.Printf("ticker=%s date=%d tradeDir=%s trendDir=%s tailDir=%s\n",
						ticker, currentDate,
						stockCandle.TradeDirection, stockCandle.TrendDirection, stockCandle.TailDirection)
				}
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

// CalculateTrend takes the end date and returns the difference between the input date's close price and the close price
// of the date that is back a number of days prior to the input date
func CalculateTrend(stockPrices map[string]map[int64]SingleStockCandle, ticker string, endDate time.Time, tickerDuration int64) (stockPricesMap map[string]map[int64]SingleStockCandle) {
	/*
	* Take the Close price from the input date and subtract a number of days equal to the duration. If a date does not
	* exist, but is not prior to the start date, subtract another day and try again until hitting a number that exists.
	 */
	var trendInt float64
	endDateMilli := endDate.UnixMilli()
	startDateMilli := endDateMilli - time.Hour.Milliseconds()*DAY*tickerDuration

	trendInt = stockPrices[ticker][endDateMilli].Close - stockPrices[ticker][startDateMilli].Close

	trendRatio := trendInt / stockPrices[ticker][endDateMilli].Close
	stockPrice := stockPrices[ticker][endDateMilli]

	switch trendRatio {
	case SHORTDURATION:

		stockPrice.SlopeShortDuration = trendRatio
	case MEDIUMDURATION:
		stockPrice.SlopeMedDuration = trendRatio
	case LONGDURATION:
		stockPrice.SlopeLongDuration = trendRatio
	}
	stockPrices[ticker][endDateMilli] = stockPrice

	return stockPrices
}
