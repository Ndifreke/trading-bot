package executor

import (
	"trading/request"
	"trading/trade"
)

type buyTrade struct {
	currentPrice float32
	connection   request.Connection
	extra        ExecutorExtra
	config       trade.TradeConfig
}

func BuyExecutor(
	config trade.TradeConfig,
	currentPrice float32,
	connection request.Connection,
	extra ExecutorExtra,
) buyTrade {
	return buyTrade{
		currentPrice,
		connection,
		extra,
		config,
	}
}

func (st buyTrade) IsProfitable() bool {
	return false
}

func buy() bool {
	return true
}

func (st buyTrade) Buy() bool {
	config := st.config
	if config.Price.Sell.MustProfit && !st.IsProfitable() {
		// Dont Sell if user wanted to make profit by force
		return false
	}
	sold := sell()
	if st.config.IsCyclick && sold && st.extra.Trader != nil{
		buyConfig := config
		buyConfig.Action = trade.TradeActionSell
		st.extra.Trader(buyConfig).Run()
	}
	//Finally close the connection used by Trader socket
	st.connection.Close()
	return sold
}
