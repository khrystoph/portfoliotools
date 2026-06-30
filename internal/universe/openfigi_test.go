package universe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/store"
	"github.com/khrystoph/portfoliotools/internal/universe"
)

func ptr(s string) *string { return &s }

func TestDetectChanges_NoChange(t *testing.T) {
	known := []store.Ticker{
		{Symbol: "AAPL", CompositeFIGI: ptr("BBG000B9XRY4")},
	}
	results := []universe.FIGIResult{
		{CompositeFIGI: "BBG000B9XRY4", Ticker: "AAPL", Found: true},
	}
	renames, delistings := universe.DetectChanges(known, results)
	assert.Empty(t, renames)
	assert.Empty(t, delistings)
}

func TestDetectChanges_Rename(t *testing.T) {
	known := []store.Ticker{
		{Symbol: "OSTK", CompositeFIGI: ptr("BBG000BL1Q17")},
	}
	results := []universe.FIGIResult{
		{CompositeFIGI: "BBG000BL1Q17", Ticker: "BBBY", Found: true},
	}
	renames, delistings := universe.DetectChanges(known, results)
	require.Len(t, renames, 1)
	assert.Equal(t, "OSTK", renames[0].OldSymbol)
	assert.Equal(t, "BBBY", renames[0].NewSymbol)
	assert.Empty(t, delistings)
}

func TestDetectChanges_Delisting(t *testing.T) {
	known := []store.Ticker{
		{Symbol: "DEAD", CompositeFIGI: ptr("BBG0000DEAD3")},
	}
	results := []universe.FIGIResult{
		{CompositeFIGI: "BBG0000DEAD3", Found: false},
	}
	renames, delistings := universe.DetectChanges(known, results)
	assert.Empty(t, renames)
	require.Len(t, delistings, 1)
	assert.Equal(t, "DEAD", delistings[0].Symbol)
}

func TestDetectChanges_Mixed(t *testing.T) {
	known := []store.Ticker{
		{Symbol: "AAPL", CompositeFIGI: ptr("BBG000B9XRY4")},
		{Symbol: "OSTK", CompositeFIGI: ptr("BBG000BL1Q17")},
		{Symbol: "DEAD", CompositeFIGI: ptr("BBG0000DEAD3")},
	}
	results := []universe.FIGIResult{
		{CompositeFIGI: "BBG000B9XRY4", Ticker: "AAPL", Found: true},
		{CompositeFIGI: "BBG000BL1Q17", Ticker: "BBBY", Found: true},
		{CompositeFIGI: "BBG0000DEAD3", Found: false},
	}
	renames, delistings := universe.DetectChanges(known, results)
	require.Len(t, renames, 1)
	assert.Equal(t, "OSTK", renames[0].OldSymbol)
	require.Len(t, delistings, 1)
	assert.Equal(t, "DEAD", delistings[0].Symbol)
}

func TestDetectChanges_EmptyInputs(t *testing.T) {
	renames, delistings := universe.DetectChanges(nil, nil)
	assert.Empty(t, renames)
	assert.Empty(t, delistings)
}

func TestDetectChanges_StaticAssetsSkipped(t *testing.T) {
	known := []store.Ticker{
		{Symbol: "GC=F", CompositeFIGI: nil},
	}
	results := []universe.FIGIResult{}
	renames, delistings := universe.DetectChanges(known, results)
	assert.Empty(t, renames)
	assert.Empty(t, delistings)
}

func TestOpenFIGIClient_MapFIGIs_Success(t *testing.T) {
	body, err := os.ReadFile("testdata/openfigi/map_figis_success.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/v3/mapping", r.URL.Path)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer srv.Close()

	client := universe.NewOpenFIGIClient(srv.URL, "")
	results, err := client.MapFIGIs(context.Background(), []string{"BBG000B9XRY4"})
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "AAPL", results[0].Ticker)
	assert.True(t, results[0].Found)
}

func TestOpenFIGIClient_MapFIGIs_RateLimit(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	client := universe.NewOpenFIGIClient(srv.URL, "")
	_, err := client.MapFIGIs(context.Background(), []string{"BBG000B9XRY4"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "rate limited")
}
