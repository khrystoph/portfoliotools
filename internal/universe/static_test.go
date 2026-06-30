package universe_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/khrystoph/portfoliotools/internal/universe"
)

func TestAllStaticAssets_NoDuplicates(t *testing.T) {
	all := universe.AllStaticAssets()
	seen := make(map[string]bool)
	for _, a := range all {
		assert.False(t, seen[a.Symbol], "duplicate symbol: %s", a.Symbol)
		assert.NotEmpty(t, a.Symbol, "empty symbol in static assets")
		assert.NotEmpty(t, a.Name, "empty name for symbol %s", a.Symbol)
		seen[a.Symbol] = true
	}
	assert.Greater(t, len(all), 40)
}
