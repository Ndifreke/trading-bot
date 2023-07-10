package manager

import (
	// "fmt"
	"math"
	"sync"
	// "time"
	"trading/helper"
	"trading/kline"
	"trading/names"
	"trading/trade/executor"
	"trading/trade/graph"
	"trading/trade/limit"
	"trading/trade/locker"

	// "github.com/davecgh/go-spew/spew"
)

type TradeManager struct {
	interval       string //'15m'
	dataPointCount int    //18
	trend          graph.TrendType
	configs        []names.TradeConfig
}

func NewTradeManager(configs ...names.TradeConfig) *TradeManager {
	return &TradeManager{
		configs: configs,
	}
}

func (tm *TradeManager) SetGraphParam(interval string, dataPointCount int) *TradeManager {
	tm.interval = interval
	tm.dataPointCount = dataPointCount
	return tm
}

func (tm *TradeManager) UstradeTrend(trend graph.TrendType) *TradeManager {
	tm.trend = trend
	return tm
}

func (tm *TradeManager) DoTrade() *TradeManager {
	tradeLocker := locker.NewTradeLocker()
	
	switch tm.trend {
	case graph.Limit:
		limit.NewLimitTradeManager(tm.configs...).
			SetTradeLocker(tradeLocker).
			SetExecutor(tm.Execute).
			Run()
	case graph.AutoTrend:
	default:
		configs := tm.tradeConfigsAutoPrepare()
		limit.NewLimitTradeManager(configs...).
			SetTradeLocker(tradeLocker).
			SetExecutor(tm.Execute).
			Run()
		// _ = spew.Dump
		// spew.Dump(configs)
	}
	return tm
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

func (tm *TradeManager) tradeConfigsAutoPrepare() []names.TradeConfig {
	wg := sync.WaitGroup{}
	var mutex sync.Mutex
	var preparedConfig []names.TradeConfig

	wg.Add(len(tm.configs))

	for _, cfg := range tm.configs {
		go (func(cg names.TradeConfig) {
			kLine := kline.NewKline(cg.Symbol.String(), tm.interval, tm.dataPointCount)
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

func (tm *TradeManager) Execute(
	config names.TradeConfig,
	marketPrice float64,
	basePrice float64,
	done func()) {
	var sold bool

	if config.Side.IsBuy() {
		sold = executor.BuyExecutor(config, marketPrice, basePrice).Execute()
	} else {
		sold = executor.SellExecutor(config, marketPrice, basePrice).Execute()
	}

	if !sold {
		return
	}

	done()
	sideBeforeSwap := config.Side
	if config.IsCyclick {
		if config.Side.IsSell() { 
			// reverse the config
			config.Side = names.TradeSideBuy
		} else {
			config.Side = names.TradeSideSell
		}

		for i, c := range tm.configs {
			if c.Symbol == config.Symbol && c.Side == sideBeforeSwap {
				tm.configs[i] = config
				break;
			}
		}
		// tm.configs.replace(config)
		//Reinitialise this trade again change sides if the
		// trade is cyclic
		NewTradeManager(tm.configs...).
			SetGraphParam(tm.interval, tm.dataPointCount).
			UstradeTrend(tm.trend).
			DoTrade()
	}

}
