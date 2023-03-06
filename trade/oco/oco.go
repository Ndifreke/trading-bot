package oco

import (
	"fmt"
	"trading/binance"
	"trading/request"
	"trading/trade"
	"trading/trade/limit"
	"trading/utils"
)

type Oco struct {
	config trade.TradeConfig
	conn   request.Connection
}

func NewOco(config trade.TradeConfig) *Oco {
	return &Oco{
		config: config,
	}
}

func (oco Oco) Run() {
	preTradePrice := trade.GetSymbolPrices(oco.Symbol)
	utils.LogInfo(fmt.Sprintf("Pre Trade prices %v", preTradePrice))

	predicateFun := func() limit.LimitPredicateFun {
		return func(ticker binance.MiniTickerData, t trade.TradeConfig) bool {
			stopPrice := trade.CalculateTradePrice(t, preTradePrice[ticker.Data.Symbol])
			return true
		}
	}

	limit.NewLimitTrade(trade.TradeConfig{
		Price: struct {
			Sell trade.Price
			Buy  trade.Price
		}{
			Sell: trade.Price{
				Value:    5,
				Type:     trade.PriceTypePercent,
				Quantity: 1,
			},
		},
		Symbol: []string{"BNBUSDT"},
		Action: "SELL",
	}).RunB(predicateFun())

}

func (oco Oco) CancelTrade() {
	oco.conn.Close()
}
