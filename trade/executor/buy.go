package executor

import (
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/utils"
)

type buyExecutor executorType

func BuyExecutor(
	config names.TradeConfig,
	marketPrice float64,
	tradeStartPrice float64,
	// basePrice float64,
	// connection request.Connection,
	// extra ExecutorExtraType,

) ExecutorInterface {
	return &buyExecutor{
		marketPrice,
		tradeStartPrice,
		config,
		helper.TradeFee{},
	}
}

func (exec buyExecutor) IsProfitable() bool {

	if !exec.config.Buy.MustProfit || exec.tradeStartPrice == 0 {
		// The Price we will be buying is less than the price when we started this
		// trade plus the charges for this trade buy and subsequent sell.
		// Note we may want to substitute the PriceAtRun to the last price that this
		// symbol was bought for that way we can have an accurate PriceAtRun called last
		// traded Price
		return true
	}
	return (exec.marketPrice + exec.fees.Value) < exec.tradeStartPrice
}

func buy(exec *buyExecutor) bool {
	//last trade price from API TODO note it could be
	// buy or sell action on this symbol. How do you calculate the profile
	lastTradePrice := exec.tradeStartPrice

	buyOrder, err := binance.CreateBuyMarketOrder(
		exec.config.Symbol,
		exec.marketPrice,
		exec.config.Sell.Quantity,
	)
	if err != nil {
		utils.LogError(err,"Error Buying")
		return false
	}
	summary(
		exec.config.Side,
		exec.config.Symbol,
		lastTradePrice,
		exec.tradeStartPrice,
		exec.marketPrice,
		exec.marketPrice-lastTradePrice,
		exec.fees,
		exec.config.Buy.Quantity,
		*buyOrder,
	)
	return true
}

func (exec *buyExecutor) Execute() bool {
	exec.fees = helper.GetTradeFee(exec.config, exec.marketPrice)

	// if !exec.IsProfitable() {
	// 	// Dont buy if user wanted to make profit by force
	// 	return false
	// }
	sold := buy(exec)
	return sold
}
