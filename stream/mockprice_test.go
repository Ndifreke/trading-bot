package stream

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMockGen_GetNextPrice(t *testing.T) {
	// Create a mockGen struct with some initial price data.
	priceSource := map[string][]float64{
		"BTC": {10000, 11000, 12000},
		"ETH": {2000, 2200, 2400},
	}
	mockGen := newMockPrice([]string{"BTC", "ETH"}, priceSource)

	// Get the next price for BTC and ETH.
	btcPrice := mockGen.GetNextPrice("BTC")
	ethPrice := mockGen.GetNextPrice("ETH")

	// Assert that the next prices are correct.
	assert.Equal(t, 10000.0, btcPrice)
	assert.Equal(t, 2000.0, ethPrice)

	// Get the next price for BTC again.
	btcPrice = mockGen.GetNextPrice("BTC")
	ethPrice = mockGen.GetNextPrice("ETH")

	assert.Equal(t, 11000.0, btcPrice)
	assert.Equal(t, 2200.0, ethPrice)

	// Get the next price for BTC again.
	btcPrice = mockGen.GetNextPrice("BTC")
	ethPrice = mockGen.GetNextPrice("ETH")
	assert.Equal(t, 12000.0, btcPrice)
	assert.Equal(t, 2400.0, ethPrice)

	// Get the next price for BTC again.
	btcPrice = mockGen.GetNextPrice("BTC")
	ethPrice = mockGen.GetNextPrice("ETH")
	assert.Equal(t, 10000.0, btcPrice)
	assert.Equal(t, 2000.0, ethPrice)
}

func TestMockGen_GetNextPrice2(t *testing.T) {
	// Create a mockGen struct with some initial price data.
	priceSource := map[string][]float64{
		"BTC": {10000, 11000, 12000},
		"ETH": {2000, 2200, 2400},
	}
	mockGen := newMockPrice([]string{"BTC", "ETH"}, priceSource)

	// Get the next prices for all symbols.
	symbolPrices := mockGen.GetNextPrice2()

	// Assert that the next prices are correct.
	assert.Equal(t, 10000.0, symbolPrices["BTC"])
	assert.Equal(t, 2000.0, symbolPrices["ETH"])
}

func TestMockGen_NextPrices(t *testing.T) {
	// Create a mockGen struct with some initial price data.
	priceSource := map[string][]float64{
		"BTC": {10000, 11000, 12000},
		"ETH": {2000, 2200, 2400},
	}
	mockGen := newMockPrice([]string{"BTC", "ETH"}, priceSource)

	// Get the next prices for all symbols.
	symbolPrices := mockGen.NextPrices()

	// Assert that the next prices are correct.
	assert.Equal(t, 10000.0, symbolPrices["BTC"])
	assert.Equal(t, 2000.0, symbolPrices["ETH"])
}

func TestNewMockPriceA(t *testing.T) {
	// Create a mockGen struct with some initial price data.
	priceSource := map[string][]float64{
		"BTC": {10000, 11000, 12000},
		"ETH": {2000, 2200, 2400},
	}
	mockGen := newMockPrice([]string{"BTC", "ETH"}, priceSource)

	// Assert that the mockGen struct was created correctly.
	assert.Equal(t, len(priceSource), len(mockGen.priceMap))
	for symbol, value := range mockGen.priceMap {
		assert.Equal(t, priceSource[symbol], value.Prices)
	}
}

func TestGetNextPrice(t *testing.T) {
	priceSource := map[string][]float64{
		"ABC": {1.0, 2.0, 3.0},
		"XYZ": {10.0, 20.0},
	}
	mock := newMockPrice([]string{"ABC", "XYZ"}, priceSource)

	// Test GetNextPrice for symbol "ABC"
	price := mock.GetNextPrice("ABC")
	assert.Equal(t, 1.0, price)
	price = mock.GetNextPrice("ABC")
	assert.Equal(t, 2.0, price)
	price = mock.GetNextPrice("ABC")
	assert.Equal(t, 3.0, price)
	price = mock.GetNextPrice("ABC") // Wrap around
	assert.Equal(t, 1.0, price)

	// Test GetNextPrice for symbol "XYZ"
	price = mock.GetNextPrice("XYZ")
	assert.Equal(t, 10.0, price)
	price = mock.GetNextPrice("XYZ")
	assert.Equal(t, 20.0, price)
	price = mock.GetNextPrice("XYZ") // Wrap around
	assert.Equal(t, 10.0, price)
}

func TestGetNextPrice2(t *testing.T) {
	priceSource := map[string][]float64{
		"ABC": {1.0, 2.0, 3.0},
		"XYZ": {10.0, 20.0},
	}
	mock := newMockPrice([]string{"ABC", "XYZ"}, priceSource)

	// Test GetNextPrice2
	prices := mock.GetNextPrice2()
	assert.Equal(t, 1.0, prices["ABC"])
	assert.Equal(t, 10.0, prices["XYZ"])
}

func TestNextPrices(t *testing.T) {
	priceSource := map[string][]float64{
		"ABC": {1.0, 2.0, 3.0},
		"XYZ": {10.0, 20.0},
	}
	mock := newMockPrice([]string{"ABC", "XYZ"}, priceSource)

	// Test NextPrices
	prices := mock.NextPrices()
	assert.Equal(t, 1.0, prices["ABC"])
	assert.Equal(t, 10.0, prices["XYZ"])
	
	prices = mock.NextPrices()
	assert.Equal(t, 2.0, prices["ABC"])
	assert.Equal(t, 20.0, prices["XYZ"])

	prices = mock.NextPrices()
	assert.Equal(t, 3.0, prices["ABC"])
	assert.Equal(t, 10.0, prices["XYZ"])
}

func TestNewMockPrice(t *testing.T) {
	priceSource := map[string][]float64{
		"ABC": {1.0, 2.0, 3.0},
		"XYZ": {10.0, 20.0},
	}
	mock := newMockPrice([]string{"ABC", "XYZ"}, priceSource)

	// Test that the price map is initialized correctly
	assert.Equal(t, 0, mock.priceMap["ABC"].Iteration)
	assert.Equal(t, priceSource["ABC"], mock.priceMap["ABC"].Prices)
	assert.Equal(t, 0, mock.priceMap["XYZ"].Iteration)
	assert.Equal(t, priceSource["XYZ"], mock.priceMap["XYZ"].Prices)
}
