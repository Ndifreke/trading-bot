package traders

import (
	"fmt"
	"sync"
	"trading/binance"
	"trading/names"
	"trading/stream"
	"trading/trade/deviation"
	"trading/trade/graph"
	"trading/trade/manager"
	"trading/utils"

	"github.com/google/uuid"
)

// Package parallelbuy provides a mechanism to execute a list of buys concurrently (in parallel),
// where the execution of each buy occurs simultaneously. Once one of the buys is successfully completed and sold,
// the remaining buys are temporarily cancelled until the successful buy is finished.
// After the successful buy is completed and sold, the program restarts the entire process with the original list of buys,
// including the ones that were previously cancelled.

// ExecuteParallelBuys executes a list of buys in parallel.
// If one of the buys is successfully completed and sold, the program restarts the process with the entire list of buys.
// All parallel buys are initiated simultaneously, and the execution flow continues without waiting for individual buy completion.
// The function returns when all buys in the list have been processed.

type autoStableBestSide struct {
	initConfigs        names.TradeConfigs
	initParams         StableTradeParam
	contentionConfigs  []names.TradeConfig
	executorFunc       names.ExecutorFunc
	tradeLockManager   names.LockManagerInterface
	trend              graph.TrendType
	bestSide           names.TradeSide
	status             status
	contentionChannels []stream.Subscription
	lock               sync.RWMutex
	fullfilConfig      names.TradeConfig
	broadcast          *stream.Broadcaster
}

func getAutoStableSideTrader(initConfigs names.TradeConfigs, initParams StableTradeParam, contentionConfigs []names.TradeConfig, bestSide names.TradeSide, status status, fullfilConfig names.TradeConfig) *autoStableBestSide {
	trader := &autoStableBestSide{
		initParams:         initParams,
		initConfigs:        initConfigs,
		contentionConfigs:  contentionConfigs,
		broadcast:          stream.NewBroadcast(uuid.New().String()),
		bestSide:           bestSide,
		status:             status,
		contentionChannels: []stream.Subscription{},
		lock:               sync.RWMutex{},
		fullfilConfig:      fullfilConfig,
	}
	return trader
}

func (t *autoStableBestSide) Run() {

	if t.status == StatusFullfilment {
		go t.Watch(t.fullfilConfig)
		return
	}
	for _, tc := range t.contentionConfigs {
		if isValidPeggedBestSideConfig(tc) {
			continue
		}
		go t.Watch(tc)
	}
}

func (t *autoStableBestSide) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *autoStableBestSide) RemoveConfig(config names.TradeConfig) bool {
	var removed bool

	if t.status == StatusFullfilment {
		removed := t.broadcast.Unsubscribe(config)
		if removed {
			return removed
		}
	}
	updatedConfigs := []names.TradeConfig{}
	for _, tc := range t.contentionConfigs {
		if tc == config {
			removed = t.broadcast.Unsubscribe(config)
		} else {
			updatedConfigs = append(updatedConfigs, tc)
		}
	}
	t.contentionConfigs = updatedConfigs
	return removed
}

// Add a new config to start watching. If this config exist already
// it will be replaced by the added config and the channel and lock assocated with
// them will also be removed
func (t *autoStableBestSide) AddConfig(config names.TradeConfig) {
	if t.status == StatusFullfilment {
		t.fullfilConfig = config
	} else {
		t.contentionConfigs = append(t.contentionConfigs, config)
	}
	go t.Watch(config)
}

func (tm *autoStableBestSide) UstradeTrend(trend graph.TrendType) *autoStableBestSide {
	tm.trend = trend
	return tm
}

