package universe_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khrystoph/portfoliotools/internal/universe"
)

func TestAlpacaClient_ListActiveSymbols(t *testing.T) {
	body, err := os.ReadFile("testdata/alpaca/assets_active.json")
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "APCA-API-KEY-ID", "APCA-API-KEY-ID")
		w.Header().Set("Content-Type", "application/json")
		w.Write(body)
	}))
	defer srv.Close()

	client := universe.NewAlpacaClient(srv.URL, "key", "secret")
	symbols, err := client.ListActiveSymbols(context.Background())
	require.NoError(t, err)
	assert.Contains(t, symbols, "AAPL")
	assert.Contains(t, symbols, "SPY")
	assert.Len(t, symbols, 2)
}
