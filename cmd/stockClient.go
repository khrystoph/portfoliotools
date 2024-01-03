package main

import (
	"cmd/pkg"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"google.golang.org/appengine/log"
	"os"
	"strings"
	"time"
)

var (
	ticker, startTime, endTime, resolution, tickerConfig string
	costBasis, currPrice, targetAnnualizedRate           float64
	short                                                bool
)

func init() {
	flag.StringVar(&tickerConfig, "config", "~/.polygon/polygonconfig.json",
		"path to the json config file containing credentials for ticker data. Default is: "+
			"~/.polygon/polygonconfig.json")
	flag.StringVar(&ticker, "ticker", "AAPL", "Enter stock ticker to look up price info for")
	flag.StringVar(&startTime, "startTime", "30 days ago", "Enter a time to start gathering data "+
		"for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&endTime, "endTime", "Today",
		"Enter a time to end gathering data for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ."+
			" Time will always assume UTC.")
	flag.StringVar(&resolution, "resolution", "day", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year."+
		" The numeric time values represent minutes. Default resolution: day.")
	flag.Float64Var(&costBasis, "costBasis", 1, "input the cost basis in decimal form. "+
		"Example: 12.34")
	flag.Float64Var(&currPrice, "currentPrice", 1, "input the current price in decimal form. "+
		"Example: 12.34")
	flag.Float64Var(&targetAnnualizedRate, "targetRate", .06,
		"enter either the risk-free rate or the rate you want as your target return rate. Default is: .06 (6%).")
	flag.BoolVar(&short, "short", false, "default: False. Presence of the flag means true.")
}

type StockDataConf struct {
	Creds string
}

func main() {
	flag.Parse()
	var (
		tickerData                                                    map[string]map[int64]pkg.SingleStockCandle
		annualizedReturnsMode                                         = false
		targetAnnualizedReturnsMode                                   = false
		err                                                           error
		userDir                                                       string
		periodRealizedVol30, periodRealizedVol60, periodRealizedVol90 []float64
		realizedVol30, realizedVol60, realizedVol90                   float64
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
	stockDataConfig := StockDataConf{}
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

	for _, val := range os.Args {
		if val == "CAR" {
			annualizedReturnsMode = true
		}
		if val == "TAR" {
			targetAnnualizedReturnsMode = true
		}
	}

	if annualizedReturnsMode {
		fmt.Println("Current Annualized Returns Selected")
		currAnnualReturn, err := pkg.GetCurrAnnualReturn(currPrice, costBasis, startTimeMilli, short)
		if err != nil {
			log.Errorf(context.TODO(), "unable to process the current annualized return.\n")
		}
		fmt.Printf("Current Annualized return is: %f.\n", currAnnualReturn)
	} else if targetAnnualizedReturnsMode {
		targetAnnualReturn, err := pkg.GetTargetAnnualReturn(costBasis, targetAnnualizedRate, startTimeMilli, short)
		if err != nil {
			log.Errorf(context.TODO(), "unable to process the target annualized return")
		}
		fmt.Printf("Target Price is: %f.\n", targetAnnualReturn)
	} else {
		tickerData, err = pkg.GetStockPrices(strings.ToUpper(ticker), stockDataConfig.Creds, resolution, startTimeMilli, endTimeMilli)
		if err != nil {
			fmt.Errorf("unable to get stock prices")
		}
		jsonTickerData, err := json.MarshalIndent(tickerData[ticker], "", "  ")
		if err != nil {
			fmt.Errorf("error marshalling data into JSON string")
		}
		realizedVolPrices := make([]float64, len(tickerData[ticker]))
		for stock := range tickerData {
			for stockByDate := range tickerData[stock] {
				realizedVolPrices = append(realizedVolPrices, tickerData[stock][stockByDate].Close)
			}
		}
		fmt.Printf("%s\n", string(jsonTickerData))
		if len(realizedVolPrices) >= 90 {
			periodRealizedVol90 = realizedVolPrices[len(realizedVolPrices)-91 : len(realizedVolPrices)-1]
			realizedVol90 = pkg.RealizedVolatility(periodRealizedVol90)
			fmt.Printf("90 day RealizedVolatility: %f\n", realizedVol90)
		}
		if len(realizedVolPrices) >= 60 {
			periodRealizedVol60 = realizedVolPrices[len(realizedVolPrices)-61 : len(realizedVolPrices)-1]
			realizedVol60 = pkg.RealizedVolatility(periodRealizedVol60)
			fmt.Printf("60 day RealizedVolatility: %f\n", realizedVol60)
		}
		if len(realizedVolPrices) >= 30 {
			periodRealizedVol30 = realizedVolPrices[len(realizedVolPrices)-31 : len(realizedVolPrices)-1]
			realizedVol30 = pkg.RealizedVolatility(periodRealizedVol30)
			fmt.Printf("30 day RealizedVolatility: %f\n", realizedVol30)
		}
	}

	if err != nil {
		fmt.Errorf("error occurred: %w", err)
	}
	if !annualizedReturnsMode && !targetAnnualizedReturnsMode {
		jsonTickerData, err := json.MarshalIndent(tickerData[ticker], "", "  ")
		if err != nil {
			fmt.Errorf("error marshalling data into JSON string")
		}
		fmt.Printf("%v\n", string(jsonTickerData))
		fmt.Printf("I made it into the final section\n")
	}
}
