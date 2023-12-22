package main

import (
	"cmd/pkg"
	"flag"
	"fmt"
	"time"
)

func init() {
	flag.StringVar(&ticker, "ticker", "AAPL", "Enter stock ticker to look up price info for")
	flag.StringVar(&startTime, "startTime", "30 days ago", "Enter a time to start gathering data for the ticker."+
		"Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&endTime, "endTime", "Today", "Enter a time to end gathering data for the ticker."+
		"Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&resolution, "resolution", "day", "Input the resolution to pull data. "+
		"Supported values: second, minute, hour, day, week, month, quarter, year. The numeric time values represent minutes. "+
		"Default resolution: day.")
}

var (
	ticker, startTime, endTime, resolution string
)

type FinnhubConf struct {
	Creds string
}

func main() {
	flag.Parse()
	var apiTokenString = ""

	startTimeMilli, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		fmt.Printf("Unable to convert startTime to milliseconds.")
		return
	}
	endTimeMilli, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		fmt.Printf("Unable to convert startTime to milliseconds.")
		return
	}

	tickerData, err := pkg.GetStockPrices(ticker, apiTokenString, resolution, startTimeMilli, endTimeMilli)

	if err != nil {
		return
	}
	fmt.Printf("%+v\n", tickerData)
}
