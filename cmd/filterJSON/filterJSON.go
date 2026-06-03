package main

import (
	"cmd/pkg"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	inFile, ticker string
)

func init() {
	flag.StringVar(&inFile, "infile", "stockRanges.json", "name of the json file to parse")
	flag.StringVar(&inFile, "f", "stockRanges.json", "name of the json file to parse")
	flag.StringVar(&ticker, "ticker", "ZZZ", "the ticker to filter for out of the data")
	flag.StringVar(&ticker, "t", "ZZZ", "the ticker to filter for out of the data")
}

func main() {
	var (
		data = map[string]pkg.CondensedRangesJSON{}
	)
	flag.Parse()
	userDir, err := os.UserHomeDir()
	if err != nil {
		_ = fmt.Errorf("error reading user's homedir: %w", err)
	}
	userDir = strings.Replace(userDir, "~", userDir, 1)
	readFile, err := os.ReadFile(inFile)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(readFile, &data)
	if err != nil {
		log.Fatal(err)
	}

	ticker = strings.ToUpper(ticker)
	if _, ok := data[ticker]; ok {
		jsonData, err := json.MarshalIndent(data[ticker], "", "  ")
		if err != nil {
			log.Fatal(err)
		}
		log.Printf("%s:\n%s\n", ticker, jsonData)
	}
	return
}
