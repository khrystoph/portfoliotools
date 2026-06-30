package universe

import "github.com/khrystoph/portfoliotools/internal/store"

// StaticAsset is a manually curated instrument fetched from Yahoo Finance.
// Static assets are never swept or deactivated by the sync job.
type StaticAsset struct {
	Symbol string
	Name   string
	Class  store.AssetClass
}

var Commodities = []StaticAsset{
	{"GC=F", "Gold", store.AssetClassCommodity},
	{"SI=F", "Silver", store.AssetClassCommodity},
	{"CL=F", "Crude Oil (WTI)", store.AssetClassCommodity},
	{"NG=F", "Natural Gas", store.AssetClassCommodity},
	{"ZW=F", "Wheat", store.AssetClassCommodity},
	{"ZC=F", "Corn", store.AssetClassCommodity},
	{"ZS=F", "Soybeans", store.AssetClassCommodity},
	{"HG=F", "Copper", store.AssetClassCommodity},
	{"ALI=F", "Aluminum", store.AssetClassCommodity},
}

var Indices = []StaticAsset{
	{"^GSPC", "S&P 500", store.AssetClassIndex},
	{"^IXIC", "NASDAQ Composite", store.AssetClassIndex},
	{"^NDX", "NASDAQ 100", store.AssetClassIndex},
	{"^DJI", "Dow Jones Industrial Avg", store.AssetClassIndex},
	{"^RUT", "Russell 2000", store.AssetClassIndex},
	{"^VIX", "CBOE Volatility Index", store.AssetClassIndex},
	{"^KS11", "KOSPI", store.AssetClassIndex},
	{"^FTSE", "FTSE 100", store.AssetClassIndex},
	{"^GDAXI", "DAX", store.AssetClassIndex},
	{"^N225", "Nikkei 225", store.AssetClassIndex},
	{"^HSI", "Hang Seng", store.AssetClassIndex},
}

var BondYields = []StaticAsset{
	{"^IRX", "13-Week T-Bill Yield", store.AssetClassBond},
	{"^FVX", "5-Year Treasury Yield", store.AssetClassBond},
	{"^TNX", "10-Year Treasury Yield", store.AssetClassBond},
	{"^TYX", "30-Year Treasury Yield", store.AssetClassBond},
}

var ForexPairs = []StaticAsset{
	{"EURUSD=X", "EUR/USD", store.AssetClassForex},
	{"JPY=X", "USD/JPY", store.AssetClassForex},
	{"GBPUSD=X", "GBP/USD", store.AssetClassForex},
	{"CHF=X", "USD/CHF", store.AssetClassForex},
	{"AUDUSD=X", "AUD/USD", store.AssetClassForex},
	{"CAD=X", "USD/CAD", store.AssetClassForex},
	{"NZDUSD=X", "NZD/USD", store.AssetClassForex},
	{"EURGBP=X", "EUR/GBP", store.AssetClassForex},
	{"EURJPY=X", "EUR/JPY", store.AssetClassForex},
	{"EURCHF=X", "EUR/CHF", store.AssetClassForex},
	{"HKD=X", "USD/HKD", store.AssetClassForex},
	{"SGD=X", "USD/SGD", store.AssetClassForex},
	{"SEK=X", "USD/SEK", store.AssetClassForex},
	{"NOK=X", "USD/NOK", store.AssetClassForex},
	{"MXN=X", "USD/MXN", store.AssetClassForex},
	{"CNY=X", "USD/CNY", store.AssetClassForex},
	{"INR=X", "USD/INR", store.AssetClassForex},
	{"ZAR=X", "USD/ZAR", store.AssetClassForex},
	{"BRL=X", "USD/BRL", store.AssetClassForex},
	{"TRY=X", "USD/TRY", store.AssetClassForex},
}

// AllStaticAssets returns every static asset across all categories.
func AllStaticAssets() []StaticAsset {
	out := make([]StaticAsset, 0, len(Commodities)+len(Indices)+len(BondYields)+len(ForexPairs))
	out = append(out, Commodities...)
	out = append(out, Indices...)
	out = append(out, BondYields...)
	out = append(out, ForexPairs...)
	return out
}
