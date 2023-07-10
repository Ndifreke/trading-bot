package limit

import (
	"fmt"
	"trading/binance"
	"trading/names"
	"trading/stream"
	"trading/utils"

	"github.com/davecgh/go-spew/spew"
)

type TradeRunner struct {
	Config    names.TradeConfig
	StreamMan stream.StreamManager
	Locker    names.TradeLockerInterface
	Executor  names.ExecutorFunc
}

type LimitTradeManager struct {
	tradeConfigs  []names.TradeConfig
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

func (t *LimitTradeManager) getTradeSymbols() []string {
	var s []string
	for _, tc := range t.tradeConfigs {
		s = append(s, tc.Symbol.String())
	}
	return s
}

func isInvalidSide(side names.SideConfig) bool {
	return side.RateType == "" || side.Quantity == 0 || side.RateLimit == 0 || side.RateType == ""
}

func (t *LimitTradeManager) Run() {

	t.streamManager.NewStream(t.getTradeSymbols())

	for _, tc := range t.tradeConfigs {
		if tc.Side.IsBuy() {
			if isInvalidSide(tc.Buy) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", spew.Sdump(tc.Buy)), "Buy Side Configuration Error")
				continue
			}

		} else if tc.Side.IsSell() {
			if isInvalidSide(tc.Sell) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", spew.Sdump(tc.Sell)), "Sell Side Configuration Error")
				continue
			}
		}
		Watch(TradeRunner{
			Config:    tc,
			StreamMan: t.streamManager,
			Locker:    t.tradeLocker,
			Executor:  t.executorFunc,
		})
		// if tc.Side.IsBuy() {
		// 	BuyRun(TradeRunner{
		// 		Config:    tc,
		// 		StreamMan: t.streamManager,
		// 		Locker:    t.tradeLocker,
		// 		Executor:  t.executorFunc,
		// 	})
		// } else if tc.Side.IsSell() {
		// 	SellRun(TradeRunner{
		// 		Config:    tc,
		// 		StreamMan: t.streamManager,
		// 		Locker:    t.tradeLocker,
		// 		Executor:  t.executorFunc,
		// 	})
		// }
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

func Watch(runner TradeRunner) {

	config, lock, executor, streamMan := runner.Config, runner.Locker, runner.Executor, runner.StreamMan
	streamer := streamMan.GetStream()
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lock.AddLock(config, pretradePrice) //we mayy not need stop for sell

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
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
	})

	reader := func(conn stream.StreamInterface, message stream.PriceStreamData) {
		configLocker.TryLockPrice(message.Price)
	}
	streamer.RegisterReader(config.Symbol.String(), reader)

}

// The BuyRun completes a trade by buying when the current price
// is lower than the last traded price for this pair. Where the stop
// price is a fixed or a lower percentile of the last traded price of the pair
func BuyRun(runner TradeRunner) {

	config, lock, executor, streamMan := runner.Config, runner.Locker, runner.Executor, runner.StreamMan
	streamer := streamMan.GetStream()
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lock.AddLock(config, pretradePrice) //we mayy not need stop for sell

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
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
	})

	reader := func(conn stream.StreamInterface, message stream.PriceStreamData) {
		configLocker.TryLockPrice(message.Price)
	}
	streamer.RegisterReader(config.Symbol.String(), reader)

}

func SellRun(runner TradeRunner) {

	config, lock, executor, streamMan := runner.Config, runner.Locker, runner.Executor, runner.StreamMan
	streamer := streamMan.GetStream()
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lock.AddLock(config, pretradePrice)

	//Set callback for when trade matures for redemption
	configLocker.SetRedemptionCandidateCallback(func(locker names.LockInterface) {
		state := locker.GetLockState()
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
	})

	streamReader := func(conn stream.StreamInterface, message stream.PriceStreamData) {
		configLocker.TryLockPrice(message.Price)
	}
	streamer.RegisterReader(config.Symbol.String(), streamReader)
}

func (t *LimitTradeManager) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}
