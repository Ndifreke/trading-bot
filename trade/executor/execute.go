package executor

import (
	"fmt"
	"trading/helper"
	"trading/trade"
	"trading/utils"
)

type TradeExecutor interface {
	IsProfitable() bool
}

type ExecutorExtra struct {
	PreTradePrice float32
	Trader        func(t ...trade.TradeConfig) trade.TradeRunner
}

func summary(action trade.TradeAction, symbol trade.Symbol, lastTradePrice, beforeTradePrice, currentPrice, profit float32, fee helper.TradeFee, quantity int) string {
	_profit := symbol.FormatBasePrice(profit)
	_quantity := symbol.FormatBasePrice(float32(quantity))
	_beforeTradePrice := symbol.FormatQuotePrice(beforeTradePrice)
	_lastTradePrice := symbol.FormatQuotePrice(lastTradePrice)
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
Profit            : %s
Trade fee         : %s
Quantity          : %s
`,
		action.String(),
		symbol.String(),
		_lastTradePrice,
		_beforeTradePrice,
		_tradedPrice,
		_profit,
		fee.String,
		_quantity,
	)
	utils.LogInfo(sm)
	return sm
}
