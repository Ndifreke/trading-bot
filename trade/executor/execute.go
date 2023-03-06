package executor

import (
	"trading/trade"
)

type TradeExecutor interface {
	IsProfitable() bool
}

type ExecutorExtra struct {
	PreTradePrice float32
	Trader        func(t trade.TradeConfig) trade.TradeRunner
}
