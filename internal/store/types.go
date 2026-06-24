package store

import "time"

// AssetClass identifies the category of a financial instrument.
type AssetClass string

const (
	AssetClassEquity    AssetClass = "equity"
	AssetClassETF       AssetClass = "etf"
	AssetClassCrypto    AssetClass = "crypto"
	AssetClassCommodity AssetClass = "commodity"
	AssetClassForex     AssetClass = "forex"
	AssetClassIndex     AssetClass = "index"
)

// DataSource identifies which data provider supplied a record.
type DataSource string

const (
	SourceAlpaca  DataSource = "alpaca"
	SourcePolygon DataSource = "polygon"
	SourceYahoo   DataSource = "yahoo"
)

// BackfillStatus is the terminal or in-progress state of a backfill job run.
type BackfillStatus string

const (
	BackfillStatusRunning   BackfillStatus = "running"
	BackfillStatusCompleted BackfillStatus = "completed"
	BackfillStatusFailed    BackfillStatus = "failed"
)

// BackfillTickerStatus is the result for a single ticker within a backfill run.
type BackfillTickerStatus string

const (
	BackfillTickerSuccess BackfillTickerStatus = "success"
	BackfillTickerFailed  BackfillTickerStatus = "failed"
	BackfillTickerSkipped BackfillTickerStatus = "skipped"
)

// Ticker is a financial instrument tracked by the system.
type Ticker struct {
	ID            int64
	Symbol        string
	Name          string
	ExchangeID    *int32     // nil for assets with no single exchange (crypto, forex, indices)
	AssetClassID  int16
	AssetClass    AssetClass
	PrimarySource DataSource
	Currency      string
	Active        bool
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// OHLCVDaily is one day's raw price candle for a ticker.
// Financial values are stored as float64 for calculation convenience;
// the underlying DB column is NUMERIC(18,6) for precision.
type OHLCVDaily struct {
	ID             int64
	TickerID       int64
	TradeDate      time.Time
	Open           float64
	High           float64
	Low            float64
	Close          float64
	Volume         float64
	WeightedVolume *float64
	Transactions   *int64
	AdjClose       *float64   // nil when not provided by the source
	Source         DataSource
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

// BackfillRun is an audit record for one execution of the daily backfill job.
type BackfillRun struct {
	ID               int64
	StartedAt        time.Time
	CompletedAt      *time.Time
	Status           BackfillStatus
	TickersProcessed int
	TickersFailed    int
	ErrorMsg         *string
	CreatedAt        time.Time
}

// BackfillTickerLog records the outcome for a single ticker within a BackfillRun.
type BackfillTickerLog struct {
	ID            int64
	RunID         int64
	TickerID      int64
	Status        BackfillTickerStatus
	CandlesStored int
	ErrorMsg      *string
	DurationMS    *int
	CreatedAt     time.Time
}
