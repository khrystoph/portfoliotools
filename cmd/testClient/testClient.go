package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/khrystoph/portfoliotools/pkg"
	"log"
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
		log.Printf("error reading user's homedir: %v", err)
	}
	tickerConfig = strings.Replace(tickerConfig, "~", userDir, 1)
	if _, err = os.Stat(tickerConfig); errors.Is(err, os.ErrNotExist) {
		log.Printf("error: config file %s does not exist. exiting", tickerConfig)
	}
	configFile, err := os.Open(tickerConfig)
	if err != nil {
		log.Printf("error opening the config file: %v", err)
	}
	defer configFile.Close()
	configDecoder := json.NewDecoder(configFile)
	stockDataConfig := pkg.StockDataConf{}
	err = configDecoder.Decode(&stockDataConfig)
	if err != nil {
		log.Printf("error decoding the json config file: %v", err)
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
	stockData, err := pkg.GetStockPricesAlpaca(stockDataConfig, ticker, resolution, startTimeMilli, endTimeMilli, false)
	if err != nil {
		log.Printf("unable to retrieve stock data: %v", err)
	}
	stockDataJSON, jsonErr := json.MarshalIndent(stockData, "", "  ")
	if jsonErr != nil {
		log.Printf("unable to marshal JSON: %v", jsonErr)
	}
	fmt.Printf("%s\n", stockDataJSON)
}
