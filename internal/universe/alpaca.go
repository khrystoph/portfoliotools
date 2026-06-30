package universe

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// AlpacaAssetLister returns the set of symbols currently tradable on Alpaca.
type AlpacaAssetLister interface {
	ListActiveSymbols(ctx context.Context) (map[string]struct{}, error)
}

// AlpacaClient calls the Alpaca broker API.
type AlpacaClient struct {
	baseURL    string
	apiKey     string
	apiSecret  string
	httpClient *http.Client
}

func NewAlpacaClient(baseURL, apiKey, apiSecret string) *AlpacaClient {
	return &AlpacaClient{baseURL: baseURL, apiKey: apiKey, apiSecret: apiSecret, httpClient: &http.Client{}}
}

func (c *AlpacaClient) ListActiveSymbols(ctx context.Context) (map[string]struct{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		c.baseURL+"/v2/assets?status=active&asset_class=us_equity", nil)
	if err != nil {
		return nil, fmt.Errorf("build alpaca request: %w", err)
	}
	req.Header.Set("APCA-API-KEY-ID", c.apiKey)
	req.Header.Set("APCA-API-SECRET-KEY", c.apiSecret)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("alpaca request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("alpaca unexpected status %d", resp.StatusCode)
	}

	var assets []struct {
		Symbol string `json:"symbol"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&assets); err != nil {
		return nil, fmt.Errorf("decode alpaca assets: %w", err)
	}

	result := make(map[string]struct{}, len(assets))
	for _, a := range assets {
		result[a.Symbol] = struct{}{}
	}
	return result, nil
}
