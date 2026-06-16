package pkg

import "testing"

func TestGetSetRVol(t *testing.T) {
	c := SingleStockCandle{
		RealizedVolatilityShort: 0.15,
		RealizedVolatilityMed:   0.20,
		RealizedVolatilityLong:  0.25,
	}
	if got := getRVol(c, SHORTDURATION); got != 0.15 {
		t.Errorf("SHORTDURATION: got %v want 0.15", got)
	}
	if got := getRVol(c, MEDIUMDURATION); got != 0.20 {
		t.Errorf("MEDIUMDURATION: got %v want 0.20", got)
	}
	if got := getRVol(c, LONGDURATION); got != 0.25 {
		t.Errorf("LONGDURATION: got %v want 0.25", got)
	}
	if got := getRVol(c, 999); got != 0 {
		t.Errorf("unknown duration: got %v want 0", got)
	}
	var w SingleStockCandle
	setRVol(&w, SHORTDURATION, 0.15)
	setRVol(&w, MEDIUMDURATION, 0.20)
	setRVol(&w, LONGDURATION, 0.25)
	if w.RealizedVolatilityShort != 0.15 {
		t.Errorf("Short field: got %v want 0.15", w.RealizedVolatilityShort)
	}
	if w.RealizedVolatilityMed != 0.20 {
		t.Errorf("Med field: got %v want 0.20", w.RealizedVolatilityMed)
	}
	if w.RealizedVolatilityLong != 0.25 {
		t.Errorf("Long field: got %v want 0.25", w.RealizedVolatilityLong)
	}
}

func TestGetSetAvgVol(t *testing.T) {
	c := SingleStockCandle{AvgVolumeShort: 1.0, AvgVolumeMed: 2.0, AvgVolumeLong: 3.0}
	if got := getAvgVol(c, SHORTDURATION); got != 1.0 {
		t.Errorf("SHORTDURATION: got %v want 1.0", got)
	}
	if got := getAvgVol(c, MEDIUMDURATION); got != 2.0 {
		t.Errorf("MEDIUMDURATION: got %v want 2.0", got)
	}
	if got := getAvgVol(c, LONGDURATION); got != 3.0 {
		t.Errorf("LONGDURATION: got %v want 3.0", got)
	}
	var w SingleStockCandle
	setAvgVol(&w, SHORTDURATION, 1.0)
	setAvgVol(&w, MEDIUMDURATION, 2.0)
	setAvgVol(&w, LONGDURATION, 3.0)
	if w.AvgVolumeShort != 1.0 || w.AvgVolumeMed != 2.0 || w.AvgVolumeLong != 3.0 {
		t.Errorf("setAvgVol: Short=%v Med=%v Long=%v", w.AvgVolumeShort, w.AvgVolumeMed, w.AvgVolumeLong)
	}
}

func TestGetSetAvgVolRatio(t *testing.T) {
	c := SingleStockCandle{AvgVolumeRatioShort: 1.1, AvgVolumeRatioMed: 1.2, AvgVolumeRatioLong: 1.3}
	if got := getAvgVolRatio(c, SHORTDURATION); got != 1.1 {
		t.Errorf("Short: got %v want 1.1", got)
	}
	if got := getAvgVolRatio(c, MEDIUMDURATION); got != 1.2 {
		t.Errorf("Med: got %v want 1.2", got)
	}
	if got := getAvgVolRatio(c, LONGDURATION); got != 1.3 {
		t.Errorf("Long: got %v want 1.3", got)
	}
	var w SingleStockCandle
	setAvgVolRatio(&w, SHORTDURATION, 1.1)
	setAvgVolRatio(&w, MEDIUMDURATION, 1.2)
	setAvgVolRatio(&w, LONGDURATION, 1.3)
	if w.AvgVolumeRatioShort != 1.1 || w.AvgVolumeRatioMed != 1.2 || w.AvgVolumeRatioLong != 1.3 {
		t.Errorf("setAvgVolRatio: Short=%v Med=%v Long=%v", w.AvgVolumeRatioShort, w.AvgVolumeRatioMed, w.AvgVolumeRatioLong)
	}
}

