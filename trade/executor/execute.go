package executor

import (
	"fmt"
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
	config names.TradeConfig
	fees   helper.TradeFee
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
