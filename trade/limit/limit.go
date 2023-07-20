package limit

import (
	"fmt"
	"github.com/google/uuid"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/trade/deviation"
	"trading/trade/manager"
	"trading/utils"
)

type limitTrader struct {
	tradeConfigs     []names.TradeConfig
	executorFunc     names.ExecutorFunc
	tradeLockManager names.LockManagerInterface
	broadcast        *stream.Broadcaster
	// if true, it will completely restart the trader using
	// the configuration for the configs that was not completed
	// and the completed trade configuration as it is updated after
	// Done

	// recreate on complete is important in the sense
	// that it allows all other comfiguration to get a new
	// price at the current rate. It is similar to what bestside trader does
	// but this time it does not have a priority side other than refreshing the configs
	refreshConfigsOnComplete bool
}

func getLimitTrader(tradeConfigs []names.TradeConfig) names.Trader {

	trader := &limitTrader{
		tradeConfigs: tradeConfigs,
		// registeredLocks: map[names.TradeConfig]registeredLocksType{},
		//TODO INJECT in NewLimitTrade WHEN CALLING THIS
		refreshConfigsOnComplete: true,
		broadcast:                stream.NewBroadcast(uuid.New().String()),
	}
	return trader
}

func isInvalidSide(side names.SideConfig) bool {
	return side.RateType == "" || side.Quantity == 0 || side.RateLimit == 0
}

func (t *limitTrader) Run() {
	for _, tc := range t.tradeConfigs {

		if tc.Side.IsBuy() {
			if isInvalidSide(tc.Buy) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc)), "Buy Side Configuration Error")
				continue
			}

		} else if tc.Side.IsSell() {
			if isInvalidSide(tc.Sell) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc)), "Sell Side Configuration Error")
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

// Add a new config to start watching. If this config exist already
// it will be replaced by the added config and the channel and lock assocated with
// them will also be removed
func (t *limitTrader) AddConfig(config names.TradeConfig) {
	t.tradeConfigs = append(t.tradeConfigs, config)
	go t.Watch(config)
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *limitTrader) RemoveConfig(config names.TradeConfig) bool {
	var removed bool
	updatedConfigs := []names.TradeConfig{}
	for _, tc := range t.tradeConfigs {
		if tc == config {
			removed = t.broadcast.Unsubscribe(config)
		} else {
			updatedConfigs = append(updatedConfigs, tc)
		}
	}
	t.tradeConfigs = updatedConfigs
	return removed
}

func (tm *limitTrader) Done(config names.TradeConfig, locker names.LockInterface) {

	if config.IsCyclick {
		if tm.refreshConfigsOnComplete {
			for i, c := range tm.tradeConfigs {
				if c == config {
					//switch sides inline
					tm.tradeConfigs[i].Side = helper.SwitchTradeSide(config.Side)
				}
			}

			// destroy other configs so they can get new price
			// recreate the completed config with sides switched
			tm.broadcast.TerminateBroadCast()
			go NewLimitTrade(tm.tradeConfigs).DoTrade()
			return
		}

		if tm.RemoveConfig(config) {
			// Keeps other configs as they are and only
			// recreate the completed config with sides switched
			updateConfig := config
			updateConfig.Side = helper.SwitchTradeSide(config.Side)
			tm.AddConfig(updateConfig)
		}
		return
	}
	tm.RemoveConfig(config)
	if !tm.shouldKeepAlive() {
		tm.broadcast.TerminateBroadCast()
	}
}

// should keep alive suggest if this trade broadcasr
// manager should be terminated or not base on if there
// is trade currently running if there is a configuration
// that is cyclic, should keep alive should never terminate
// this is a cuncurrency challenge.
// should keep alive is only allowed to terminate when there
// is no configuration that is cyclick and every other running
// configuration has been completed
func (tm *limitTrader) shouldKeepAlive() bool {
	return len(tm.tradeConfigs) != 0
}

func (trader *limitTrader) SetLockManager(tl names.LockManagerInterface) names.Trader {
	trader.tradeLockManager = tl
	return trader
}

func (trader *limitTrader) Watch(config names.TradeConfig) {
	executor := trader.executorFunc
	subscription := trader.broadcast.Subscribe(config)
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := trader.tradeLockManager.AddLock(config, pretradePrice) //we mayy not need stop for sell

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
		executor(
			state.TradeConfig,
			state.Price,
			state.PretradePrice,
			func() {
				trader.Done(state.TradeConfig, configLocker)
			},
		)
	})

	deviationManager := deviation.NewDeviationManager(trader, configLocker)

	for sub := range subscription.GetChannel() {
		go deviationManager.CheckDeviation(&subscription)
		configLocker.TryLockPrice(sub.Price)
	}
}

func NewLimitTrade(configs []names.TradeConfig) *manager.TradeManager {
	limitTrade := getLimitTrader(configs)
	return manager.NewTradeManager(limitTrade)
}
