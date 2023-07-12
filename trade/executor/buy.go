package executor

import (
	"fmt"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/user"
	"trading/utils"
)

type buyExecutor executorType

func BuyExecutor(
	config names.TradeConfig,
	marketPrice float64,
	tradeStartPrice float64,
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

	lastTradePrice := exec.tradeStartPrice
	quoteBalance := user.CreateUser().GetAccount().GetBalance(exec.config.Symbol.ParseTradingPair().Quote)
	preciseQuantity := exec.config.Sell.Quantity

	if preciseQuantity < 0 {
		preciseQuantity = names.GetSymbols().PreciseValue(exec.config.Symbol.String(), quoteBalance.Locked/exec.marketPrice)
	}

	fmt.Println() 
	buyOrder, err := binance.CreateBuyMarketOrder(
		exec.config.Symbol.String(),
		preciseQuantity,
	)
	if err != nil {
		utils.LogError(err, fmt.Sprintf("Error  Buying %s, Qty=%f", exec.config.Symbol, preciseQuantity))
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
