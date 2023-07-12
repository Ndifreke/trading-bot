package executor

import (
	"fmt"
	"time"
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
ID                : %d
Status            : %s
Time              : %s
`,
		action.String(),
		order.Symbol,
		symbol.FormatQuotePrice(marketPrice),
		symbol.FormatQuotePrice(tradeStartPrice),
		order.Price,
		symbol.FormatQuotePrice(currentPrice),
		symbol.FormatBasePrice(profit),
		fee.String,
		order.ExecutedQuantity,
		order.OrderID,
		order.Status,
		time.Now().Format(time.UnixDate),
	)
	utils.LogInfo(sm)
	return sm
}