// TODO Rename to small letter done and remove from interface
func (tm *autoStableBestSide) Done(tradedConfig names.TradeConfig, locker names.LockInterface) {
	tm.broadcast.TerminateBroadCast()
	nextStatus := changeStatus(tm.status)

	if nextStatus == StatusFullfilment {
		// lets fullfil this best configuration that was traded out of
		// the other list of configurations
		// tts important to retrieve the config from init config before it is redeamed
		// this is to ensure that when we prepare best config, the stopLimits, delta, deviation, etc
		// are calculated from the initial config and not an already decorated config which will
		// not be the expected outcome

		fullfilConfig, exist := tm.initConfigs.Find(tradedConfig.Id)

		if !exist {
			panic("an error occured, could not find tradedConfig config, ensure the config with id was called by <names.NewIdTradeConfigs>")
		}

		fullfilConfig.Side = tm.bestSide
		trader := createAutoStable_v1(
			tm.initParams,
			// tm.tradeConfigs,
			tm.bestSide,
			nextStatus,
			fullfilConfig,
		)
		trader.DoTrade()
		return
	} else {
		//Run operation to choose the best side
		trader := createAutoStable_v1(
			tm.initParams,
			// tm.initConfigs,
			// tm.tradeConfigs,
			tm.bestSide,
			nextStatus,
			names.TradeConfig{},
		)
		trader.DoTrade()
	}
}

func (t *autoStableBestSide) SetLockManager(lockMan names.LockManagerInterface) names.Trader {
	t.tradeLockManager = lockMan
	return t
}

func (trader *autoStableBestSide) Watch(config names.TradeConfig) {
	executor := trader.executorFunc
	subscription := trader.broadcast.Subscribe(config)

	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := trader.tradeLockManager.AddLock(config, pretradePrice)

	trader.lock.Lock()
	trader.contentionChannels = append(trader.contentionChannels, subscription)
	trader.lock.Unlock()

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

	deviation := deviation.NewDeviationManager(trader, configLocker)

	deviation.PostAddConfig(func(config names.TradeConfig) names.TradeConfig {
		// search this config from the initConfig. Note the initConfig is a blueprint
		// from which other stableTradeConfig can be created from. It represent the users
		// intent in stable asset and not the percentage or fixed value that can be used by a
		// trade locker
		cfg, exist := trader.initConfigs.Find(config.Id)
		if !exist {
			utils.LogError(fmt.Errorf("attempt to post insert config %s %s failed", config.Symbol, config.Id), "<autostablebestside_v1>")
			return config
		}
		stableConfig := getStableTradeConfigs(names.NewIdTradeConfigs(cfg))
		return stableConfig[0]
	})


	for sub := range subscription.GetChannel() {
		// if trader.status == StatusContention {

		// Deviation is executed selectively, specifically when the status is in contention.
		// This approach is adopted to prevent potential loss of gains while fulfilling.
		// TODO provide configuration to either enable or disable this behaviour
		go deviation.CheckDeviation(&subscription)
		// }
		configLocker.TryLockPrice(sub.Price)
	}
}

// bestSide the side that the contention will fall to after the parallel side finds a candidate
func createAutoStable_v1(params StableTradeParam, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *manager.TradeManager {
	var initConfigs names.TradeConfigs
	bestConfig = names.NewIdTradeConfigs(bestConfig)[0]

	if status == StatusContention {
		initConfigs = names.NewIdTradeConfigs(GenerateStableTradeConfigs(params)...)
	}

	initConfigs, fullfilConfig := configsSideToContention(initConfigs, bestSide, bestConfig, status)

	//GET update with stable limit also does what we are trying to avoid above
	contentionConfigs := getStableTradeConfigs(initConfigs)
	if status == StatusFullfilment {
		
		//Decorate the config thats needs to be fullfilled
		initConfigs = names.NewIdTradeConfigs(fullfilConfig)
		fullfilConfig = getStableTradeConfigs(initConfigs)[0]
		
	}

	autoStableBestSide := getAutoStableSideTrader(initConfigs, params, contentionConfigs, bestSide, status, fullfilConfig)

	return manager.NewTradeManager(autoStableBestSide)
}

func NewAutoStableBestSide(params StableTradeParam) *manager.TradeManager {
	bestSide, status := params.BestSide, params.Status

	if bestSide == "" {
		bestSide = names.TradeSideSell
	}
	if status == "" {
		status = StatusContention
	}
	contentionConfigs := GenerateStableTradeConfigs(params)
	if len(contentionConfigs) == 0 {
		return &manager.TradeManager{}
	}
	return createAutoStable_v1(params, bestSide, status, contentionConfigs[0])
}
