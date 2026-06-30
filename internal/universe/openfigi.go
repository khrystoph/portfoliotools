package universe

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/khrystoph/portfoliotools/internal/store"
)

// FIGIResult is the outcome of a single FIGI lookup.
type FIGIResult struct {
	CompositeFIGI string
	Ticker        string
	Found         bool
}

// FIGIChange records a ticker symbol change detected via FIGI comparison.
type FIGIChange struct {
	OldSymbol     string
	NewSymbol     string
	CompositeFIGI string
}

// FIGIResolver maps known composite FIGIs to their current ticker symbols.
type FIGIResolver interface {
	MapFIGIs(ctx context.Context, figis []string) ([]FIGIResult, error)
}

// OpenFIGIClient calls the OpenFIGI /v3/mapping endpoint.
type OpenFIGIClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

func NewOpenFIGIClient(baseURL, apiKey string) *OpenFIGIClient {
	return &OpenFIGIClient{baseURL: baseURL, apiKey: apiKey, httpClient: &http.Client{}}
}

func (c *OpenFIGIClient) MapFIGIs(ctx context.Context, figis []string) ([]FIGIResult, error) {
	type reqItem struct {
		IDType  string `json:"idType"`
		IDValue string `json:"idValue"`
	}
	items := make([]reqItem, len(figis))
	for i, f := range figis {
		items[i] = reqItem{IDType: "COMPOSITE_GLOBAL_STANDARD", IDValue: f}
	}

	body, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal openfigi request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		c.baseURL+"/v3/mapping", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("build openfigi request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		req.Header.Set("X-OPENFIGI-APIKEY", c.apiKey)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("openfigi request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("openfigi rate limited (429)")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("openfigi unexpected status %d", resp.StatusCode)
	}

	type dataItem struct {
		CompositeFigi string `json:"compositeFigi"`
		Ticker        string `json:"ticker"`
	}
	type resultItem struct {
		Data  []dataItem `json:"data"`
		Error string     `json:"error"`
	}
	var raw []resultItem
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("decode openfigi response: %w", err)
	}

	results := make([]FIGIResult, len(figis))
	for i := range figis {
		results[i].CompositeFIGI = figis[i]
		if i < len(raw) && len(raw[i].Data) > 0 {
			results[i].Ticker = raw[i].Data[0].Ticker
			results[i].Found = true
		}
	}
	return results, nil
}

// DetectChanges classifies known tickers against OpenFIGI results.
// Pure function: no DB access, no HTTP calls.
func DetectChanges(known []store.Ticker, results []FIGIResult) (renames []FIGIChange, delistings []store.Ticker) {
	byFIGI := make(map[string]FIGIResult, len(results))
	for _, r := range results {
		byFIGI[r.CompositeFIGI] = r
	}
	for _, tk := range known {
		if tk.CompositeFIGI == nil {
			continue
		}
		r, ok := byFIGI[*tk.CompositeFIGI]
		if !ok || !r.Found {
			delistings = append(delistings, tk)
			continue
		}
		if r.Ticker != tk.Symbol {
			renames = append(renames, FIGIChange{
				OldSymbol:     tk.Symbol,
				NewSymbol:     r.Ticker,
				CompositeFIGI: *tk.CompositeFIGI,
			})
		}
	}
	return renames, delistings
}
