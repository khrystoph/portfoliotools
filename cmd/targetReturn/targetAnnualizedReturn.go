package main

import (
	"cmd/pkg"
	"context"
	"flag"
	"fmt"
	"google.golang.org/appengine/log"
	"time"
)

var (
	startTime, endTime              string
	costBasis, targetAnnualizedRate float64
	short                           bool
)

func init() {
	flag.StringVar(&startTime, "startTime", "30 days ago", "Enter a time to start gathering data "+
		"for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ. Time will always assume UTC.")
	flag.StringVar(&endTime, "endTime", "Today",
		"Enter a time to end gathering data for the ticker. Time must be formatted as YYYY-MM-DDTHH:MM:SSZ."+
			" Time will always assume UTC.")
	flag.Float64Var(&costBasis, "costBasis", 1, "input the cost basis in decimal form. "+
		"Example: 12.34")
	flag.Float64Var(&targetAnnualizedRate, "targetRate", .06,
		"enter either the risk-free rate or the rate you want as your target return rate. Default is: .06 (6%).")
	flag.BoolVar(&short, "short", false, "default: False. Presence of the flag means true.")
}

func main() {
	flag.Parse()

	if endTime == "Today" {
		endTime = time.Now().Format(time.RFC3339)
	}

	startTimeMilli, err := time.Parse(time.RFC3339, startTime)
	if err != nil {
		fmt.Printf("Unable to convert startTime to milliseconds.\nstartTime: %s\n", startTime)
		return
	}
	//TODO: change GetTargetAnnualReturn in TechAnalysis.go to accept optional endTimeMilli, then uncomment below
	/* endTimeMilli, err := time.Parse(time.RFC3339, endTime)
	if err != nil {
		fmt.Printf("Unable to convert endTime to milliseconds.\n")
		return
	}
	*/

	targetAnnualReturn, err := pkg.GetTargetAnnualReturn(costBasis, targetAnnualizedRate, startTimeMilli, short)
	if err != nil {
		log.Errorf(context.TODO(), "unable to process the target annualized return")
	}
	fmt.Printf("Target Price is: %f.\n", targetAnnualReturn)
}
