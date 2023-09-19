package traders

import (
	"fmt"
	"sync"
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

// Package parallelbuy provides a mechanism to execute a list of buys concurrently (in parallel),
// where the execution of each buy occurs simultaneously. Once one of the buys is successfully completed and sold,
// the remaining buys are temporarily cancelled until the successful buy is finished.
// After the successful buy is completed and sold, the program restarts the entire process with the original list of buys,
// including the ones that were previously cancelled.

// ExecuteParallelBuys executes a list of buys in parallel.
// If one of the buys is successfully completed and sold, the program restarts the process with the entire list of buys.
// All parallel buys are initiated simultaneously, and the execution flow continues without waiting for individual buy completion.
// The function returns when all buys in the list have been processed.

type peggedBestSide struct {
	tradeConfigs       []names.TradeConfig
	executorFunc       names.ExecutorFunc
	tradeLockManager   names.LockManagerInterface
	trend              graph.TrendType
	bestSide           names.TradeSide
	status             status
	contentionChannels []stream.Subscription
	lock               sync.RWMutex
	bestConfig         names.TradeConfig
	broadcast          *stream.Broadcaster
}

func getPeggedSideTrader(tradeConfigs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *peggedBestSide {
	trader := &peggedBestSide{
		tradeConfigs:       tradeConfigs,
		broadcast:          stream.NewBroadcast(uuid.New().String()),
		bestSide:           bestSide,
		status:             status,
		contentionChannels: []stream.Subscription{},
		lock:               sync.RWMutex{},
		bestConfig:         bestConfig,
	}
	return trader
}

func isValidPeggedBestSideConfig(tc names.TradeConfig) bool {
	if tc.Side.IsBuy() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Buy)), "Buy Side Configuration Error")
	}
	if tc.Side.IsSell() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Sell)), "Sell Side Configuration Error")
	}
	return false
}

func (t *peggedBestSide) Run() {
	for _, tc := range t.tradeConfigs {
		if isValidPeggedBestSideConfig(tc) {
			continue
		}
		if t.status == StatusFullfilment {
			go t.Watch(t.bestConfig)
			break
		}
		go t.Watch(tc)
	}
}

func (t *peggedBestSide) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *peggedBestSide) RemoveConfig(config names.TradeConfig) bool {
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
func (t *peggedBestSide) AddConfig(config names.TradeConfig) {
	t.tradeConfigs = append(t.tradeConfigs, config)
	go t.Watch(config)
}

func (tm *peggedBestSide) UstradeTrend(trend graph.TrendType) *peggedBestSide {
	tm.trend = trend
	return tm
}

// TODO Rename to small letter done and remove from interface
func (tm *peggedBestSide) Done(bestConfig names.TradeConfig, locker names.LockInterface) {
	tm.broadcast.TerminateBroadCast()
	nextStatus := changeStatus(tm.status)

	if nextStatus == StatusFullfilment {
		// Lets fullfil this best configuration that was traded out of
		// the other list of configurations
		bestSideConfig := bestConfig
		bestSideConfig.Side = tm.bestSide
		NewPeggedBestSide(
			tm.tradeConfigs,
			tm.bestSide,
			nextStatus,
			bestSideConfig,
		).DoTrade()
		return
	} else {
		//Run operation to choose the best side
		NewPeggedBestSide(
			tm.tradeConfigs,
			tm.bestSide,
			nextStatus,
			names.TradeConfig{},
		).DoTrade()
	}
}

func (t *peggedBestSide) SetLockManager(tl names.LockManagerInterface) names.Trader {
	t.tradeLockManager = tl
	return t
}

func (trader *peggedBestSide) Watch(config names.TradeConfig) {
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

	for sub := range subscription.GetChannel() {
		if trader.status == StatusContention {

			//We only want to run deviation when the status is in contention
			// to avoid loosing gains while fulliling our contention
			go deviation.CheckDeviation(&subscription)
		}
		configLocker.TryLockPrice(sub.Price)
	}
}

// type AutoBestBestSideConfig struct {
// 	Symbol    names.Symbol
// 	Deviation names.SideConfigDeviation
// }

// func NewAutoBestSide(config []string, datapoints int, interval string, bestSide names.TradeSide, status status, bestSymbol string) *manager.TradeManager {
// 	var emptyConfigs = []names.TradeConfig{}
// 	for _, c := range config {
// 		emptyConfigs = append(emptyConfigs, names.TradeConfig{
// 			Symbol: names.Symbol(c),
// 			// Deviation: c.Deviation
// 		})
// 		// emptyConfigs = append(emptyConfigs, names.TradeConfig{
// 		// 	Symbol: c.Symbol,
// 		// 	// Deviation: c.Deviation
// 		// })
// 	}
// 	return NewBestSideTrade(emptyConfigs, datapoints, interval, bestSide, status, names.TradeConfig{Symbol: names.Symbol(bestSymbol)})
// }

// bestSide the side that the contention will fall to after the parallel side finds a candidate
func NewPeggedBestSide(configs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *manager.TradeManager {
	updatedConfigs, updatedBestConfig := configsSideToContention(configs, bestSide, bestConfig, status)
	// preparedConfig := alignStopWithGraph(updatedConfigs, interval, datapoints)

	preparedConfig := getUpdateWithPeggedLimit(updatedConfigs)

	if status == StatusFullfilment {
		updatedBestConfig = getUpdateWithPeggedLimit([]names.TradeConfig{updatedBestConfig})[0]
	}

	peggedBestSide := getPeggedSideTrader(preparedConfig, bestSide, status, updatedBestConfig)

	return manager.NewTradeManager(peggedBestSide)
}
