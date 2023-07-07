package executor

import (
	"trading/binance"
	"trading/helper"
	"trading/names"
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
	if !exec.config.Price.Sell.MustProfit || exec.tradeStartPrice == 0 {
		return true
	}
	return (exec.marketPrice - exec.tradeStartPrice) > exec.fees.Value*2
}

func sell(st *sellExecutor) bool {
	lastTradePrice := st.tradeStartPrice
	sellOrder, err := binance.CreateSellMarketOrder(
		st.config.Symbol,
		st.marketPrice,
		st.config.Price.Sell.Quantity,
	)
	if err != nil {
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
		st.config.Price.Sell.Quantity,
		*sellOrder,
	)
	return true
}

func (exec *sellExecutor) Execute() bool {
	config := exec.config
	exec.fees = helper.GetTradeFee(exec.config, exec.marketPrice)
	if !exec.IsProfitable() {
		// Dont Sell if user wanted to make profit by force
		return false
	}
	sold := sell(exec)
	if !sold {
		return sold
	}

	//Finally close the connection used by Trader socket
	// st.extra.Connection.CloseLog(fmt.Sprintf("%s Sell trade completed", config.Symbol))

	if exec.config.IsCyclick && sold {
		buyConfig := config
		buyConfig.Side = names.TradeSideBuy
	}

	return sold
}