func TestGetSetRVolVel(t *testing.T) {
	c := SingleStockCandle{VelocityRealizedVolShort: 0.01, VelocityRealizedVolMed: 0.02, VelocityRealizedVolLong: 0.03}
	if got := getRVolVel(c, SHORTDURATION); got != 0.01 {
		t.Errorf("Short: got %v want 0.01", got)
	}
	if got := getRVolVel(c, MEDIUMDURATION); got != 0.02 {
		t.Errorf("Med: got %v want 0.02", got)
	}
	if got := getRVolVel(c, LONGDURATION); got != 0.03 {
		t.Errorf("Long: got %v want 0.03", got)
	}
	var w SingleStockCandle
	setRVolVel(&w, SHORTDURATION, 0.01)
	setRVolVel(&w, MEDIUMDURATION, 0.02)
	setRVolVel(&w, LONGDURATION, 0.03)
	if w.VelocityRealizedVolShort != 0.01 || w.VelocityRealizedVolMed != 0.02 || w.VelocityRealizedVolLong != 0.03 {
		t.Errorf("setRVolVel: Short=%v Med=%v Long=%v", w.VelocityRealizedVolShort, w.VelocityRealizedVolMed, w.VelocityRealizedVolLong)
	}
}

func TestGetSetRVolAccel(t *testing.T) {
	c := SingleStockCandle{RealizedVolAccelShort: 0.001, RealizedVolAccelMed: 0.002, RealizedVolAccelLong: 0.003}
	if got := getRVolAccel(c, SHORTDURATION); got != 0.001 {
		t.Errorf("Short: got %v want 0.001", got)
	}
	if got := getRVolAccel(c, MEDIUMDURATION); got != 0.002 {
		t.Errorf("Med: got %v want 0.002", got)
	}
	if got := getRVolAccel(c, LONGDURATION); got != 0.003 {
		t.Errorf("Long: got %v want 0.003", got)
	}
	var w SingleStockCandle
	setRVolAccel(&w, SHORTDURATION, 0.001)
	setRVolAccel(&w, MEDIUMDURATION, 0.002)
	setRVolAccel(&w, LONGDURATION, 0.003)
	if w.RealizedVolAccelShort != 0.001 || w.RealizedVolAccelMed != 0.002 || w.RealizedVolAccelLong != 0.003 {
		t.Errorf("setRVolAccel: Short=%v Med=%v Long=%v", w.RealizedVolAccelShort, w.RealizedVolAccelMed, w.RealizedVolAccelLong)
	}
}

func TestGetSetRVolHigh(t *testing.T) {
	c := SingleStockCandle{RVolHighShort: 0.3, RVolHighMed: 0.4, RVolHighLong: 0.5}
	if got := getRVolHigh(c, SHORTDURATION); got != 0.3 {
		t.Errorf("Short: got %v want 0.3", got)
	}
	if got := getRVolHigh(c, MEDIUMDURATION); got != 0.4 {
		t.Errorf("Med: got %v want 0.4", got)
	}
	if got := getRVolHigh(c, LONGDURATION); got != 0.5 {
		t.Errorf("Long: got %v want 0.5", got)
	}
	var w SingleStockCandle
	setRVolHigh(&w, SHORTDURATION, 0.3)
	setRVolHigh(&w, MEDIUMDURATION, 0.4)
	setRVolHigh(&w, LONGDURATION, 0.5)
	if w.RVolHighShort != 0.3 || w.RVolHighMed != 0.4 || w.RVolHighLong != 0.5 {
		t.Errorf("setRVolHigh: Short=%v Med=%v Long=%v", w.RVolHighShort, w.RVolHighMed, w.RVolHighLong)
	}
}

func TestGetSetRVolLow(t *testing.T) {
	c := SingleStockCandle{RVolLowShort: 0.1, RVolLowMed: 0.15, RVolLowLong: 0.2}
	if got := getRVolLow(c, SHORTDURATION); got != 0.1 {
		t.Errorf("Short: got %v want 0.1", got)
	}
	if got := getRVolLow(c, MEDIUMDURATION); got != 0.15 {
		t.Errorf("Med: got %v want 0.15", got)
	}
	if got := getRVolLow(c, LONGDURATION); got != 0.2 {
		t.Errorf("Long: got %v want 0.2", got)
	}
	var w SingleStockCandle
	setRVolLow(&w, SHORTDURATION, 0.1)
	setRVolLow(&w, MEDIUMDURATION, 0.15)
	setRVolLow(&w, LONGDURATION, 0.2)
	if w.RVolLowShort != 0.1 || w.RVolLowMed != 0.15 || w.RVolLowLong != 0.2 {
		t.Errorf("setRVolLow: Short=%v Med=%v Long=%v", w.RVolLowShort, w.RVolLowMed, w.RVolLowLong)
	}
}

