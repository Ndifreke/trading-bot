package traders

import (
	"math"
	"sync"
	"trading/helper"
	"trading/kline"
	"trading/names"
	"trading/trade/graph"
)


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

// 	if true {
// 		//if breakout or uptrend,
// 		//lets reduce the buy stop limit so that we can always catch on the the upgrowth of the graph

// 	}
// 	return cfg
// }
