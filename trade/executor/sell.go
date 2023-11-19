package executor

import (

	"trading/helper"
	"trading/names"
	"trading/user"
)

type sellExecutor executorType

func SellExecutor(
	config names.TradeConfig,
	marketPrice float64,
	tradeStartPrice float64,
) ExecutorInterface {

	return &sellExecutor{
		marketPrice,
		tradeStartPrice,
		config,
		helper.TradeFee{},
	}
}

func (exec *sellExecutor) IsProfitable() bool {
	// We can look at the trade history and get the last trade
	// for symbol if action == Sell last trade == symbol.lastBuy
	// if does not exist then assume asset must have been transfered in
	if !exec.config.Sell.MustProfit || exec.tradeStartPrice == 0 {
		 return true
	}
	return (exec.marketPrice - exec.tradeStartPrice) > exec.fees.Value*2
}

func sell(sell *sellExecutor) bool {
	pretradePrice := sell.tradeStartPrice
	account := user.CreateUser().GetAccount()
	sellOrder, err := account.TradeSellConfig(sell.config, sell.marketPrice)
	
	if err != nil {
		return false
	}

	summary(
		sell.config,
		sell.config.Side,
		sell.config.Symbol,
		pretradePrice,
		sell.tradeStartPrice,
		sell.marketPrice,
		sell.marketPrice-pretradePrice,
		sell.fees,
		sell.config.Sell.Quantity,
		*sellOrder,
	)
	return true
}

func (exec *sellExecutor) Execute() bool {
	sold := sell(exec)
	return sold
}
