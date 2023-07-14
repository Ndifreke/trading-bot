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
	"trading/trade/graph"
	"trading/trade/manager"
	"trading/utils"

	"github.com/davecgh/go-spew/spew"
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
	tradeLocker        names.TradeLockerInterface
	streamManager      stream.StreamManager
	interval           string //'15m'
	datapoints         int    //18
	trend              graph.TrendType
	bestSide           names.TradeSide
	status             status
	contentionChannels []stream.SubscriptionChan
	lock               sync.RWMutex
	bestConfig         names.TradeConfig
}

func getBestSideTrader(tradeConfigs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *bestSideTrader {
	return &bestSideTrader{
		tradeConfigs:       tradeConfigs,
		streamManager:      stream.StreamManager{},
		bestSide:           bestSide,
		status:             status,
		contentionChannels: []stream.SubscriptionChan{},
		lock:               sync.RWMutex{},
		bestConfig:         bestConfig,
	}
}

func isInvalidSide(tc names.TradeConfig) bool {
	if tc.Side.IsBuy() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", spew.Sdump(tc.Buy)), "Buy Side Configuration Error")
	}
	if tc.Side.IsSell() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", spew.Sdump(tc.Sell)), "Sell Side Configuration Error")
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

func changeSide(current names.TradeSide) names.TradeSide {
	if current == names.TradeSideBuy {
		return names.TradeSideSell
	}
	return names.TradeSideBuy
}

func closeChannels(chs []chan stream.PriceStreamData) {
	//stop all other channel
	for _, c := range chs {
		_, ok := <-c
		if ok {
			fmt.Println("OKAY CHAN")
			close(c)
		}
	}
}

// TODO Rename to small letter done and remove from interface
func (tm *bestSideTrader) Done(bestConfig names.TradeConfig) {

	nextStatus := changeStatus(tm.status)
	// go closeChannels(tm.contentionChannels)

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
	}
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

func (t *bestSideTrader) SetTradeLocker(tl names.TradeLockerInterface) names.Trader {
	t.tradeLocker = tl
	return t
}

func (t *bestSideTrader) Watch(config names.TradeConfig) {
	executor := t.executorFunc
	lock := t.tradeLocker

	subscription := stream.Broadcaster.Subscribe(config.Symbol.String())

	t.lock.Lock()
	t.contentionChannels = append(t.contentionChannels, stream.SubscriptionChan{Id: config.Symbol.String(), Subscription: subscription})
	t.lock.Unlock()

	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := lock.AddLock(config, pretradePrice) //we mayy not need stop for sell
	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
		executor(
			state.TradeConfig,
			state.Price,
			state.PretradePrice,
			func() { 
				stream.Broadcaster.UnsubscribeList(t.contentionChannels)
				t.Done(state.TradeConfig)
			},
		)
	})

	for sub := range subscription {
		configLocker.TryLockPrice(sub.Price)
	}
}

func (t *bestSideTrader) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}

func updateConfigs(configs []names.TradeConfig, bestSide names.TradeSide, bestConfig names.TradeConfig, status status) ([]names.TradeConfig, names.TradeConfig) {
	if status == StatusFullfilment {
		bestConfig.Side = bestSide
	}
	configsUpdate := []names.TradeConfig{}
	contentionSide := changeSide(bestSide)
	for _, cfg := range configs {
		cfg.Side = contentionSide
		configsUpdate = append(configsUpdate, cfg)
	}
	return configsUpdate, bestConfig
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

	percentFromMidPointToHighestGain := helper.GetPercentGrowth(sellLimit, midpoint)
	sell.RateLimit = percentFromMidPointToHighestGain //pullpercentageOfMaxorMiN * mininmumAvagersteps IF BUll * 3 sell if Buy *2 buy
	sell.RateType = names.RatePercent
	sell.MustProfit = true

	cfg.Sell = sell

	if true {
		//if breakout or uptrend,
		//lets reduce the buy stop limit so that we can always catch on the the upgrowth of the graph

	}
	return cfg
}
