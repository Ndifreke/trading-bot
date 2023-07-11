package executor

import (
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/user"
	"trading/utils"
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

func sell(st *sellExecutor) bool {
	lastTradePrice := st.tradeStartPrice
	quoteBalance := user.CreateUser().GetAccount().GetBalance( st.config.Symbol.ParseTradingPair().Base)
	
	quantity := st.config.Sell.Quantity
	if quantity < 0 {
		quantity = quoteBalance.Locked
	}

	sellOrder, err := binance.CreateSellMarketOrder(
		st.config.Symbol,
		quantity,
	)
	if err != nil {
		utils.LogError(err,"Error Selling")
		return false
	}
	summary(
		st.config.Side,
		st.config.Symbol,
		lastTradePrice,
		st.tradeStartPrice,
		st.marketPrice,
		st.marketPrice-lastTradePrice,
		st.fees,
		st.config.Sell.Quantity,
		*sellOrder,
	)
	return true
}

func (exec *sellExecutor) Execute() bool {

	// exec.fees = helper.GetTradeFee(exec.config, exec.marketPrice)
	// if !exec.IsProfitable() {
	// 	// Dont Sell if user wanted to make profit by force
	// 	return false
	// }
	return sell(exec)
	
}
