package executor

import (
	"fmt"
	"trading/helper"
	"trading/request"
	"trading/trade"
)

type sellTrade struct {
	latestPrice float32
	connection  request.Connection
	extra       ExecutorExtra
	config      trade.TradeConfig
	fees        helper.TradeFee
}

func SellExecutor(
	config trade.TradeConfig,
	currentPrice float32,
	connection request.Connection,
	extra ExecutorExtra,
) *sellTrade {

	return &sellTrade{
		currentPrice,
		connection,
		extra,
		config,
		helper.TradeFee{},
	}
}

func (st *sellTrade) IsProfitable() bool {
	// We can look at the trade history and get the last trade
	// for symbol if action == Sell last trade == symbol.lastBuy
	// if does not exist then assume asset must have been transfered in
	if !st.config.Price.Sell.MustProfit || st.extra.PreTradePrice == 0 {
		return true
	}
	return (st.latestPrice - st.extra.PreTradePrice) > st.fees.Value*2
}

func sell(st *sellTrade) bool {
	lastTradePrice := st.extra.PreTradePrice
	summary(
		st.config.Side,
		st.config.Symbol,
		lastTradePrice,
		st.extra.PreTradePrice,
		st.latestPrice,
		st.latestPrice-lastTradePrice,
		st.fees,
		st.config.Price.Sell.Quantity,
	)
	return true
}

func (st *sellTrade) Execute() bool {
	config := st.config
	st.fees = helper.GetTradeFee(st.config, st.latestPrice)
	if !st.IsProfitable() {
		// Dont Sell if user wanted to make profit by force
		return false
	}
	sold := sell(st)
	if !sold {
		return sold
	}

	//Finally close the connection used by Trader socket
	st.connection.CloseLog(fmt.Sprintf("%s Sell trade completed", config.Symbol))

	if st.config.IsCyclick && sold && st.extra.TradeManager != nil {
		buyConfig := config
		buyConfig.Side = trade.TradeSideBuy
		st.extra.TradeManager(buyConfig).Run()
	}

	return sold
}
