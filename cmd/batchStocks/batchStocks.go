package main

import (
	"cmd/pkg"
	"encoding/csv"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

var (
	csvFile, outFile, tickerConfig, batchStockRangesFile, timeDuration string
	debug, excelOut, noEmail, showTail                                 bool
)

func init() {
	flag.StringVar(&tickerConfig, "config", ".stockclientconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			".stockclientconfig.json")
	flag.StringVar(&tickerConfig, "c", ".stockclientconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			".stockclientconfig.json")
	flag.BoolVar(&debug, "debug", false, "Toggles debug output for purposes"+
		" of showing more information. Default value: false.")
	flag.BoolVar(&debug, "d", false, "Toggles debug output for purposes"+
		" of showing more information. Default value: false.")
	flag.BoolVar(&noEmail, "n", false, "Toggles whether to send the email or not. For debug purposes.")
	flag.BoolVar(&noEmail, "noemail", false, "Toggles whether to send the email or not. For debug purposes.")
	flag.StringVar(&csvFile, "f", "tickers.csv", "path to csv file of tickers in format: ABC,X:DEF,ghi,x:jkl")
	flag.StringVar(&csvFile, "file", "tickers.csv", "path to csv file of tickers in format: ABC,X:DEF,ghi,x:jkl")
	flag.StringVar(&outFile, "outFile", "tickers_out.json", "output file for ")
	flag.StringVar(&batchStockRangesFile, "o", "stockRanges.json",
		"Set the value of the output file of the batch stock ranges.json")
	flag.StringVar(&batchStockRangesFile, "outfile", "stockRanges.json",
		"Set the value of the output file of the batch stock ranges.json")
	flag.BoolVar(&excelOut, "x", false, "Writes a file in excel format using same outfile name as -o "+
		"except it swaps the file type")
	flag.BoolVar(&excelOut, "excelfmt", false, "Writes a file in excel format using same outfile name as -o "+
		"except it swaps the file type")
	flag.StringVar(&timeDuration, "t", "SHORT",
		"give the duration in terms of number of candles, such as SHORT for the short-term trend duration")
	flag.BoolVar(&showTail, "tail", false, "Include Tail Slope and Tail Dir columns in Excel output")
	flag.BoolVar(&showTail, "tail-cols", false, "Include Tail Slope and Tail Dir columns in Excel output")
}

func main() {
	flag.Parse()
	var (
		tickerData          = make(map[string]map[int64]pkg.SingleStockCandle)
		tickerBatch         = make(map[string]map[int64]pkg.SingleStockCandle)
		batchStockRanges    = make(map[string]pkg.CondensedRangesJSON)
		tickerArray         = make([]string, 0)
		err                 error
		userDir, resolution string
	)

	// Section parses the config file location, opens it, decodes the JSON and loads the API creds
	userDir, err = os.UserHomeDir()
	if err != nil {
		log.Printf("error reading user's homedir: %v", err)
	}
	tickerConfig = strings.Replace(tickerConfig, "~", userDir, 1)
	if _, err = os.Stat(tickerConfig); errors.Is(err, os.ErrNotExist) {
		log.Printf("error: config file %s does not exist. exiting", tickerConfig)
	}

	configFile, err := os.Open(tickerConfig)
	defer configFile.Close()
	if err != nil {
		log.Printf("error opening the config file: %v", err)
	}
	configDecoder := json.NewDecoder(configFile)
	stockDataConfig := pkg.StockDataConf{}
	err = configDecoder.Decode(&stockDataConfig)
	if err != nil {
		log.Printf("error decoding the json config file: %v", err)
		os.Exit(1)
	}

	// Section uses today's date in milliseconds, then subtracts a year for the start date for simplicity
	// TODO: make the start and end dates configurable
	endDate := time.Now()
	endDateMilli := endDate.UnixMilli()

	startDateMilli := time.UnixMilli(endDateMilli).AddDate(-1, 0, 0)

	// Section parses the list of tickers and then loops over them to create a single slice of stocks to iterate over
	file, err := os.Open(csvFile)
	if err != nil {
		log.Fatal(err)
	}
	stockReader := csv.NewReader(file)
	csvItem, err := stockReader.ReadAll()
	for _, row := range csvItem {
		log.Printf("row: %v", row)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}
		for _, ticker := range row {
			tickerArray = append(tickerArray, ticker)
		}
	}

	for _, tickerItem := range tickerArray {
		isCrypto := false
		if strings.HasPrefix(strings.ToUpper(tickerItem), "X:") {
			isCrypto = true
		}
		if stockDataConfig.AlpacaAPIKey != "" {
			switch resolution {
			case "minute", "Minute", "MINUTE", "M", "m":
				resolution = "1T"
			case "hour", "Hour", "HOUR", "H", "h":
				resolution = "1H"
			case "Week", "week", "WEEK", "W", "w":
				resolution = "1W"
			case "Month", "month", "MONTH", "Mo", "mo":
				resolution = "1M"
			case "DAY", "day", "Day", "D", "d":
				fallthrough
			default:
				resolution = "1D"
			}

			tickerData, err = pkg.GetStockPricesAlpaca(stockDataConfig, strings.ToUpper(tickerItem), resolution, startDateMilli, endDate, debug)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			tickerData, err = pkg.GetStockPrices(strings.ToUpper(tickerItem), stockDataConfig.PolygonAPIToken, resolution, startDateMilli, endDate)
			if err != nil {
				log.Printf("unable to get stock prices for %s", strings.ToUpper(tickerItem))
			}
		}

		// Call functions to calculate each day's realized volatility, ranges, and adjusted ranges given each duration available (30, 60, 90)
		tickerData = pkg.StoreRealizedVols(tickerData)
		tickerData = pkg.GetAvgVolume(tickerData)
		tickerData = pkg.CalculateAvgVolumeRatios(tickerData)
		tickerData = pkg.GetRelHighLowVol(tickerData)
		tickerData = pkg.CalculateRiskRanges(tickerData)
		tickerData = pkg.CalculateVolumeAdjustedRiskRanges(tickerData)
		tickerData = pkg.CalculateVelocities(tickerData)
		tickerData = pkg.CalculateAccelerations(tickerData)
		tickerData = pkg.GetProbAdjRiskRanges(tickerData, stockDataConfig.RangeAdjustment)
		tickerData = pkg.GetSimpleSlopes(tickerData, debug)
		tickerData = pkg.CalculateTrendDirections(tickerData, debug)
		//tickerData = pkg.GetLinearRegressionSlope(tickerData, debug)
		tickerStripped := strings.ToUpper(tickerItem)
		if strings.HasPrefix(tickerStripped, "X:") {
			tickerStripped = strings.Split(tickerStripped, ":")[1]
		}
		if _, ok := tickerData[tickerStripped]; !ok {
			tickerBatch[tickerStripped] = make(map[int64]pkg.SingleStockCandle)
		}

		for ticker, stock := range tickerData {
			latestDate := int64(0)
			var rrHigh, rrLow, rvolpct, avgvolratio float64
			for date, _ := range tickerData[ticker] {
				// Looking for the "max" date to get the most recent datetime
				if date > latestDate {
					latestDate = date
				}
			}
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
		}
	}

	jsonData, err := json.MarshalIndent(batchStockRanges, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	err = os.WriteFile(batchStockRangesFile, jsonData, 0600)
	if err != nil {
		log.Fatal(err)
	}

	if excelOut {
		excelOutFile := strings.Split(batchStockRangesFile, ".")[0] + ".xlsx"
		pkg.GenerateStockReportXLSX(batchStockRanges, excelOutFile, showTail)
	}

	if !noEmail {
		err = pkg.SendEmail(stockDataConfig, batchStockRangesFile, excelOut)
		if err != nil {
			log.Fatal(err)
		}
	}
	return
}
