package bestside

import (
	"fmt"
	"math"
	"sync"
	"trading/binance"
	"trading/helper"
	"trading/kline"
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

type status string

var StatusContention status = "CONTENTION"
var StatusFullfilment status = "FULLFILMENT"

type bestSideTrader struct {
	tradeConfigs       []names.TradeConfig
	executorFunc       names.ExecutorFunc
	tradeLockManager   names.LockManagerInterface
	interval           string //'15m'
	datapoints         int    //18
	trend              graph.TrendType
	bestSide           names.TradeSide
	status             status
	contentionChannels []stream.Subscription
	lock               sync.RWMutex
	bestConfig         names.TradeConfig
	broadcast          *stream.Broadcaster
}

func getBestSideTrader(tradeConfigs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *bestSideTrader {
	trader := &bestSideTrader{
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

func isInvalidSide(tc names.TradeConfig) bool {
	if tc.Side.IsBuy() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Buy)), "Buy Side Configuration Error")
	}
	if tc.Side.IsSell() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Sell)), "Sell Side Configuration Error")
	}
	return false
}

func (t *bestSideTrader) Run() {
	for _, tc := range t.tradeConfigs {
		if isInvalidSide(tc) {
			continue
		}
		if t.status == StatusFullfilment {
			go t.Watch(t.bestConfig)
			break
		}
		go t.Watch(tc)
	}
}

