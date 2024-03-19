package main

import (
	"cmd/pkg"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"
)

var (
	ticker, startTime, endTime, resolution, tickerConfig string
	debug                                                bool
)

func init() {
	flag.StringVar(&tickerConfig, "config", ".stockclientconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			".stockclientconfig.json")
	flag.StringVar(&tickerConfig, "c", ".stockclientconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			".stockclientconfig.json")
	flag.StringVar(&ticker, "ticker", "AAPL", "Enter stock ticker to look up price info for")
	flag.StringVar(&ticker, "t", "AAPL", "Enter stock ticker to look up price info for")
	flag.StringVar(&startTime, "startTime", "30 days ago", "Enter a time to start gathering data "+
		"for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&startTime, "s", "30 days ago", "Enter a time to start gathering data "+
		"for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&endTime, "endTime", "Today",
		"Enter a time to end gathering data for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ."+
			" Time will always assume UTC.")
	flag.StringVar(&endTime, "e", "Today",
		"Enter a time to end gathering data for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ."+
			" Time will always assume UTC.")
	flag.StringVar(&resolution, "resolution", "day", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year."+
		" The numeric time values represent minutes. Default resolution: day.")
	flag.StringVar(&resolution, "r", "day", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year."+
		" The numeric time values represent minutes. Default resolution: day.")
	flag.BoolVar(&debug, "debug", false, "Toggles debug output for purposes"+
		" of showing more information. Default value: false.")
	flag.BoolVar(&debug, "d", false, "Toggles debug output for purposes"+
		" of showing more information. Default value: false.")
}

func main() {
	flag.Parse()
	var (
		tickerData map[string]map[int64]pkg.SingleStockCandle
		err        error
		userDir    string
	)

	if endTime == "Today" {
		endTime = time.Now().Format(time.RFC3339)
	}

	// Section parses the config file location, opens it, decodes the JSON and loads the API creds
	userDir, err = os.UserHomeDir()
	if err != nil {
		_ = fmt.Errorf("error reading user's homedir: %w", err)
	}
	tickerConfig = strings.Replace(tickerConfig, "~", userDir, 1)
	if _, err = os.Stat(tickerConfig); errors.Is(err, os.ErrNotExist) {
		_ = fmt.Errorf("error: config file %s does not exist. exiting", tickerConfig)
	}
	configFile, err := os.Open(tickerConfig)
	if err != nil {
		_ = fmt.Errorf("error opening the config file. %w", err)
	}
	defer configFile.Close()
	configDecoder := json.NewDecoder(configFile)
	stockDataConfig := pkg.StockDataConf{}
	err = configDecoder.Decode(&stockDataConfig)
	if err != nil {
		_ = fmt.Errorf("error decoding the json config file. exiting. error msg: %w", err)
		os.Exit(1)
	}

	startTimeMilli, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		_ = fmt.Errorf("Unable to convert startTime to milliseconds.\nstartTime: %s\n", startTime)
		os.Exit(1)
	}
	endTimeMilli, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		_ = fmt.Errorf("Unable to convert endTime to milliseconds.\n")
		os.Exit(1)
	}
	// retrieve stock ticker's prices and store in a map

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
		tickerData, err = pkg.GetStockPricesAlpaca(stockDataConfig, strings.ToUpper(ticker), resolution, startTimeMilli, endTimeMilli, debug)
		if err != nil {
			fmt.Errorf("unable to retrieve stock data: %e", err)
		}
		if strings.HasPrefix(strings.ToUpper(ticker), "X:") {
			ticker = strings.Split(ticker, ":")[1]
			fmt.Printf("ticker: %s\n", ticker)
		}
	} else {
		tickerData, err = pkg.GetStockPrices(strings.ToUpper(ticker), stockDataConfig.PolygonAPIToken, resolution, startTimeMilli, endTimeMilli)
		if err != nil {
			fmt.Errorf("unable to get stock prices")
		}
	}

	// Call functions to calculate each day's realized volatility, ranges, and adjusted ranges given each duration available (30, 60, 90)
	tickerData = pkg.StoreRealizedVols(tickerData, strings.ToUpper(ticker))
	tickerData = pkg.GetRelHighLowVol(tickerData)
	tickerData = pkg.GetAvgVolume(tickerData)
	tickerData = pkg.CalculateAvgVolumeRatios(tickerData)
	tickerData = pkg.CalculateRiskRanges(tickerData)
	tickerData = pkg.CalculateVolumeAdjustedRiskRanges(tickerData)
	tickerData = pkg.CalculateVelocityOfVolatility(tickerData)
	tickerData = pkg.CalculateRealizedVolatilityAccel(tickerData)
	tickerData = pkg.GetProbAdjRiskRanges(tickerData, stockDataConfig.RangeAdjustment)

	if err != nil {
		fmt.Errorf("error occurred: %w", err)
	}
	pkg.PrintData(tickerData, debug)
}
