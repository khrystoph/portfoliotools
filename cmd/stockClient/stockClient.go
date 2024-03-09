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
)

func init() {
	flag.StringVar(&tickerConfig, "config", ".stockclientconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			".stockclientconfig.json")
	flag.StringVar(&ticker, "ticker", "AAPL", "Enter stock ticker to look up price info for")
	flag.StringVar(&startTime, "startTime", "30 days ago", "Enter a time to start gathering data "+
		"for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&endTime, "endTime", "Today",
		"Enter a time to end gathering data for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ."+
			" Time will always assume UTC.")
	flag.StringVar(&resolution, "resolution", "day", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year."+
		" The numeric time values represent minutes. Default resolution: day.")
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
		fmt.Errorf("error reading user's homedir: %w", err)
	}
	tickerConfig = strings.Replace(tickerConfig, "~", userDir, 1)
	if _, err = os.Stat(tickerConfig); errors.Is(err, os.ErrNotExist) {
		fmt.Errorf("error: config file %s does not exist. exiting", tickerConfig)
	}
	configFile, err := os.Open(tickerConfig)
	if err != nil {
		fmt.Errorf("error opening the config file. %w", err)
	}
	defer configFile.Close()
	configDecoder := json.NewDecoder(configFile)
	stockDataConfig := pkg.StockDataConf{}
	err = configDecoder.Decode(&stockDataConfig)
	if err != nil {
		fmt.Errorf("error decoding the json config file. exiting. error msg: %w", err)
	}

	startTimeMilli, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		fmt.Printf("Unable to convert startTime to milliseconds.\nstartTime: %s\n", startTime)
		return
	}
	endTimeMilli, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		fmt.Printf("Unable to convert endTime to milliseconds.\n")
		return
	}
	// retrieve stock ticker's prices and store in a map

	if stockDataConfig.AlpacaSecretKey != "" {
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
		tickerData, err = pkg.GetStockPricesAlpaca(stockDataConfig, ticker, resolution, startTimeMilli, endTimeMilli)
		if err != nil {
			fmt.Errorf("unable to retrieve stock data: %e", err)
		}
	} else {
		tickerData, err = pkg.GetStockPrices(strings.ToUpper(ticker), stockDataConfig.PolygonAPIToken, resolution, startTimeMilli, endTimeMilli)
		if err != nil {
			fmt.Errorf("unable to get stock prices")
		}
	}

	// Call functions to calculate each day's realized volatility, ranges, and adjusted ranges given each duration available (30, 60, 90)
	tickerData = pkg.StoreRealizedVols(tickerData, strings.ToUpper(ticker))
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
	jsonTickerData, err := json.MarshalIndent(tickerData[strings.ToUpper(ticker)], "", "  ")
	if err != nil {
		fmt.Errorf("error marshalling data into JSON string")
	}
	fmt.Printf("%v\n", string(jsonTickerData))
}
