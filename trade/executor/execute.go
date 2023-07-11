package executor

import (
	"fmt"
	"math"
	"trading/helper"
	"trading/names"
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

type ExecutorInterface interface {
	IsProfitable() bool
	Execute() bool
}

type executorType struct {
	marketPrice     float64
	tradeStartPrice float64
	// extra  names.ExecutorExtraType
	config names.TradeConfig
	fees   helper.TradeFee
}

// quoteAmount the total amount that will be exchanged for the base asset
// baseUnitPrice the price of one unit of base asset in quote value
func getBuyQuote(quoteAmount, baseUnitPrice float64) float64 {
	quantity := quoteAmount / baseUnitPrice
	var decimalPlaces int
	if baseUnitPrice != 0 {
		decimalPlaces = int(math.Floor(math.Log10(math.Abs(baseUnitPrice))) + 1)
	}
	scale := math.Pow(10, float64(decimalPlaces))
	truncated := math.Trunc(quantity*scale) / scale
	return truncated
}

func summary(action names.TradeSide, symbol names.Symbol, marketPrice, tradeStartPrice, currentPrice, profit float64, fee helper.TradeFee, quantity float64, order binance.CreateOrderResponse) string {
	_profit := symbol.FormatBasePrice(profit)
	_quantity := symbol.FormatBasePrice(float64(quantity))
	_beforeTradePrice := symbol.FormatQuotePrice(tradeStartPrice)
	_lastTradePrice := symbol.FormatQuotePrice(marketPrice)
	_tradedPrice := symbol.FormatQuotePrice(currentPrice)

	if action.IsBuy() {

	}
	sm := fmt.Sprintf(
		`
===== %s TRADE SUMMARY =====
Symbol            : %s
Last Trade Price  : %s
Started Trade     : %s
Traded Price      : %s
Ticker Price      : %s
Profit            : %s
Calculated fee    : %s
Quantity          : %s
ID                : %s
Status            : %s
`,
		action.String(),
		symbol.String(),
		_lastTradePrice,
		_beforeTradePrice,
		order.Price,
		_tradedPrice,
		_profit,
		fee.String,
		_quantity,
		order.OrderID,
		order.Status,
	)
	utils.LogInfo(sm)
	return sm
}
