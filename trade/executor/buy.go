package executor

import (
	"fmt"
	"trading/helper"
	"trading/request"
	"trading/trade"
)

type buyTrade struct {
	latestPrice float32
	connection  request.Connection
	extra       ExecutorExtra
	config      trade.TradeConfig
	fees        helper.TradeFee
}

func BuyExecutor(
	config trade.TradeConfig,
	currentPrice float32,
	symbol string,
	connection request.Connection,
	extra ExecutorExtra,

) *buyTrade {
	return &buyTrade{
		currentPrice,
		connection,
		extra,
		config,
		helper.TradeFee{},
	}
}

func (st buyTrade) IsProfitable() bool {

	if !st.config.Price.Buy.MustProfit || st.extra.PreTradePrice == 0 {
		// The Price we will be buying is less than the price when we started this
		// trade plus the charges for this trade buy and subsequent sell.
		// Note we may want to substitute the PriceAtRun to the last price that this
		// symbol was bought for that way we can have an accurate PriceAtRun called last
		// traded Price
		return true
	}
	fmt.Println(st.latestPrice, st.extra.PreTradePrice, st.fees.Value, "WHO")
	return (st.latestPrice + st.fees.Value) < st.extra.PreTradePrice
}

func buy(st *buyTrade) bool {
	//last trade price from API TODO note it could be
	// buy or sell action on this symbol. How do you calculate the profile
	lastTradePrice := st.extra.PreTradePrice
	summary(
		st.config.Action,
		st.config.Symbol,
		lastTradePrice,
		st.extra.PreTradePrice,
		st.latestPrice,
		st.latestPrice-lastTradePrice,
		st.fees,
		st.config.Price.Buy.Quantity,
	)
	return true
}

func (st *buyTrade) Execute() bool {
	config := st.config
	st.fees = helper.GetTradeFee(st.config, st.latestPrice)

	if !st.IsProfitable() {
		// Dont buy if user wanted to make profit by force
		return false
	}
	sold := buy(st)
	if st.config.IsCyclick && sold && st.extra.TradeManager != nil {
		buyConfig := config
		buyConfig.Action = trade.TradeActionSell
		st.extra.TradeManager(buyConfig).Run()
	}
	//Finally close the connection used by Trader socket
	st.connection.Close()
	return sold
}
