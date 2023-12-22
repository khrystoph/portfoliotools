package pkg

type MinipoolStats struct {
	effectivenessRate float64 `json:"effectiveness"`
	minipoolSize      float64 `json:"minipoolsize"`
}

/*CalculateSmoothingPoolRewards estimates the rewards for the current rewards period (every 28 days) based on minipools
*  currently in the smoothing pool. It does not currently account for when they entered the smoothing pool, rather their
*  effective share assuming all minipools were in the smoothing pool at the start of the rewards cycle all the way
*  through the end of the rewards cycle.
 */
func CalculateSmoothingPoolRewards(smoothingPoolRewards, currTotalMinipools, effectivenessRate, minipoolSize,
	commission float64) (currExpectedRewards float64, err error) {

	minipoolShareSize := float64(minipoolSize / 32)
	rewardsSharePerMinipool := minipoolShareSize + (1-minipoolShareSize)*commission
	currExpectedRewards = smoothingPoolRewards / currTotalMinipools * (rewardsSharePerMinipool) * effectivenessRate
	return
}

// CalculateTotSmoothingPoolRewards aggregates CalculateSmoothingPoolRewards to
func CalculateTotSmoothingPoolRewards() {

}
