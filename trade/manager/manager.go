package manager

import (
	// "time"
	"trading/names"
	"trading/trade/executor"

	"trading/trade/locker"
)

type TradeManager struct {
	trader names.Trader
}

func NewTradeManager(trader names.Trader) *TradeManager {
	return &TradeManager{
		trader: trader,
	}
}

func (tm *TradeManager) DoTrade() *TradeManager {
	tradeLocker := locker.NewTradeLocker()
	tm.trader.
		SetTradeLocker(tradeLocker).
		SetExecutor(tm.Execute).
		Run()
	return tm
}

func (tm *TradeManager) Execute(
	config names.TradeConfig,
	marketPrice float64,
	basePrice float64,
	done func()) {
	var sold bool

	if config.Side.IsBuy() {
		sold = executor.BuyExecutor(config, marketPrice, basePrice).Execute()
	} else {
		sold = executor.SellExecutor(config, marketPrice, basePrice).Execute()
	}
	if !sold {
		// return
	}
	done()
}