func (t *bestSideTrader) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *bestSideTrader) RemoveConfig(config names.TradeConfig) bool {
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
func (t *bestSideTrader) AddConfig(config names.TradeConfig) {
	t.tradeConfigs = append(t.tradeConfigs, config)
	go t.Watch(config)
}

func (tm *bestSideTrader) UstradeTrend(trend graph.TrendType) *bestSideTrader {
	tm.trend = trend
	return tm
}

func changeStatus(current status) status {
	if current == StatusContention {
		return StatusFullfilment
	}
	return StatusContention
}

// TODO Rename to small letter done and remove from interface
func (tm *bestSideTrader) Done(bestConfig names.TradeConfig, locker names.LockInterface) {
	tm.broadcast.TerminateBroadCast()
	nextStatus := changeStatus(tm.status)

	if nextStatus == StatusFullfilment {
		// Lets fullfil this best configuration that was traded out of
		// the other list of configurations
		bestSideConfig := bestConfig
		bestSideConfig.Side = tm.bestSide
		NewBestSideTrade(
			tm.tradeConfigs,
			tm.datapoints,
			tm.interval,
			tm.bestSide,
			nextStatus,
			bestSideConfig,
		).DoTrade()
		return
	} else {
		//Run operation to choose the best side
		NewBestSideTrade(
			tm.tradeConfigs,
			tm.datapoints,
			tm.interval,
			tm.bestSide,
			nextStatus,
			names.TradeConfig{},
		).DoTrade()
	}
}

func (t *bestSideTrader) SetLockManager(tl names.LockManagerInterface) names.Trader {
	t.tradeLockManager = tl
	return t
}

func (trader *bestSideTrader) Watch(config names.TradeConfig) {
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

	deviationManager := deviation.NewDeviationManager(trader, configLocker)
	for sub := range subscription.GetChannel() {
		go deviationManager.CheckDeviation(&subscription)
		configLocker.TryLockPrice(sub.Price)
	}
}

func updateConfigs(configs []names.TradeConfig, bestSide names.TradeSide, bestConfig names.TradeConfig, status status) ([]names.TradeConfig, names.TradeConfig) {
	if status == StatusFullfilment {
		bestConfig.Side = bestSide
	}
	configsUpdate := []names.TradeConfig{}
	contentionSide := helper.SwitchTradeSide(bestSide)
	for _, cfg := range configs {
		cfg.Side = contentionSide
		configsUpdate = append(configsUpdate, cfg)
	}
	return configsUpdate, bestConfig
}

type AutoBestBestSideConfig struct {
	Symbol    names.Symbol
	Deviation names.SideConfigDeviation
}

func NewAutoBestSide(config []string, datapoints int, interval string, bestSide names.TradeSide, status status, bestSymbol string) *manager.TradeManager {
	var emptyConfigs = []names.TradeConfig{}
	for _, c := range config {
		emptyConfigs = append(emptyConfigs, names.TradeConfig{
			Symbol: names.Symbol(c),
			// Deviation: c.Deviation
		})
		// emptyConfigs = append(emptyConfigs, names.TradeConfig{
		// 	Symbol: c.Symbol,
		// 	// Deviation: c.Deviation
		// })
	}
	return NewBestSideTrade(emptyConfigs, datapoints, interval, bestSide, status, names.TradeConfig{Symbol: names.Symbol(bestSymbol)})
}

// bestSide the side that the contention will fall to after the parallel side finds a candidate
func NewBestSideTrade(configs []names.TradeConfig, datapoints int, interval string, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *manager.TradeManager {
	updatedConfigs, updatedBestConfig := updateConfigs(configs, bestSide, bestConfig, status)
	preparedConfig := tradeConfigsAutoPrepare(updatedConfigs, interval, datapoints)

	if status == StatusFullfilment {
		updatedBestConfig = tradeConfigsAutoPrepare([]names.TradeConfig{updatedBestConfig}, interval, datapoints)[0]
	}

	bestSideTrader := getBestSideTrader(preparedConfig, bestSide, status, updatedBestConfig)
	bestSideTrader.datapoints = datapoints
	bestSideTrader.interval = interval
	return manager.NewTradeManager(bestSideTrader)
}

func tradeConfigsAutoPrepare(configs []names.TradeConfig, interval string, datapoints int) []names.TradeConfig {
	wg := sync.WaitGroup{}
	var mutex sync.Mutex
	var preparedConfig []names.TradeConfig

	wg.Add(len(configs))

	for _, cfg := range configs {
		go (func(cg names.TradeConfig) {
			kLine := kline.NewKline(cg.Symbol.String(), interval, datapoints)
			graph := graph.NewGraph(kLine)
			config := configureFromGraph(cg, graph)
			mutex.Lock() // Lock the mutex before appending to the slice
			preparedConfig = append(preparedConfig, config)
			mutex.Unlock() // Unlock the mutex after appending
			wg.Done()

		})(cfg)
	}

	wg.Wait()
	return preparedConfig
}

func configureFromGraph(cfg names.TradeConfig, graph *graph.Graph) names.TradeConfig {

	currentPrice := graph.Kline()[len(graph.Kline())-1].Close
	midpoint := graph.GetPriceMidpoint()
	priceAvgMovement := graph.CalculateAveragePriceMovement()
	entryPoints := graph.FindAverageEntryPoints()

	sell := cfg.Sell

	// will lock profit everytime the price increases or decreases by priceAvgMovement

	sell.LockDelta = helper.GetUnitPercentageOfPrice(currentPrice, priceAvgMovement)
	//price from midpoint of the trend to the highes reported gain price by graph
	sellLimit := math.Max(entryPoints.GainHighPrice, (midpoint + priceAvgMovement))

	percentFromMidPointToHighestGain := helper.GrowthPercent(sellLimit, midpoint)
	sell.RateLimit = percentFromMidPointToHighestGain //pullpercentageOfMaxorMiN * mininmumAvagersteps IF BUll * 3 sell if Buy *2 buy
	sell.RateType = names.RatePercent
	sell.MustProfit = true

	cfg.Sell = sell
	cfg.Buy = sell

	if true {
		//if breakout or uptrend,
		//lets reduce the buy stop limit so that we can always catch on the the upgrowth of the graph

	}
	return cfg
}




// package bestside

// import (
// 	"fmt"
// 	"math"
// 	"sync"
// 	"trading/binance"
// 	"trading/helper"
// 	"trading/kline"
// 	"trading/names"
// 	"trading/stream"
// 	"trading/trade/deviation"
// 	"trading/trade/graph"
// 	"trading/trade/manager"
// 	"trading/utils"

// 	"github.com/google/uuid"
// )

// // Package parallelbuy provides a mechanism to execute a list of buys concurrently (in parallel),
// // where the execution of each buy occurs simultaneously. Once one of the buys is successfully completed and sold,
// // the remaining buys are temporarily cancelled until the successful buy is finished.
// // After the successful buy is completed and sold, the program restarts the entire process with the original list of buys,
// // including the ones that were previously cancelled.

// // ExecuteParallelBuys executes a list of buys in parallel.
// // If one of the buys is successfully completed and sold, the program restarts the process with the entire list of buys.
// // All parallel buys are initiated simultaneously, and the execution flow continues without waiting for individual buy completion.
// // The function returns when all buys in the list have been processed.

// type status string

// var StatusContention status = "CONTENTION"
// var StatusFullfilment status = "FULLFILMENT"

// type bestSideTrader struct {
// 	tradeConfigs       []names.TradeConfig
// 	executorFunc       names.ExecutorFunc
// 	tradeLockManager   names.LockManagerInterface
// 	interval           string //'15m'
// 	datapoints         int    //18
// 	trend              graph.TrendType
// 	bestSide           names.TradeSide
// 	status             status
// 	contentionChannels []stream.Subscription
// 	lock               sync.RWMutex
// 	bestConfig         names.TradeConfig
// 	broadcast          *stream.Broadcaster
// }

// func getBestSideTrader(tradeConfigs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *bestSideTrader {
// 	trader := &bestSideTrader{
// 		tradeConfigs:       tradeConfigs,
// 		broadcast:          stream.NewBroadcast(uuid.New().String()),
// 		bestSide:           bestSide,
// 		status:             status,
// 		contentionChannels: []stream.Subscription{},
// 		lock:               sync.RWMutex{},
// 		bestConfig:         bestConfig,
// 	}
// 	return trader
// }

// func isInvalidSide(tc names.TradeConfig) bool {
// 	if tc.Side.IsBuy() && false {
// 		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Buy)), "Buy Side Configuration Error")
// 	}
// 	if tc.Side.IsSell() && false {
// 		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Sell)), "Sell Side Configuration Error")
// 	}
// 	return false
// }

// func (t *bestSideTrader) Run() {
// 	for _, tc := range t.tradeConfigs {
// 		if isInvalidSide(tc) {
// 			continue
// 		}
// 		if t.status == StatusFullfilment {
// 			go t.Watch(t.bestConfig)
// 			break
// 		}
// 		go t.Watch(tc)
// 	}
// }

// func (t *bestSideTrader) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
// 	t.executorFunc = executorFunc
// 	return t
// }

// // Remove a config and it associated registeredLocks (subscription and lock)
// func (t *bestSideTrader) RemoveConfig(config names.TradeConfig) bool {
// 	var removed bool
// 	updatedConfigs := []names.TradeConfig{}
// 	for _, tc := range t.tradeConfigs {
// 		if tc != config {
// 			updatedConfigs = append(updatedConfigs, tc)
// 			removed = t.broadcast.Unsubscribe(config)
// 		}
// 	}
// 	t.tradeConfigs = updatedConfigs
// 	return removed
// }

// // Add a new config to start watching. If this config exist already
// // it will be replaced by the added config and the channel and lock assocated with
// // them will also be removed
// func (t *bestSideTrader) AddConfig(config names.TradeConfig) {
// 	t.tradeConfigs = append(t.tradeConfigs, config)
// 	go t.Watch(config)
// }

// func (tm *bestSideTrader) UstradeTrend(trend graph.TrendType) *bestSideTrader {
// 	tm.trend = trend
// 	return tm
// }

// func changeStatus(current status) status {
// 	if current == StatusContention {
// 		return StatusFullfilment
// 	}
// 	return StatusContention
// }

// // TODO Rename to small letter done and remove from interface
// func (tm *bestSideTrader) Done(bestConfig names.TradeConfig, locker names.LockInterface) {
// 	tm.broadcast.TerminateBroadCast()
// 	nextStatus := changeStatus(tm.status)

// 	if nextStatus == StatusFullfilment {
// 		// Lets fullfil this best configuration that was traded out of
// 		// the other list of configurations
// 		bestSideConfig := bestConfig
// 		bestSideConfig.Side = tm.bestSide
// 		NewBestSideTrade(
// 			tm.tradeConfigs,
// 			tm.datapoints,
// 			tm.interval,
// 			tm.bestSide,
// 			nextStatus,
// 			bestSideConfig,
// 		).DoTrade()
// 		return
// 	} else {
// 		//Run operation to choose the best side
// 		NewBestSideTrade(
// 			tm.tradeConfigs,
// 			tm.datapoints,
// 			tm.interval,
// 			tm.bestSide,
// 			nextStatus,
// 			names.TradeConfig{},
// 		).DoTrade()
// 	}
// }

// func (t *bestSideTrader) SetLockManager(tl names.LockManagerInterface) names.Trader {
// 	t.tradeLockManager = tl
// 	return t
// }

// func (trader *bestSideTrader) Watch(config names.TradeConfig) {
// 	executor := trader.executorFunc
// 	subscription := trader.broadcast.Subscribe(config)

// 	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
// 	configLocker := trader.tradeLockManager.AddLock(config, pretradePrice)

// 	trader.lock.Lock()
// 	trader.contentionChannels = append(trader.contentionChannels, subscription)
// 	trader.lock.Unlock()

// 	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
// 		state := l.GetLockState()
// 		executor(
// 			state.TradeConfig,
// 			state.Price,
// 			state.PretradePrice,
// 			func() {
// 				trader.Done(state.TradeConfig, configLocker)
// 			},
// 		)
// 	})

// 	deviationManager := deviation.NewDeviationManager(config, pretradePrice, configLocker, trader)
// 	for sub := range subscription.GetChannel() {
// 		go deviationManager.CheckDeviation(&subscription)
// 		configLocker.TryLockPrice(sub.Price)
// 	}
// }

// func updateConfigs(configs []names.TradeConfig, bestSide names.TradeSide, bestConfig names.TradeConfig, status status) ([]names.TradeConfig, names.TradeConfig) {
// 	if status == StatusFullfilment {
// 		bestConfig.Side = bestSide
// 	}
// 	configsUpdate := []names.TradeConfig{}
// 	contentionSide := helper.SwitchTradeSide(bestSide)
// 	for _, cfg := range configs {
// 		cfg.Side = contentionSide
// 		configsUpdate = append(configsUpdate, cfg)
// 	}
// 	return configsUpdate, bestConfig
// }

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

// // bestSide the side that the contention will fall to after the parallel side finds a candidate
// func NewBestSideTrade(configs []names.TradeConfig, datapoints int, interval string, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *manager.TradeManager {
// 	updatedConfigs, updatedBestConfig := updateConfigs(configs, bestSide, bestConfig, status)
// 	preparedConfig := tradeConfigsAutoPrepare(updatedConfigs, interval, datapoints)

// 	if status == StatusFullfilment {
// 		updatedBestConfig = tradeConfigsAutoPrepare([]names.TradeConfig{updatedBestConfig}, interval, datapoints)[0]
// 	}

// 	bestSideTrader := getBestSideTrader(preparedConfig, bestSide, status, updatedBestConfig)
// 	bestSideTrader.datapoints = datapoints
// 	bestSideTrader.interval = interval
// 	return manager.NewTradeManager(bestSideTrader)
// }

// func tradeConfigsAutoPrepare(configs []names.TradeConfig, interval string, datapoints int) []names.TradeConfig {
// 	wg := sync.WaitGroup{}
// 	var mutex sync.Mutex
// 	var preparedConfig []names.TradeConfig

// 	wg.Add(len(configs))

// 	for _, cfg := range configs {
// 		go (func(cg names.TradeConfig) {
// 			kLine := kline.NewKline(cg.Symbol.String(), interval, datapoints)
// 			graph := graph.NewGraph(kLine)
// 			config := configureFromGraph(cg, graph)
// 			mutex.Lock() // Lock the mutex before appending to the slice
// 			preparedConfig = append(preparedConfig, config)
// 			mutex.Unlock() // Unlock the mutex after appending
// 			wg.Done()

// 		})(cfg)
// 	}

// 	wg.Wait()
// 	return preparedConfig
// }

// func configureFromGraph(cfg names.TradeConfig, graph *graph.Graph) names.TradeConfig {

// 	currentPrice := graph.Kline()[len(graph.Kline())-1].Close
// 	midpoint := graph.GetPriceMidpoint()
// 	priceAvgMovement := graph.CalculateAveragePriceMovement()
// 	entryPoints := graph.FindAverageEntryPoints()

// 	sell := cfg.Sell

// 	// will lock profit everytime the price increases or decreases by priceAvgMovement

// 	sell.LockDelta = helper.GetUnitPercentageOfPrice(currentPrice, priceAvgMovement)
// 	//price from midpoint of the trend to the highes reported gain price by graph
// 	sellLimit := math.Max(entryPoints.GainHighPrice, (midpoint + priceAvgMovement))

// 	percentFromMidPointToHighestGain := helper.GrowthPercent(sellLimit, midpoint)
// 	sell.RateLimit = percentFromMidPointToHighestGain //pullpercentageOfMaxorMiN * mininmumAvagersteps IF BUll * 3 sell if Buy *2 buy
// 	sell.RateType = names.RatePercent
// 	sell.MustProfit = true

// 	cfg.Sell = sell
// 	cfg.Buy = sell

// 	if true {
// 		//if breakout or uptrend,
// 		//lets reduce the buy stop limit so that we can always catch on the the upgrowth of the graph

// 	}
// 	return cfg
// }




