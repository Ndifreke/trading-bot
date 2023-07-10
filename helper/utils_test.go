package helper

import (
	"testing"
	"trading/names"

	"github.com/stretchr/testify/assert"
)

var sellConfigPercent = names.TradeConfig{
	Symbol: "BTCUSD",
	Side:   names.TradeSideSell,
	Sell: names.SideConfig{
		RateLimit: 3,
		RateType:  names.RatePercent,
		Quantity:  2,
		LockDelta: 10,
	},
}

var sellConfigFixed = names.TradeConfig{
	Symbol: "BTCUSD",
	Side:   names.TradeSideSell,
	Sell: names.SideConfig{
		RateLimit: 55,
		RateType:  names.RateFixed,
		Quantity:  2,
		LockDelta: 10,
	},
}

var buyConfigPercent = names.TradeConfig{
	Symbol: "ETHUSDC",
	Side:   names.TradeSideBuy,
	Buy: names.SideConfig{
		RateLimit: 5,
		RateType:  names.RatePercent,
		Quantity:  2,
		LockDelta: 2,
	},
}

var buyConfigFixed = names.TradeConfig{
	Symbol: "ETHUSDC",
	Side:   names.TradeSideBuy,
	Buy: names.SideConfig{
		RateLimit: 5,
		RateType:  names.RateFixed,
		Quantity:  2,
		LockDelta: 2,
	},
}

func TestCalculateTradePrice(t *testing.T) {
	sellpercentPrice := CalculateTradePrice(sellConfigPercent, 250)
	assert.EqualValues(t, sellpercentPrice.Limit, 257.5, "calclulate sell fixed price from percentage rateLimit")
	assert.EqualValues(t, sellpercentPrice.Percent, 3, "calclulate sell percentage price from percentage rateLimit")

	sellFixedPrice := CalculateTradePrice(sellConfigFixed, 50)
	assert.EqualValues(t, sellFixedPrice.Limit, 55, "calclulate sell fixed price fixed rateLimit")
	assert.EqualValues(t, sellFixedPrice.Percent, 10, "calclulate sell percent price from fixed rateLimit")

	buyPercentPrice := CalculateTradePrice(buyConfigPercent, 50)
	assert.EqualValues(t, buyPercentPrice.Limit, 47.5, "calclulate buy fixed price from percentage rateLimit")
	assert.EqualValues(t, buyPercentPrice.Percent, 5, "calclulate buy percentage price from percentage rateLimit")

	buyFixedPrice := CalculateTradePrice(buyConfigFixed, 50)
	assert.EqualValues(t, buyFixedPrice.Limit, buyConfigFixed.Buy.RateLimit, "calclulate sell fixed price fixed rateLimit")
	assert.EqualValues(t, buyFixedPrice.Percent, 90, "calclulate sell percent price from fixed rateLimit")
}
