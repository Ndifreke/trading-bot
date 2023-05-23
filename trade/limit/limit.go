package limit

import (
	"fmt"
	"trading/binance"
	"trading/request"
	"trading/trade"

	"trading/helper"
	"trading/trade/executor"
	"trading/utils"
)

type LimitTradeManager struct {
	tradeConfigs []trade.TradeConfig
	socket       *request.Socket[binance.MiniTickerData]
}

func NewLimitTradeManager(tradeConfigs ...trade.TradeConfig) trade.Trader {
	return &LimitTradeManager{
		tradeConfigs: tradeConfigs,
		socket:       nil,
	}
}

// checks if the current market price is equal to or greater/less than a specified stop price,
// based on a given trade configuration.
func isStopPrice(t trade.TradeConfig, stopPrice, currentPrice float32) bool {
	if t.Side == trade.TradeSideBuy {
		//Introduce charges
		return currentPrice <= stopPrice
	} else if t.Side == trade.TradeSideSell {
		return currentPrice >= stopPrice
	}
	panic(fmt.Sprintf("Unknon trade action %s", t.Side))
}

func (t *LimitTradeManager) getConfig(symbol trade.Symbol) trade.TradeConfig {
	for _, tc := range t.tradeConfigs {
		if tc.Symbol == symbol {
			return tc
		}
	}
	return trade.TradeConfig{}
}

func (t *LimitTradeManager) getTradeSymbols() []trade.Symbol {
	var s []trade.Symbol
	for _, tc := range t.tradeConfigs {
		s = append(s, tc.Symbol)
	}
	return s
}

func (t *LimitTradeManager) Run() {
	//can be passed as arg
	socket := binance.PriceStream(t.getTradeSymbols())
	// t.socket = socket
	// for _, tc := range t.tradeConfigs {
	// 	if tc.Action.IsBuy() {
	// 		BuyRun(trade.TradeRunner[binance.MiniTickerData]{
	// 			Config: tc,
	// 			Socket: socket,
	// 		})
	// 	} else if tc.Action.IsSell() {
	// 		SellRun(trade.TradeRunner[binance.MiniTickerData]{
	// 			Config: tc,
	// 			Socket: socket,
	// 		})
	// 	}
	// }
	// t.socket.
	// 	SetIdGetter(
	// 		func(d binance.MiniTickerData) string {
	// 			return d.Data.Symbol
	// 		}).
	// 	SubscribeReaders()
	t.RunWithConnection(socket)
}

func (t *LimitTradeManager) RunWithConnection(socket *request.Socket[binance.MiniTickerData]) {
	//can be passed as arg
	// socket := binance.PriceStream(t.getTradeSymbols())
	t.socket = socket
	for _, tc := range t.tradeConfigs {
		if tc.Side.IsBuy() {
			BuyRun(trade.TradeRunner[binance.MiniTickerData]{
				Config: tc,
				Socket: socket,
			})
		} else if tc.Side.IsSell() {
			SellRun(trade.TradeRunner[binance.MiniTickerData]{
				Config: tc,
				Socket: socket,
			})
		}
	}
	t.socket.
		SetIdGetter(
			func(d binance.MiniTickerData) string {
				return d.Data.Symbol
			}).
		SubscribeReaders()
}

type LimitPredicateFun = func(ticker binance.MiniTickerData, tradingConfig trade.TradeConfig) bool

// func (t LimitTrade) RunB(predicateFunc LimitPredicateFun) {
// 	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {
// 		if predicateFunc(message, t.config) {
// 			conn.Close()
// 		}
// 	}
// 	binance.PriceStream(t.config.Symbol).ReadMessage(handleReadMessage)
// }

// func getPretradePrices(t LimitTrade) map[string]float32 {
// 	preTradePrices := trade.GetSymbolPrices(t.config.Symbol)
// 	t.preTradePrices = preTradePrices
// 	utils.LogInfo(fmt.Sprintf("Pre Trade prices %v", preTradePrices))
// 	return preTradePrices
// }

func isBuyStop(stopPrice, currentPrice float32) bool {
	return currentPrice <= stopPrice
}

// The BuyRun completes a trade by buying when the current price
// is lower than the last traded price for this pair. Where the stop
// price is a fixed or a lower percentile of the last traded price of the pair
func BuyRun(t trade.TradeRunner[binance.MiniTickerData]) {
	// var preTradePrice map[string]float32
	// if t.Price.Sell.RateIn.IsPercent() {
	config := t.Config
	priceAtRun := binance.GetPriceLatest(config.Symbol.String()).Body.Price
	// }

	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {

		buyStopPrice := helper.CalculateTradeBuyPrice(config.Price.Buy, priceAtRun)
		utils.LogWarn(fmt.Sprintf("%s Buy Stop Price is %f", message.Data.Symbol, buyStopPrice))
		utils.LogInfo(fmt.Sprintf("%s Current Buy Price is %v", message.Data.Symbol, message.Data.ClosePrice))
		utils.LogInfo(fmt.Sprintf("%s Buy Pretrade Price is %v", message.Data.Symbol, priceAtRun))

		//Implement a guard that allows for price to go lower before buying despite stop pricc
		// {lastStop: either in percent allways updated after every of the percent decrease}
		if isBuyStop(buyStopPrice, message.Data.ClosePrice) {
			//Executor will handle the Sell or Buy of this Asset
			executor.BuyExecutor(
				config,
				message.Data.ClosePrice,
				message.Data.Symbol,
				conn,
				executor.ExecutorExtra{
					PreTradePrice: priceAtRun,
					TradeManager:  NewLimitTradeManager,
				}).Execute()
		}
	}
	t.Socket.RegisterReader(config.Symbol.String(), handleReadMessage)

}

func isSellStop(stopPrice, currentPrice float32) bool {
	fmt.Println(int(currentPrice) >= int(stopPrice))
	return int(currentPrice) >= int(stopPrice)
}

func SellRun(t trade.TradeRunner[binance.MiniTickerData]) {
	config := t.Config
	priceAtRun := binance.GetPriceLatest(config.Symbol.String()).Body.Price

	var handleReadMessage = func(conn request.Connection, message binance.MiniTickerData) {

		sellStopPrice := helper.CalculateTradeSellPrice(config.Price.Sell, priceAtRun)
		utils.LogWarn(fmt.Sprintf("%s Sell Stop Price is %f", message.Data.Symbol, sellStopPrice))
		utils.LogInfo(fmt.Sprintf("%s Current Sell Price is %v", message.Data.Symbol, message.Data.ClosePrice))
		utils.LogInfo(fmt.Sprintf("%s Sell Pretrade Price is %v", message.Data.Symbol, priceAtRun))

		if isSellStop(sellStopPrice, message.Data.ClosePrice) {
			//Executor will handle the Sell or Buy of this Asset
			executor.SellExecutor(
				config,
				message.Data.ClosePrice,
				conn,
				executor.ExecutorExtra{
					PreTradePrice: priceAtRun,
					TradeManager:  NewLimitTradeManager,
				}).Execute()
		}
	}
	t.Socket.RegisterReader(config.Symbol.String(), handleReadMessage)
}
