package universe

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/khrystoph/portfoliotools/internal/store"
)

const defaultMcapThreshold = int64(100_000_000)

// BackoffConfig controls exponential backoff for sync jobs.
type BackoffConfig struct {
	InitialDelay time.Duration
	Multiplier   float64
	Cap          time.Duration
}

func defaultBackoffConfig() BackoffConfig {
	return BackoffConfig{
		InitialDelay: time.Minute,
		Multiplier:   2.0,
		Cap:          30 * time.Minute,
	}
}

// SyncConfig holds runtime parameters for the sync job.
type SyncConfig struct {
	McapThresholdUSD int64
	BackoffConfig    BackoffConfig
}

// LoadSyncConfig reads mcap_threshold_usd from system_config.
// Falls back to SYNC_MCAP_THRESHOLD env var, then the compiled default.
func LoadSyncConfig(ctx context.Context, sysConf *store.SystemConfigStore) (SyncConfig, error) {
	fallback := defaultMcapThreshold
	if v := os.Getenv("SYNC_MCAP_THRESHOLD"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			fallback = n
		}
	}
	threshold, err := sysConf.GetInt64(ctx, "mcap_threshold_usd", fallback)
	if err != nil {
		return SyncConfig{}, fmt.Errorf("load sync config: %w", err)
	}
	return SyncConfig{
		McapThresholdUSD: threshold,
		BackoffConfig:    defaultBackoffConfig(),
	}, nil
}
