package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/khrystoph/portfoliotools/internal/db"
	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/universe"
)

func main() {
	mode := flag.String("mode", "", "sync mode: daily or weekly")
	flag.Parse()

	if *mode != "daily" && *mode != "weekly" {
		log.Fatalf("--mode must be 'daily' or 'weekly', got %q", *mode)
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set")
	}

	polygonKey := os.Getenv("POLYGON_API_KEY")
	alpacaKey := os.Getenv("ALPACA_API_KEY")
	alpacaSecret := os.Getenv("ALPACA_API_SECRET")
	openFIGIKey := os.Getenv("OPENFIGI_API_KEY")

	ctx := context.Background()
	pool, err := db.Connect(ctx, connStr)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	defer pool.Close()

	ts := store.NewTickerStore(pool)
	ms := store.NewTickerMigrationStore(pool)
	sc := store.NewSystemConfigStore(pool)

	cfg, err := universe.LoadSyncConfig(ctx, sc)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	polygonClient := universe.NewPolygonClient("https://api.polygon.io", polygonKey)
	figiClient := universe.NewOpenFIGIClient("https://api.openfigi.com", openFIGIKey)
	alpacaClient := universe.NewAlpacaClient("https://paper-api.alpaca.markets", alpacaKey, alpacaSecret)

	syncer := universe.NewSyncer(ts, ms, sc, polygonClient, figiClient, alpacaClient, cfg)

	switch *mode {
	case "daily":
		if err := syncer.Daily(ctx); err != nil {
			log.Fatalf("daily sync: %v", err)
		}
	case "weekly":
		if err := syncer.Weekly(ctx); err != nil {
			log.Fatalf("weekly sync: %v", err)
		}
	}
	log.Printf("%s sync complete", *mode)
}
