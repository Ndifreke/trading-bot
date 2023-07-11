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
		tradeConfigs:  tradeConfigs,
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

		go Watch(TradeRunner{
			Config: tc,
			Locker:   t.tradeLocker,
			Executor: t.executorFunc,
		})

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

	config, lock, executor := runner.Config, runner.Locker, runner.Executor

	subscription := stream.Broadcaster.Subscribe(config.Symbol.String())
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lock.AddLock(config, pretradePrice) //we mayy not need stop for sell

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
		executor(
			state.TradeConfig,
			state.Price,
			state.PretradePrice,
			func() {
				stream.Broadcaster.Unsubscribe(state.TradeConfig.Symbol.String(), subscription)
			},
		)
	})

	for sub := range subscription {
		configLocker.TryLockPrice(sub.Price)
	}

}

func (t *LimitTradeManager) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}
