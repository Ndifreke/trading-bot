package limit

import (
	"fmt"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/stream"
)

type TradeRunner struct {
	Config    names.TradeConfig
	StreamMan stream.StreamManager
	Locker    names.TradeLockerInterface
	Executor  names.ExecutorFunc
}

type LimitTradeManager struct {
	tradeConfigs []names.TradeConfig
	executorFunc  names.ExecutorFunc
	tradeLocker   names.TradeLockerInterface
	streamManager stream.StreamManager
}

func NewLimitTradeManager(tradeConfigs ...names.TradeConfig) names.Trader {
	return &LimitTradeManager{
		tradeConfigs: tradeConfigs,
		// streamer:      nil,
		streamManager: stream.StreamManager{},
	}
}

// checks if the current market price is equal to or greater/less than a specified stop price,
// based on a given trade configuration.
func isStopPrice(t names.TradeConfig, stopPrice, currentPrice float64) bool {
	if t.Side == names.TradeSideBuy {
		//Introduce charges
		return currentPrice <= stopPrice
	} else if t.Side == names.TradeSideSell {
		return currentPrice >= stopPrice
	}
	panic(fmt.Sprintf("Unknon trade action %s", t.Side))
}

func (t *LimitTradeManager) getConfig(symbol names.Symbol) names.TradeConfig {
	for _, tc := range t.tradeConfigs {
		if tc.Symbol == symbol {
			return tc
		}
	}
	return names.TradeConfig{}
}

func (t *LimitTradeManager) getTradeSymbols() []string {
	var s []string
	for _, tc := range t.tradeConfigs {
		s = append(s, tc.Symbol.String())
	}
	return s
}

func (t *LimitTradeManager) Run() {

	t.streamManager.NewStream(t.getTradeSymbols())

	for _, tc := range t.tradeConfigs {
		if tc.Side.IsBuy() {
			BuyRun(TradeRunner{
				Config:    tc,
				StreamMan: t.streamManager,
				Locker:    t.tradeLocker,
				Executor:  t.executorFunc,
			})
		} else if tc.Side.IsSell() {
			SellRun(TradeRunner{
				Config:    tc,
				StreamMan: t.streamManager,
				Locker:    t.tradeLocker,
				Executor:  t.executorFunc,
			})
		}
	}
}

func (t *LimitTradeManager) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

func (t *LimitTradeManager) Done(confg names.TradeConfig) {
	// todo decide if to close connection
}

func (t *LimitTradeManager) SetTradeLocker(tl names.TradeLockerInterface) names.Trader {
	t.tradeLocker = tl
	return t
}

// type LimitPredicateFun = func(ticker binance.MiniTickerData, tradingConfig names.TradeConfig) bool

func isBuyStop(stopPrice, currentPrice float64) bool {
	return currentPrice <= stopPrice
}

// The BuyRun completes a trade by buying when the current price
// is lower than the last traded price for this pair. Where the stop
// price is a fixed or a lower percentile of the last traded price of the pair
func BuyRun(runner TradeRunner) {
	config, lock, executor, streamMan := runner.Config, runner.Locker, runner.Executor, runner.StreamMan
	streamer := streamMan.GetStream()
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	buyStopPrice := helper.CalculateTradeBuyFixedPrice(config.Price.Buy, pretradePrice)

	//initialise configLocker for this config
	configLocker := lock.AddLock(config, pretradePrice, buyStopPrice) //we mayy not need stop for sell

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
		if isBuyStop(buyStopPrice, state.Price) {
			//Executor will handle the Sell or Buy of this Asset
			executor(
				state.TradeConfig,
				state.Price,
				state.PretradePrice,
				func() {
					// Will interrupt other trades since they all use the
					// same socket connection. This should be determined by
					// the manager if to call done or not
					streamer.Close()
				},
			)
		}

	})

	reader := func(conn stream.StreamInterface, message stream.PriceStreamData) {
		configLocker.TryLockPrice(message.Price)
	}
	streamer.RegisterReader(config.Symbol.String(), reader)

}

func isSellStop(stopPrice, currentPrice float64) bool {
	return int(currentPrice) >= int(stopPrice)
}

func SellRun(runner TradeRunner) {
	config, lock, executor, streamMan := runner.Config, runner.Locker, runner.Executor, runner.StreamMan
	streamer := streamMan.GetStream()
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	sellStopPrice := helper.CalculateTradeFixedSellPrice(config.Price.Sell, pretradePrice)

	// //initialise configLocker for this config
	configLocker := lock.AddLock(config, pretradePrice, sellStopPrice)

	//Set callback for when trade matures for redemption
	configLocker.SetRedemptionCandidateCallback(func(locker names.LockInterface) {
		state := locker.GetLockState()

		// 	//Needless to check sell stop as locker handles this already by ensuring trade is profitable
		if isSellStop(sellStopPrice, state.Price) {

			executor(
				state.TradeConfig,
				state.Price,
				state.PretradePrice,
				func() {
					// Will interrupt other trades since they all use the
					// same socket connection. This should be determined by
					// the manager if to call done or not
					streamer.Close()
				},
			)
		}
	})

	reader := func(conn stream.StreamInterface, message stream.PriceStreamData) {
		configLocker.TryLockPrice(message.Price)

	}
	streamer.RegisterReader(config.Symbol.String(), reader)
}

func (t *LimitTradeManager) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}
