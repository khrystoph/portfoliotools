package pkg

func getPrices(c SingleStockCandle, d int) map[string]float64 {
	switch d {
	case SHORTDURATION:
		return c.ShortPrices
	case MEDIUMDURATION:
		return c.MedPrices
	case LONGDURATION:
		return c.LongPrices
	}
	return nil
}

func setPrices(c *SingleStockCandle, d int, v map[string]float64) {
	switch d {
	case SHORTDURATION:
		c.ShortPrices = v
	case MEDIUMDURATION:
		c.MedPrices = v
	case LONGDURATION:
		c.LongPrices = v
	}
}

func getRVol(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.RealizedVolatilityShort
	case MEDIUMDURATION:
		return c.RealizedVolatilityMed
	case LONGDURATION:
		return c.RealizedVolatilityLong
	}
	return 0
}

func setRVol(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.RealizedVolatilityShort = v
	case MEDIUMDURATION:
		c.RealizedVolatilityMed = v
	case LONGDURATION:
		c.RealizedVolatilityLong = v
	}
}

func getAvgVol(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.AvgVolumeShort
	case MEDIUMDURATION:
		return c.AvgVolumeMed
	case LONGDURATION:
		return c.AvgVolumeLong
	}
	return 0
}

func setAvgVol(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.AvgVolumeShort = v
	case MEDIUMDURATION:
		c.AvgVolumeMed = v
	case LONGDURATION:
		c.AvgVolumeLong = v
	}
}

func getAvgVolRatio(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.AvgVolumeRatioShort
	case MEDIUMDURATION:
		return c.AvgVolumeRatioMed
	case LONGDURATION:
		return c.AvgVolumeRatioLong
	}
	return 0
}

func setAvgVolRatio(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.AvgVolumeRatioShort = v
	case MEDIUMDURATION:
		c.AvgVolumeRatioMed = v
	case LONGDURATION:
		c.AvgVolumeRatioLong = v
	}
}

func getRVolVel(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.VelocityRealizedVolShort
	case MEDIUMDURATION:
		return c.VelocityRealizedVolMed
	case LONGDURATION:
		return c.VelocityRealizedVolLong
	}
	return 0
}

func setRVolVel(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.VelocityRealizedVolShort = v
	case MEDIUMDURATION:
		c.VelocityRealizedVolMed = v
	case LONGDURATION:
		c.VelocityRealizedVolLong = v
	}
}

func getRVolAccel(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.RealizedVolAccelShort
	case MEDIUMDURATION:
		return c.RealizedVolAccelMed
	case LONGDURATION:
		return c.RealizedVolAccelLong
	}
	return 0
}

func setRVolAccel(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.RealizedVolAccelShort = v
	case MEDIUMDURATION:
		c.RealizedVolAccelMed = v
	case LONGDURATION:
		c.RealizedVolAccelLong = v
	}
}

func getRVolHigh(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.RVolHighShort
	case MEDIUMDURATION:
		return c.RVolHighMed
	case LONGDURATION:
		return c.RVolHighLong
	}
	return 0
}

func setRVolHigh(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.RVolHighShort = v
	case MEDIUMDURATION:
		c.RVolHighMed = v
	case LONGDURATION:
		c.RVolHighLong = v
	}
}

func getRVolLow(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.RVolLowShort
	case MEDIUMDURATION:
		return c.RVolLowMed
	case LONGDURATION:
		return c.RVolLowLong
	}
	return 0
}

func setRVolLow(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.RVolLowShort = v
	case MEDIUMDURATION:
		c.RVolLowMed = v
	case LONGDURATION:
		c.RVolLowLong = v
	}
}

func getRVolPercent(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.RVolPercentShort
	case MEDIUMDURATION:
		return c.RVolPercentMed
	case LONGDURATION:
		return c.RVolPercentLong
	}
	return 0
}

func setRVolPercent(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.RVolPercentShort = v
	case MEDIUMDURATION:
		c.RVolPercentMed = v
	case LONGDURATION:
		c.RVolPercentLong = v
	}
}

func getRiskRange(c SingleStockCandle, d int) map[string]float64 {
	switch d {
	case SHORTDURATION:
		return c.TradeRange
	case MEDIUMDURATION:
		return c.TrendRange
	case LONGDURATION:
		return c.TailRange
	}
	return nil
}

func setRiskRange(c *SingleStockCandle, d int, v map[string]float64) {
	switch d {
	case SHORTDURATION:
		c.TradeRange = v
	case MEDIUMDURATION:
		c.TrendRange = v
	case LONGDURATION:
		c.TailRange = v
	}
}

func getAdjRiskRange(c SingleStockCandle, d int) map[string]float64 {
	switch d {
	case SHORTDURATION:
		return c.TradeRangeAdj
	case MEDIUMDURATION:
		return c.TrendRangeAdj
	case LONGDURATION:
		return c.TailRangeAdj
	}
	return nil
}

func setAdjRiskRange(c *SingleStockCandle, d int, v map[string]float64) {
	switch d {
	case SHORTDURATION:
		c.TradeRangeAdj = v
	case MEDIUMDURATION:
		c.TrendRangeAdj = v
	case LONGDURATION:
		c.TailRangeAdj = v
	}
}

func getProbRiskRange(c SingleStockCandle, d int) map[string]float64 {
	switch d {
	case SHORTDURATION:
		return c.PTradeRange
	case MEDIUMDURATION:
		return c.PTrendRange
	case LONGDURATION:
		return c.PTailRange
	}
	return nil
}

func setProbRiskRange(c *SingleStockCandle, d int, v map[string]float64) {
	switch d {
	case SHORTDURATION:
		c.PTradeRange = v
	case MEDIUMDURATION:
		c.PTrendRange = v
	case LONGDURATION:
		c.PTailRange = v
	}
}

func getProbAdjRiskRange(c SingleStockCandle, d int) map[string]float64 {
	switch d {
	case SHORTDURATION:
		return c.PTradeRangeAdj
	case MEDIUMDURATION:
		return c.PTrendRangeAdj
	case LONGDURATION:
		return c.PTailRangeAdj
	}
	return nil
}

func setProbAdjRiskRange(c *SingleStockCandle, d int, v map[string]float64) {
	switch d {
	case SHORTDURATION:
		c.PTradeRangeAdj = v
	case MEDIUMDURATION:
		c.PTrendRangeAdj = v
	case LONGDURATION:
		c.PTailRangeAdj = v
	}
}

func getSlope(c SingleStockCandle, d int) float64 {
	switch d {
	case SHORTDURATION:
		return c.SlopeShortDuration
	case MEDIUMDURATION:
		return c.SlopeMedDuration
	case LONGDURATION:
		return c.SlopeLongDuration
	}
	return 0
}

func setSlope(c *SingleStockCandle, d int, v float64) {
	switch d {
	case SHORTDURATION:
		c.SlopeShortDuration = v
	case MEDIUMDURATION:
		c.SlopeMedDuration = v
	case LONGDURATION:
		c.SlopeLongDuration = v
	}
}
