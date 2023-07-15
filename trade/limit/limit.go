package limit

import (
	"fmt"
	"trading/binance"
	"trading/names"
	"trading/stream"
	"trading/trade/manager"
	"trading/utils"

	"github.com/davecgh/go-spew/spew"
)

type limitTrader struct {
	tradeConfigs  []names.TradeConfig
	executorFunc  names.ExecutorFunc
	tradeLocker   names.LockManagerInterface
	streamManager stream.StreamManager
}

func getLimitTrader(tradeConfigs []names.TradeConfig) names.Trader {
	return &limitTrader{
		tradeConfigs:  tradeConfigs,
		streamManager: stream.StreamManager{},
	}
}

func (t *limitTrader) getTradeSymbols() []string {
	var s []string
	for _, tc := range t.tradeConfigs {
		s = append(s, tc.Symbol.String())
	}
	return s
}

func isInvalidSide(side names.SideConfig) bool {
	return side.RateType == "" || side.Quantity == 0 || side.RateLimit == 0 || side.RateType == ""
}

func (t *limitTrader) Run() {
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
		go t.Watch(tc)
	}
}

func (t *limitTrader) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

func (tm *limitTrader) Done(config names.TradeConfig) {
	sideBeforeSwap := config.Side
	if config.IsCyclick {
		if config.Side.IsSell() {
			// reverse the config
			config.Side = names.TradeSideBuy
		} else {
			config.Side = names.TradeSideSell
		}

		for i, c := range tm.tradeConfigs {
			if c.Symbol == config.Symbol && c.Side == sideBeforeSwap {
				tm.tradeConfigs[i] = config
				break
			}
		}
		// tm.configs.replace(config)
		//Reinitialise this trade again change sides if the
		// trade is cyclic
		NewLimitTrade(tm.tradeConfigs).
			// SetGraphParam(tm.interval, tm.dataPointCount).
			DoTrade()
	}
}

func (t *limitTrader) SetLockManager(tl names.LockManagerInterface) names.Trader {
	t.tradeLocker = tl
	return t
}

func (t *limitTrader) Watch(config names.TradeConfig) {
	executor := t.executorFunc
	lock := t.tradeLocker

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
				t.Done(state.TradeConfig)
			},
		)
	})

	for sub := range subscription {
		configLocker.TryLockPrice(sub.Price)
	}
}

func (t *limitTrader) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}

func NewLimitTrade(configs []names.TradeConfig) *manager.TradeManager {
	// graphing will be done here
	limitTrade := getLimitTrader(configs)
	return manager.NewTradeManager(limitTrade)
}
