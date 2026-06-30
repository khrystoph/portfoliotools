package universe

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

// ErrNotFound is returned when a resource does not exist on the remote API.
var ErrNotFound = errors.New("not found")

// PolygonAsset is a financial instrument record from the Polygon reference API.
type PolygonAsset struct {
	Ticker          string  `json:"ticker"`
	Name            string  `json:"name"`
	Market          string  `json:"market"`
	Type            string  `json:"type"`
	CurrencyName    string  `json:"currency_name"`
	MarketCap       float64 `json:"market_cap"`
	CompositeFIGI   string  `json:"composite_figi"`
	ShareClassFIGI  string  `json:"share_class_figi"`
	Active          bool    `json:"active"`
	PrimaryExchange string  `json:"primary_exchange"`
}

// PolygonDiscoverer fetches the active ticker universe from Polygon.
type PolygonDiscoverer interface {
	FetchActiveAssets(ctx context.Context, mcapThreshold int64) ([]PolygonAsset, error)
	FetchTickerDetails(ctx context.Context, symbol string) (PolygonAsset, error)
}

// PolygonClient calls the Polygon.io reference API.
type PolygonClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewPolygonClient(baseURL, apiKey string) *PolygonClient {
	return &PolygonClient{baseURL: baseURL, apiKey: apiKey, httpClient: &http.Client{}}
}

func (c *PolygonClient) FetchActiveAssets(ctx context.Context, mcapThreshold int64) ([]PolygonAsset, error) {
	nextURL := c.baseURL + "/v3/reference/tickers?active=true&market=stocks&order=asc&limit=250&apiKey=" + url.QueryEscape(c.apiKey)
	var all []PolygonAsset

	for nextURL != "" {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("build polygon request: %w", err)
		}
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("polygon request: %w", err)
		}

		var page struct {
			Results []PolygonAsset `json:"results"`
			NextURL string         `json:"next_url"`
			Status  string         `json:"status"`
		}
		decErr := json.NewDecoder(resp.Body).Decode(&page)
		resp.Body.Close()
		if decErr != nil {
			return nil, fmt.Errorf("decode polygon response: %w", decErr)
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("polygon unexpected status %d", resp.StatusCode)
		}

		for _, a := range page.Results {
			if int64(a.MarketCap) >= mcapThreshold {
				all = append(all, a)
			}
		}

		if page.NextURL != "" {
			nextURL = page.NextURL + "&apiKey=" + url.QueryEscape(c.apiKey)
		} else {
			nextURL = ""
		}
	}
	return all, nil
}

func (c *PolygonClient) FetchTickerDetails(ctx context.Context, symbol string) (PolygonAsset, error) {
	endpoint := fmt.Sprintf("%s/v3/reference/tickers/%s?apiKey=%s",
		c.baseURL, url.PathEscape(symbol), url.QueryEscape(c.apiKey))
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return PolygonAsset{}, fmt.Errorf("build polygon details request: %w", err)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return PolygonAsset{}, fmt.Errorf("polygon details request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return PolygonAsset{}, fmt.Errorf("polygon ticker %s: %w", symbol, ErrNotFound)
	}
	if resp.StatusCode != http.StatusOK {
		return PolygonAsset{}, fmt.Errorf("polygon details unexpected status %d", resp.StatusCode)
	}

	var result struct {
		Results PolygonAsset `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return PolygonAsset{}, fmt.Errorf("decode polygon details: %w", err)
	}
	return result.Results, nil
}
