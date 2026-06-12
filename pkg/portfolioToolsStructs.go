package pkg

import (
	"time"
)

// Vapr
type Vapr struct {
	Ticker     string  `json:"ticker"`
	LatestHigh float64 `json:"latest_high"`
	LatestLow  float64 `json:"latest_low"`
}

type StockDataConf struct {
	PolygonAPIToken string   `json:"polygon-api-key"`
	AlpacaAPIKey    string   `json:"alpaca-api-key"`
	AlpacaSecretKey string   `json:"alpaca-secret-key"`
	RangeAdjustment float64  `json:"probable-range-adj"`
	EmailAddress    string   `json:"email-address"`
	EmailPassword   string   `json:"email-password"`
	Hostname        string   `json:"hostname"`
	Port            int      `json:"port"`
	MailTo          []string `json:"mail-to"`
}

// OHLC is a struct that contains the Open, High, Low, and Close values from a range of times for a specific ticker
type OHLC struct {
	Close            []float64 `json:"c"`
	High             []float64 `json:"h"`
	Low              []float64 `json:"l"`
	Status           string    `json:"s"`
	Timestamp        []int64   `json:"t"`
	TransactionCount []int64   `json:"n"`
	Volume           []int64   `json:"v"`
}

type HistoricalBar struct {
	Close            float64 `json:"c"`
	High             float64 `json:"h"`
	Low              float64 `json:"l"`
	TransactionCount int64   `json:"n"`
	Open             float64 `json:"o"`
	Timestamp        string  `json:"t"`
	Volume           int64   `json:"v"`
	WeightedVolume   float64 `json:"vw"`
}

type SingleStockCandle struct {
	Ticker              string             `json:"ticker"`
	Close               float64            `json:"close"`
	High                float64            `json:"high"`
	Low                 float64            `json:"low"`
	Open                float64            `json:"open"`
	Transactions        int64              `json:"transactions"`
	Timestamp           time.Time          `json:"timestamp"`
	Volume              float64            `json:"volume"`
	WeightedVolume      float64            `json:"weighted-volume"`
	PriceVelocity       float64            `json:"price-velocity"`
	PriceAccel          float64            `json:"price-acceleration"`
	AvgVolumeShort      float64            `json:"short-avg-volume"`
	AvgVolumeMed        float64            `json:"med-avg-volume"`
	AvgVolumeLong       float64            `json:"long-avg-volume"`
	AvgVolumeRatioShort float64            `json:"short-avg-volume-ratio"`
	AvgVolumeRatioMed   float64            `json:"med-avg-volume-ratio"`
	AvgVolumeRatioLong  float64            `json:"long-avg-volume-ratio"`
	ShortPrices         map[string]float64 `json:"short-prices"`
	MedPrices           map[string]float64 `json:"med-prices"`
	LongPrices          map[string]float64 `json:"long-prices"`
	SlopeShortDuration  float64            `json:"trade-slope"`
	SlopeMedDuration    float64            `json:"trend-slope"`
	SlopeLongDuration   float64            `json:"tail-slope"`
	SlopeShortValid     bool               `json:"-"`
	SlopeMedValid       bool               `json:"-"`
	SlopeLongValid      bool               `json:"-"`
	TradeDirection      string             `json:"trade-direction"`
	TrendDirection      string             `json:"trend-direction"`
	TailDirection       string             `json:"tail-direction"`
	RVolHighShort       float64            `json:"short-rvol-high"`
	RVolLowShort        float64            `json:"short-rvol-low"`
	RVolHighMed         float64            `json:"med-rvol-high"`
	RVolLowMed          float64            `json:"med-rvol-low"`
	RVolHighLong        float64            `json:"long-rvol-high"`
	RVolLowLong         float64            `json:"long-rvol-low"`
	RVolPercentShort    float64            `json:"short-rvol-percent"`
	RVolPercentMed      float64            `json:"med-rvol-percent"`
	RVolPercentLong     float64            `json:"long-rvol-percent"`
	RealizedVolatilityShort  float64       `json:"short-realized-volatility"`
	RealizedVolatilityMed    float64       `json:"med-realized-volatility"`
	RealizedVolatilityLong   float64       `json:"long-realized-volatility"`
	VelocityRealizedVolShort float64       `json:"short-rvol-velocity"`
	VelocityRealizedVolMed   float64       `json:"med-rvol-velocity"`
	VelocityRealizedVolLong  float64       `json:"long-rvol-velocity"`
	RealizedVolAccelShort    float64       `json:"short-rvol-accel"`
	RealizedVolAccelMed      float64       `json:"med-rvol-accel"`
	RealizedVolAccelLong     float64       `json:"long-rvol-accel"`
	TradeRange          map[string]float64 `json:"trade-range"`
	TrendRange          map[string]float64 `json:"trend-range"`
	TailRange           map[string]float64 `json:"tail-range"`
	TradeRangeAdj       map[string]float64 `json:"trade-range-vadj"`
	TrendRangeAdj       map[string]float64 `json:"trend-range-vadj"`
	TailRangeAdj        map[string]float64 `json:"tail-range-vadj"`
	PTradeRange         map[string]float64 `json:"prob-adj-trade-range"`
	PTrendRange         map[string]float64 `json:"prob-adj-trend-range"`
	PTailRange          map[string]float64 `json:"prob-adj-tail-range"`
	PTradeRangeAdj      map[string]float64 `json:"prob-trade-range-vadj"`
	PTrendRangeAdj      map[string]float64 `json:"prob-trend-range-vadj"`
	PTailRangeAdj       map[string]float64 `json:"prob-tail-range-vadj"`
}

