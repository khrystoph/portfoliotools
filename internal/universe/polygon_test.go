package universe_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/universe"
)

func TestPolygonClient_FetchActiveAssets_Pagination(t *testing.T) {
	p1, err := os.ReadFile("testdata/polygon/reference_tickers_p1.json")
	require.NoError(t, err)
	p2, err := os.ReadFile("testdata/polygon/reference_tickers_p2.json")
	require.NoError(t, err)

	callCount := 0
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		callCount++
		if callCount == 1 {
			body := strings.ReplaceAll(string(p1), "PLACEHOLDER_REPLACED_IN_TEST", srv.URL+"/v3/reference/tickers?cursor=page2")
			w.Write([]byte(body))
		} else {
			w.Write(p2)
		}
	}))
	defer srv.Close()

	client := universe.NewPolygonClient(srv.URL, "test-key")
	assets, err := client.FetchActiveAssets(context.Background(), 100_000_000)
	require.NoError(t, err)
	assert.Len(t, assets, 2)
	assert.Equal(t, 2, callCount)
}

func TestPolygonClient_FetchActiveAssets_McapFilter(t *testing.T) {
	body := `{"results":[
		{"ticker":"BIG","name":"Big Co","market":"stocks","type":"CS","currency_name":"usd","market_cap":500000000.0,"composite_figi":"BBG0000BIG01","active":true},
		{"ticker":"TINY","name":"Tiny Co","market":"stocks","type":"CS","currency_name":"usd","market_cap":50000000.0,"composite_figi":"BBG0000TINY1","active":true}
	],"status":"OK","count":2}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(body))
	}))
	defer srv.Close()

	client := universe.NewPolygonClient(srv.URL, "test-key")
	assets, err := client.FetchActiveAssets(context.Background(), 100_000_000)
	require.NoError(t, err)
	require.Len(t, assets, 1)
	assert.Equal(t, "BIG", assets[0].Ticker)
}

func TestPolygonClient_FetchTickerDetails_Found(t *testing.T) {
	body, err := os.ReadFile("testdata/polygon/ticker_details_found.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "BBBY")
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer srv.Close()

	client := universe.NewPolygonClient(srv.URL, "test-key")
	asset, err := client.FetchTickerDetails(context.Background(), "BBBY")
	require.NoError(t, err)
	assert.Equal(t, "BBBY", asset.Ticker)
	assert.True(t, asset.Active)
}

func TestPolygonClient_FetchTickerDetails_NotFound(t *testing.T) {
	body, err := os.ReadFile("testdata/polygon/ticker_details_404.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write(body)
	}))
	defer srv.Close()

	client := universe.NewPolygonClient(srv.URL, "test-key")
	_, err = client.FetchTickerDetails(context.Background(), "GONE")
	require.Error(t, err)
	assert.ErrorIs(t, err, universe.ErrNotFound)
}

var _ = json.Marshal
