package auto

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

type autoTrader struct {
	tradeConfigs  []names.TradeConfig
	executorFunc  names.ExecutorFunc
	tradeLocker   names.TradeLockerInterface
	streamManager stream.StreamManager
	interval      string //'15m'
	datapoints    int    //18
	trend         graph.TrendType
}

func getAutoTrader(tradeConfigs []names.TradeConfig) *autoTrader {
	return &autoTrader{
		tradeConfigs:  tradeConfigs,
		streamManager: stream.StreamManager{},
	}
}

func (t *autoTrader) getTradeSymbols() []string {
	var s []string
	for _, tc := range t.tradeConfigs {
		s = append(s, tc.Symbol.String())
	}
	return s
}

func isInvalidSide(side names.SideConfig) bool {
	return false
}

func (t *autoTrader) Run() {
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

func (t *autoTrader) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

func (tm *autoTrader) UstradeTrend(trend graph.TrendType) *autoTrader {
	tm.trend = trend
	return tm
}

func (tm *autoTrader) Done(config names.TradeConfig) {
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
		NewAutoTrade(tm.tradeConfigs, tm.datapoints, tm.interval).
			DoTrade()
	}
}

func (t *autoTrader) SetTradeLocker(tl names.TradeLockerInterface) names.Trader {
	t.tradeLocker = tl
	return t
}

func (t *autoTrader) Watch(config names.TradeConfig) {
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

func (t *autoTrader) SetStreamManager(sm stream.StreamManager) {
	t.streamManager = sm
}

func NewAutoTrade(configs []names.TradeConfig, datapoints int, interval string) *manager.TradeManager {
	preparedConfig := tradeConfigsAutoPrepare(configs, interval, datapoints)
	autoTrader := getAutoTrader(preparedConfig)
	autoTrader.datapoints = datapoints
	autoTrader.interval = interval
	return manager.NewTradeManager(autoTrader)
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
