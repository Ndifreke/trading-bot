package traders

import (
	"trading/names"
	"trading/trade/manager"
)

// bestSide the side that the contention will fall to after the parallel side finds a candidate
func createAutoStable(configs []names.TradeConfig, bestSide names.TradeSide, status status, bestConfig names.TradeConfig) *manager.TradeManager {
	config := configs[0]
	if status == StatusContention {
		//Lets look for a fresh new pair of configs to trade
		// but the problem is that we dont have context to the original
		//params, this params will keep getting smaller as they are
		// converted to percentage every time. Solution is to keep track of 
		// the original params and reuse it
		params := StableTradeParam{
			QuoteAsset:         config.Symbol.ParseTradingPair().Quote,
			SellDeviationDelta: config.Sell.DeviationSync.Delta,
			SellStopLimit:      config.Sell.StopLimit,
			SellLockDelta:      config.Sell.LockDelta,
			BuyDeviationDelta:  config.Buy.DeviationSync.Delta,
			BuyStopLimit:       config.Buy.StopLimit,
			BuyLockDelta:       config.Buy.LockDelta,
		}
		configs = GenerateStableTradeConfigs(params)
	}
	return createStableTrader(configs, bestSide, status, NewStableBestSide, bestConfig)
}

func NewAutoStableBestSide_v1(params StableTradeParam) *manager.TradeManager {
	bestSide, status := params.BestSide, params.Status

	if bestSide == ""{
		bestSide = names.TradeSideSell
	}

	if status == ""{
		status = StatusContention
	}

	contentionConfig := GenerateStableTradeConfigs(params)
	return createStableTrader(contentionConfig, bestSide, status, createAutoStable, contentionConfig[0])
}