func TestGetSetRVolPercent(t *testing.T) {
	c := SingleStockCandle{RVolPercentShort: 0.5, RVolPercentMed: 0.6, RVolPercentLong: 0.7}
	if got := getRVolPercent(c, SHORTDURATION); got != 0.5 {
		t.Errorf("Short: got %v want 0.5", got)
	}
	if got := getRVolPercent(c, MEDIUMDURATION); got != 0.6 {
		t.Errorf("Med: got %v want 0.6", got)
	}
	if got := getRVolPercent(c, LONGDURATION); got != 0.7 {
		t.Errorf("Long: got %v want 0.7", got)
	}
	var w SingleStockCandle
	setRVolPercent(&w, SHORTDURATION, 0.5)
	setRVolPercent(&w, MEDIUMDURATION, 0.6)
	setRVolPercent(&w, LONGDURATION, 0.7)
	if w.RVolPercentShort != 0.5 || w.RVolPercentMed != 0.6 || w.RVolPercentLong != 0.7 {
		t.Errorf("setRVolPercent: Short=%v Med=%v Long=%v", w.RVolPercentShort, w.RVolPercentMed, w.RVolPercentLong)
	}
}

func TestGetSetPrices(t *testing.T) {
	short := map[string]float64{"2024-01-01": 100.0}
	med := map[string]float64{"2024-01-01": 101.0}
	long := map[string]float64{"2024-01-01": 102.0}
	c := SingleStockCandle{ShortPrices: short, MedPrices: med, LongPrices: long}
	if got := getPrices(c, SHORTDURATION); got["2024-01-01"] != 100.0 {
		t.Errorf("Short: got %v want 100.0", got)
	}
	if got := getPrices(c, MEDIUMDURATION); got["2024-01-01"] != 101.0 {
		t.Errorf("Med: got %v want 101.0", got)
	}
	if got := getPrices(c, LONGDURATION); got["2024-01-01"] != 102.0 {
		t.Errorf("Long: got %v want 102.0", got)
	}
	var w SingleStockCandle
	setPrices(&w, SHORTDURATION, short)
	setPrices(&w, MEDIUMDURATION, med)
	setPrices(&w, LONGDURATION, long)
	if w.ShortPrices["2024-01-01"] != 100.0 || w.MedPrices["2024-01-01"] != 101.0 || w.LongPrices["2024-01-01"] != 102.0 {
		t.Errorf("setPrices mismatch")
	}
}

func TestGetSetRiskRange(t *testing.T) {
	trade := map[string]float64{"high": 110.0, "low": 90.0}
	trend := map[string]float64{"high": 115.0, "low": 85.0}
	tail := map[string]float64{"high": 120.0, "low": 80.0}
	c := SingleStockCandle{TradeRange: trade, TrendRange: trend, TailRange: tail}
	if got := getRiskRange(c, SHORTDURATION); got["high"] != 110.0 {
		t.Errorf("Trade range high: got %v want 110.0", got["high"])
	}
	if got := getRiskRange(c, MEDIUMDURATION); got["high"] != 115.0 {
		t.Errorf("Trend range high: got %v want 115.0", got["high"])
	}
	if got := getRiskRange(c, LONGDURATION); got["high"] != 120.0 {
		t.Errorf("Tail range high: got %v want 120.0", got["high"])
	}
	var w SingleStockCandle
	setRiskRange(&w, SHORTDURATION, trade)
	setRiskRange(&w, MEDIUMDURATION, trend)
	setRiskRange(&w, LONGDURATION, tail)
	if w.TradeRange["high"] != 110.0 || w.TrendRange["high"] != 115.0 || w.TailRange["high"] != 120.0 {
		t.Errorf("setRiskRange mismatch")
	}
}

func TestGetSetAdjRiskRange(t *testing.T) {
	trade := map[string]float64{"high": 109.0, "low": 91.0}
	trend := map[string]float64{"high": 114.0, "low": 86.0}
	tail := map[string]float64{"high": 119.0, "low": 81.0}
	c := SingleStockCandle{TradeRangeAdj: trade, TrendRangeAdj: trend, TailRangeAdj: tail}
	if got := getAdjRiskRange(c, SHORTDURATION); got["high"] != 109.0 {
		t.Errorf("TradeAdj high: got %v want 109.0", got["high"])
	}
	if got := getAdjRiskRange(c, MEDIUMDURATION); got["high"] != 114.0 {
		t.Errorf("TrendAdj high: got %v want 114.0", got["high"])
	}
	if got := getAdjRiskRange(c, LONGDURATION); got["high"] != 119.0 {
		t.Errorf("TailAdj high: got %v want 119.0", got["high"])
	}
	var w SingleStockCandle
	setAdjRiskRange(&w, SHORTDURATION, trade)
	setAdjRiskRange(&w, MEDIUMDURATION, trend)
	setAdjRiskRange(&w, LONGDURATION, tail)
	if w.TradeRangeAdj["high"] != 109.0 || w.TrendRangeAdj["high"] != 114.0 || w.TailRangeAdj["high"] != 119.0 {
		t.Errorf("setAdjRiskRange mismatch")
	}
}

