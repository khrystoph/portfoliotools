package main

import (
	"cmd/pkg"
	"context"
	"flag"
	"fmt"
	"google.golang.org/appengine/log"
	"os"
	"time"
)

var (
	ticker, startTime, endTime, resolution, tickerConfig string
	costBasis, currPrice                                 float64
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
	flag.BoolVar(&short, "short", false, "default: False. Presence of the flag means true.")
}

type StockDataConf struct {
	Creds string
}

func main() {
	flag.Parse()
	var (
		apiTokenString        = ""
		tickerData            map[string]map[int64]pkg.SingleStockCandle
		annualizedReturnsMode = false
	)

	if endTime == "Today" {
		endTime = time.Now().Format(time.RFC3339)
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
	}

	if annualizedReturnsMode {
		fmt.Println("Current Annualized Returns Selected")
		currAnnualReturn, err := pkg.GetCurrAnnualReturn(currPrice, costBasis, startTimeMilli, short)
		if err != nil {
			log.Errorf(context.TODO(), "unable to process the current anualized return.\n")
		}
		fmt.Printf("Current Annualized return is: %f.\n", currAnnualReturn)
	} else {
		tickerData, err = pkg.GetStockPrices(ticker, apiTokenString, resolution, startTimeMilli, endTimeMilli)
	}

	if err != nil {
		return
	}
	if !annualizedReturnsMode {
		fmt.Printf("%+v\n", tickerData)
	}
}
