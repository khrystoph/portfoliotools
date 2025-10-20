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
	flag.StringVar(&resolution, "resolution", "1D", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year."+
		" The numeric time values represent minutes. Default resolution: day.")
}
func main() {
	flag.Parse()
	var (
		err     error
		userDir string
	)

	if endTime == "Today" {
		endTime = time.Now().Format(time.RFC3339)
	}

	// Section parses the config file location, opens it, decodes the JSON and loads the API creds
	userDir, err = os.UserHomeDir()
	if err != nil {
		fmt.Errorf("error reading user's homedir: %w", err) //nolint:govet
	}
	tickerConfig = strings.Replace(tickerConfig, "~", userDir, 1)
	if _, err = os.Stat(tickerConfig); errors.Is(err, os.ErrNotExist) {
		fmt.Errorf("error: config file %s does not exist. exiting", tickerConfig) //nolint:govet
	}
	configFile, err := os.Open(tickerConfig)
	if err != nil {
		fmt.Errorf("error opening the config file. %w", err) //nolint:govet
	}
	defer configFile.Close()
	configDecoder := json.NewDecoder(configFile)
	stockDataConfig := pkg.StockDataConf{}
	err = configDecoder.Decode(&stockDataConfig)
	if err != nil {
		fmt.Errorf("error decoding the json config file. exiting. error msg: %w", err) //nolint:govet
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
	stockData, err := pkg.GetStockPricesAlpaca(stockDataConfig, ticker, resolution, startTimeMilli, endTimeMilli)
	if err != nil {
		fmt.Errorf("unable to retrieve stock data: %e", err)
	}
	stockDataJSON, jsonErr := json.MarshalIndent(stockData, "", "  ")
	if jsonErr != nil {
		fmt.Errorf("unable to marshal JSON: %e", jsonErr)
	}
	fmt.Printf("%s\n", stockDataJSON)
}