func TestGetSetProbRiskRange(t *testing.T) {
	trade := map[string]float64{"high": 108.0, "low": 92.0}
	trend := map[string]float64{"high": 113.0, "low": 87.0}
	tail := map[string]float64{"high": 118.0, "low": 82.0}
	c := SingleStockCandle{PTradeRange: trade, PTrendRange: trend, PTailRange: tail}
	if got := getProbRiskRange(c, SHORTDURATION); got["high"] != 108.0 {
		t.Errorf("PTradeRange high: got %v want 108.0", got["high"])
	}
	if got := getProbRiskRange(c, MEDIUMDURATION); got["high"] != 113.0 {
		t.Errorf("PTrendRange high: got %v want 113.0", got["high"])
	}
	if got := getProbRiskRange(c, LONGDURATION); got["high"] != 118.0 {
		t.Errorf("PTailRange high: got %v want 118.0", got["high"])
	}
	var w SingleStockCandle
	setProbRiskRange(&w, SHORTDURATION, trade)
	setProbRiskRange(&w, MEDIUMDURATION, trend)
	setProbRiskRange(&w, LONGDURATION, tail)
	if w.PTradeRange["high"] != 108.0 || w.PTrendRange["high"] != 113.0 || w.PTailRange["high"] != 118.0 {
		t.Errorf("setProbRiskRange mismatch")
	}
}

func TestGetSetProbAdjRiskRange(t *testing.T) {
	trade := map[string]float64{"high": 107.0, "low": 93.0}
	trend := map[string]float64{"high": 112.0, "low": 88.0}
	tail := map[string]float64{"high": 117.0, "low": 83.0}
	c := SingleStockCandle{PTradeRangeAdj: trade, PTrendRangeAdj: trend, PTailRangeAdj: tail}
	if got := getProbAdjRiskRange(c, SHORTDURATION); got["high"] != 107.0 {
		t.Errorf("PTradeRangeAdj high: got %v want 107.0", got["high"])
	}
	if got := getProbAdjRiskRange(c, MEDIUMDURATION); got["high"] != 112.0 {
		t.Errorf("PTrendRangeAdj high: got %v want 112.0", got["high"])
	}
	if got := getProbAdjRiskRange(c, LONGDURATION); got["high"] != 117.0 {
		t.Errorf("PTailRangeAdj high: got %v want 117.0", got["high"])
	}
	var w SingleStockCandle
	setProbAdjRiskRange(&w, SHORTDURATION, trade)
	setProbAdjRiskRange(&w, MEDIUMDURATION, trend)
	setProbAdjRiskRange(&w, LONGDURATION, tail)
	if w.PTradeRangeAdj["high"] != 107.0 || w.PTrendRangeAdj["high"] != 112.0 || w.PTailRangeAdj["high"] != 117.0 {
		t.Errorf("setProbAdjRiskRange mismatch")
	}
}

func TestGetSetSlope(t *testing.T) {
	c := SingleStockCandle{SlopeShortDuration: 1.0, SlopeMedDuration: 2.0, SlopeLongDuration: 3.0}
	if got := getSlope(c, SHORTDURATION); got != 1.0 {
		t.Errorf("Short: got %v want 1.0", got)
	}
	if got := getSlope(c, MEDIUMDURATION); got != 2.0 {
		t.Errorf("Med: got %v want 2.0", got)
	}
	if got := getSlope(c, LONGDURATION); got != 3.0 {
		t.Errorf("Long: got %v want 3.0", got)
	}
	var w SingleStockCandle
	setSlope(&w, SHORTDURATION, 1.0)
	setSlope(&w, MEDIUMDURATION, 2.0)
	setSlope(&w, LONGDURATION, 3.0)
	if w.SlopeShortDuration != 1.0 || w.SlopeMedDuration != 2.0 || w.SlopeLongDuration != 3.0 {
		t.Errorf("setSlope: Short=%v Med=%v Long=%v", w.SlopeShortDuration, w.SlopeMedDuration, w.SlopeLongDuration)
	}
}
