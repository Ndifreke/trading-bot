package executor

import (
	// "fmt"
	// "time"
	// tradeBinance "trading/binance"
	"trading/helper"
	"trading/names"
	"trading/user"
	// "trading/utils"
	// binance "github.com/adshao/go-binance/v2"
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

func buy(buy *buyExecutor) bool {

	pretradePrice := buy.tradeStartPrice
	account := user.CreateUser().GetAccount()

	buyOrder, err := account.TradeBuyConfig(buy.config, buy.marketPrice)
	if err != nil {
		return false
	}
	summary(
		buy.config,
		buy.config.Side,
		buy.config.Symbol,
		pretradePrice,
		buy.tradeStartPrice,
		buy.marketPrice,
		buy.marketPrice-pretradePrice,
		buy.fees,
		buy.config.Buy.Quantity,
		*buyOrder,
	)
	return true
}

func (exec *buyExecutor) Execute() bool {
	exec.fees = helper.GetTradeFee(exec.config, exec.marketPrice)
	bought := buy(exec)
	return bought
}