type condensedStockCandle struct {
	Ticker              string             `json:"ticker"`
	Close               float64            `json:"close"`
	Volume              float64            `json:"volume"`
	PriceVelocity       float64            `json:"price-velocity"`
	PriceAcceleration   float64            `json:"price-acceleration"`
	Timestamp           time.Time          `json:"timestamp"`
	AvgVolumeShort      float64            `json:"short-avg-volume"`
	AvgVolumeRatioShort float64            `json:"short-avg-volume-ratio"`
	TradeSlope          float64            `json:"trade-slope"`
	RVolShort           float64            `json:"rvol-short"`
	RVolShortVel        float64            `json:"rvol-short-vel"`
	RVolShortAccel      float64            `json:"rvol-short-accel"`
	RVolPercentShort    float64            `json:"short-day-rvol-range-percent"`
	RVolHighShort       float64            `json:"short-day-rvol-high"`
	RVolLowShort        float64            `json:"short-day-rvol-low"`
	TradeRangeAdj       map[string]float64 `json:"trade-range-vadj"`
	PtradeRangeAdj      map[string]float64 `json:"prob-trade-range-vadj"`
	AvgVolumeMed        float64            `json:"med-avg-volume"`
	AvgVolumeRatioMed   float64            `json:"med-avg-volume-ratio"`
	TrendSlope          float64            `json:"trend-slope"`
	RVolMed             float64            `json:"rvol-med"`
	RVolMedVel          float64            `json:"rvol-med-vel"`
	RVolMedAccel        float64            `json:"rvol-med-accel"`
	RVolPercentMed      float64            `json:"med-day-rvol-range-percent"`
	RVolHighMed         float64            `json:"med-day-rvol-high"`
	RVolLowMed          float64            `json:"med-day-rvol-low"`
	TrendRangeAdj       map[string]float64 `json:"trend-range-vadj"`
	PTrendRangeAdj      map[string]float64 `json:"prob-trend-range-vadj"`
	AvgVolumeLong       float64            `json:"long-avg-volume"`
	AvgVolumeRatioLong  float64            `json:"long-avg-volume-ratio"`
	TailSlope           float64            `json:"tail-slope"`
	TradeDirection      string             `json:"trade-direction"`
	TrendDirection      string             `json:"trend-direction"`
	TailDirection       string             `json:"tail-direction"`
	RVolLong            float64            `json:"rvol-long"`
	RVolLongVel         float64            `json:"rvol-long-vel"`
	RVolLongAccel       float64            `json:"rvol-long-accel"`
	RVolPercentLong     float64            `json:"long-day-rvol-range-percent"`
	RVolHighLong        float64            `json:"long-day-rvol-high"`
	RVolLowLong         float64            `json:"long-day-rvol-low"`
	TailRangeAdj        map[string]float64 `json:"tail-range-vadj"`
	PTailRangeAdj       map[string]float64 `json:"prob-tail-range-vadj"`
}

type CondensedRangesJSON struct {
	Ticker         string    `json:"ticker"`
	Close          float64   `json:"close"`
	AvgVolRatio    float64   `json:"avg_vol_ratio"`
	RVolPercent    float64   `json:"rvol_percent"`
	RiskRangeHigh  float64   `json:"rr_high"`
	RiskRangeLow   float64   `json:"rr_low"`
	TradeSlope     float64   `json:"trade-slope"`
	TrendSlope     float64   `json:"trend-slope"`
	TailSlope      float64   `json:"tail-slope"`
	TradeDirection string    `json:"trade-direction"`
	TrendDirection string    `json:"trend-direction"`
	TailDirection  string    `json:"tail-direction"`
	Timestamp      time.Time `json:"timestamp"`
}
