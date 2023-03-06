package executor

import (
	"trading/request"
	"trading/trade"
)

type sellTrade struct {
	currentPrice float32
	connection   request.Connection
	extra        ExecutorExtra
	config       trade.TradeConfig
}

func SellExecutor(
	config trade.TradeConfig,
	currentPrice float32,
	connection request.Connection,
	extra ExecutorExtra,
) sellTrade {
	return sellTrade{
		currentPrice,
		connection,
		extra,
		config,
	}
}

func (st sellTrade) IsProfitable() bool {
	return false
}

func sell() bool {
	return true
}

func (st sellTrade) Sell() bool {
	config := st.config
	if config.Price.Sell.MustProfit && !st.IsProfitable() {
		// Dont Sell if user wanted to make profit by force
		return false
	}
	sold := sell()
	if st.config.IsCyclick && sold && st.extra.Trader != nil{
		buyConfig := config
		buyConfig.Action = trade.TradeActionBuy
		st.extra.Trader(buyConfig).Run()
	}
	//Finally close the connection used by Trader socket
	st.connection.Close()
	return sold
}
