package traders

import (
	"fmt"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/trade/deviation"
	"trading/trade/graph"
	"trading/trade/manager"
	"trading/utils"

	"github.com/google/uuid"
)

type autoTrader struct {
	tradeConfigs             []names.TradeConfig
	executorFunc             names.ExecutorFunc
	tradeLockManager         names.LockManagerInterface
	interval                 string //'15m'
	datapoints               int    //18
	trend                    graph.TrendType
	broadcast                *stream.Broadcaster
	refreshConfigsOnComplete bool
}

func getAutoTrader(tradeConfigs []names.TradeConfig) *autoTrader {
	trader := &autoTrader{
		tradeConfigs:             tradeConfigs,
		broadcast:                stream.NewBroadcast(uuid.New().String()),
		refreshConfigsOnComplete: true,
	}

	return trader
}

func isValidAutoSideConfig(side names.SideConfig) bool {
	return false
}

func (t *autoTrader) Run() {
	for _, tc := range t.tradeConfigs {

		if tc.Side.IsBuy() {
			if isValidAutoSideConfig(tc.Buy) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Buy)), "Buy Side Configuration Error")
				continue
			}

		} else if tc.Side.IsSell() {
			if isValidAutoSideConfig(tc.Sell) {
				utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Sell)), "Sell Side Configuration Error")
				continue
			}
		}
		go t.Watch(tc)
	}
}

func (t *autoTrader) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *autoTrader) RemoveConfig(config names.TradeConfig) bool {
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

// Add a new config to start watching. If this config exist already
// it will be replaced by the added config and the channel and lock assocated with
// them will also be removed
func (t *autoTrader) AddConfig(config names.TradeConfig) {
	t.tradeConfigs = append(t.tradeConfigs, config)
	go t.Watch(config)
}

func (tm *autoTrader) UstradeTrend(trend graph.TrendType) *autoTrader {
	tm.trend = trend
	return tm
}

func (tm *autoTrader) Done(config names.TradeConfig, locker names.LockInterface) {

	if config.IsCyclick {
		if tm.refreshConfigsOnComplete {
			for i, c := range tm.tradeConfigs {
				if c == config {
					//switch sides inline
					tm.tradeConfigs[i].Side = helper.SwitchTradeSide(config.Side)
				}
			}
			tm.broadcast.TerminateBroadCast()

			// destroy other configs so they can get new price
			// recreate the completed config with sides switched
			go NewAutoTrade(tm.tradeConfigs, tm.datapoints, tm.interval).DoTrade()
			return
		}

		if tm.RemoveConfig(config) {
			// Keeps other configs as they are and only
			// recreate the completed config with sides switched
			fmt.Println("KEEP OTHER CONFIGS")
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
func (tm *autoTrader) shouldKeepAlive() bool {
	return len(tm.tradeConfigs) != 0
}

func (trader *autoTrader) SetLockManager(tl names.LockManagerInterface) names.Trader {
	trader.tradeLockManager = tl
	return trader
}

func (trader *autoTrader) Watch(config names.TradeConfig) {
	executor := trader.executorFunc
	lockManager := trader.tradeLockManager

	subscription := trader.broadcast.Subscribe(config)
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lockManager.AddLock(config, pretradePrice) //we mayy not need stop for sell

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

func NewAutoTrade(configs []names.TradeConfig, datapoints int, interval string) *manager.TradeManager {
	preparedConfig := alignStopWithGraph(configs, interval, datapoints)
	autoTrader := getAutoTrader(preparedConfig)
	autoTrader.datapoints = datapoints
	autoTrader.interval = interval
	return manager.NewTradeManager(autoTrader)
}
