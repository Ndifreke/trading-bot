package limit

import (
	"fmt"
	"trading/binance"
	"trading/request"
	"trading/trade"

	"trading/trade/executor"
	"trading/utils"
)

type LimitTrade struct {
	config         trade.TradeConfig
	preTradePrices map[string]float32
}

func NewLimitTrade(config trade.TradeConfig) trade.TradeRunner {
	return LimitTrade{
		config: config,
	}
}

// checks if the current market price is equal to or greater/less than a specified stop price,
// based on a given trade configuration.
func isStopPrice(t trade.TradeConfig, stopPrice, currentPrice float32) bool {
	if t.Action == trade.TradeActionBuy {
		//Introduce charges
		return currentPrice <= stopPrice
	} else if t.Action == trade.TradeActionSell {
		return currentPrice >= stopPrice
	}
	panic(fmt.Sprintf("Unknon trade action %s", t.Action))
}

func (t LimitTrade) Run() {
	if t.config.Action.IsBuy() {
		t.BuyRun()
		return
	} else if t.config.Action.IsSell() {
		t.SellRun()
		return
	}
	panic("Unkown Trade Action Type")
}

type LimitPredicateFun = func(ticker binance.MiniTickerData, tradingConfig trade.TradeConfig) bool

func (t LimitTrade) RunB(predicateFunc LimitPredicateFun) {
	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {
		if predicateFunc(message, t.config) {
			conn.Close()
		}
	}
	binance.PriceStream(t.config.Symbol).ReadMessage(handleReadMessage)
}

func getPretradePrices(t LimitTrade) map[string]float32 {
	preTradePrices := trade.GetSymbolPrices(t.config.Symbol)
	t.preTradePrices = preTradePrices
	utils.LogInfo(fmt.Sprintf("Pre Trade prices %v", preTradePrices))
	return preTradePrices
}

func isBuyStop(stopPrice, currentPrice float32) bool {
	return currentPrice <= stopPrice
}
func (t LimitTrade) BuyRun() {
	var preTradePrices map[string]float32

	if t.config.Price.Sell.Type.IsPercent() {
		preTradePrices = getPretradePrices(t)
	}

	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {
		buyStopPrice := trade.CalculateTradeBuyPrice(t.config.Price.Buy, preTradePrices[message.Data.Symbol])
		utils.LogWarn(fmt.Sprintf("Buy Stop Price is %f", buyStopPrice))
		utils.LogInfo(fmt.Sprintf("Current Buy Price is %v", message.Data.ClosePrice))
		utils.LogInfo(fmt.Sprintf("Buy Pretrade Price is %v", preTradePrices))

		if isBuyStop(buyStopPrice, message.Data.ClosePrice) {
			//Executor will handle the Sell or Buy of this Asset
			executor.BuyExecutor(
				t.config,
				message.Data.ClosePrice,
				conn,
				executor.ExecutorExtra{
					PreTradePrice: preTradePrices[message.Data.Symbol],
					Trader:        NewLimitTrade,
				}).Buy()
		}
	}
	binance.PriceStream(t.config.Symbol).ReadMessage(handleReadMessage)

}

func isSellStop(stopPrice, currentPrice float32) bool {
	return currentPrice >= stopPrice
}

func (t LimitTrade) SellRun() {

	var preTradePrices map[string]float32
	if t.config.Price.Sell.Type.IsPercent() {
		preTradePrices = getPretradePrices(t)
	}

	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {
		sellStopPrice := trade.CalculateTradeSellPrice(t.config.Price.Sell, preTradePrices[message.Data.Symbol])
		utils.LogWarn(fmt.Sprintf("Sell Stop Price is %f", sellStopPrice))
		utils.LogInfo(fmt.Sprintf("Current Sell Price is %v", message.Data.ClosePrice))
		utils.LogInfo(fmt.Sprintf("Sell Pretrade Price is %v", preTradePrices))

		if isSellStop(sellStopPrice, message.Data.ClosePrice) {
			//Executor will handle the Sell or Buy of this Asset
			executor.SellExecutor(
				t.config,
				message.Data.ClosePrice,
				conn,
				executor.ExecutorExtra{
					PreTradePrice: preTradePrices[message.Data.Symbol],
					Trader:        NewLimitTrade,
				}).Sell()
		}
	}
	binance.PriceStream(t.config.Symbol).ReadMessage(handleReadMessage)

}
