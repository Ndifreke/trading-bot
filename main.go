package main

import (
	"sync"
	"trading/names"

	// "github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"

	"trading/trade/graph"
	"trading/trade/traders"
)

func init() {
	godotenv.Load()
	names.LoadStoredExchangeInfo()
}

func main() {

	// Start the timer
	// start := time.Now()

	// Create a wait group to track goroutines
	var wg sync.WaitGroup

	// Process items in parallel
	// for i := 0; i < numChunks; i++ {
	// 	wg.Add(1)
	// 	go processChunk(i*itemsPerChunk, &wg)
	// }

	// Wait for all goroutines to finish
	wg.Wait()

	// bn := binance.New[binance.PriceJson](args{
	// 	Api: binance.Endpoints.PriceLatest,
	// })
	params := map[string]string{"symbol": "USDTAPT"}
	_ = params

	config := names.TradeConfig{
		IsCyclick: true,

		Sell: names.SideConfig{
			StopLimit:  0,
			LimitType:  names.RatePercent,
			Quantity:   1,
			MustProfit: true,
		},
		Buy: names.SideConfig{
			StopLimit:  1,
			LimitType:  names.RatePercent,
			Quantity:   1,
			MustProfit: true,
		},
		Symbol: "BTCUSDT",
		Side:   names.TradeSideSell,
		// IsCyclick: true,
	}
	config2 := names.TradeConfig{
		Sell: names.SideConfig{
			StopLimit: 0,
			LimitType: names.RatePercent,
			Quantity:  200,
		},
		Buy: names.SideConfig{
			StopLimit: 99,
			LimitType: names.RatePercent,
			Quantity:  1,
		},
		Symbol: "BTCBUSD",
		Side:   names.TradeSideBuy,
	}
	a, b := names.Symbol("BNBBUSD"), names.Symbol("BTCUSDT")
	unused(b, a)
	v := "LTCUSDT"
	g := graph.NewBinanceGraph(v, "15m", 8)

	g.SaveToFile("")

	config3 := names.TradeConfig{
		Symbol:    "DIAUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  1,
			LockDelta:  0.4,
			Quantity:   -1,
			DeviationSync: names.DeviationSync{
				FlipSide: true,
				Delta:    0.00034,
			},
		},
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  1,
			LockDelta:  0.4,
			Quantity:   -1,
			DeviationSync: names.DeviationSync{
				FlipSide: true,
				Delta:    0.00034,
			},
		},
	}

	config4 := names.TradeConfig{
		IsCyclick: true,

		Symbol: "BTCBUSD",
		Side:   names.TradeSideBuy,
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RateFixed,
			StopLimit:  50,
			LockDelta:  0.1,
			Quantity:   0.54,
		},
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RateFixed,
			StopLimit:  50,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}
	_ = config2
	_ = config4
	_ = config
	// j := []names.TradeConfig{config4, config3, config5 } //config

	wg.Add(1)

	autoConfig := names.TradeConfig{
		Symbol:    "BNBUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: !true,
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RateFixed,
			StopLimit:  10,
			LockDelta:  10,
			Quantity:   -3,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RateFixed,
			StopLimit:  50,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}
	// traders.NewLimitTrade([]names.TradeConfig{autoConfig, config4}).DoTrade()
	bestConfigs := []names.TradeConfig{config3, config}

	//  bestside.NewBestSideTrade(bestConfigs, 12, "15m", names.TradeSideSell, bestside.StatusContention, names.TradeConfig{}).DoTrade()
	j := []names.TradeConfig{config3, config4, config2, autoConfig}
	// x := []names.TradeConfig{
	// 	// config3,
	// 	config4,
	// }
	// y := []names.TradeConfig{config4}
	// limit.NewLimitTrade(j).DoTrade()
	// traders.NewLimitTrade(x).DoTrade()
	// autoBestConfig := []string{"BNBUSDT", "XRPUSDT", "SOLUSDT", "TROYUSDT", "ETHUSDT", "BTCUSDT", "SOLUSDT", "AVAXUSDT"}
	unused(bestConfigs)

	autoBestConfig := []string{"BNBUSDT", "ETHUSDT", "BTCUSDT", "PEPEUSDT"}
	// traders.NewAutoBestSide(autoBestConfig, 12, "15m", names.TradeSideSell, traders.StatusContention, autoBestConfig[0]).DoTrade()
	unused(autoBestConfig)

	assetGainsConfig := names.TradeConfig{
		Symbol:    "BNBUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  1,
			LockDelta:  10,
			Quantity:   -1,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  1,
			LockDelta:  20,
			Quantity:   names.MAX_QUANTITY,
		},
	}
	// traders.NewPeggedLimit([]names.TradeConfig{assetGainsConfig}).DoTrade()

	unused(traders.NewBestSideTrade)
	unused(j)

	BNBUSDT := names.TradeConfig{
		Symbol:    "BNBUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  90,
			LockDelta:  20,
			Quantity:   -1,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  9,
			LockDelta:  30,
			Quantity:   -1,
		},
	}

	// XRPUSDT := names.TradeConfig{
	// 	Symbol:    "XRPUSDT",
	// 	Side:      names.TradeSideSell,
	// 	IsCyclick: true,
	// 	Buy: names.SideConfig{
	// 		MustProfit: true,
	// 		RateType:   names.RatePercent,
	// 		RateLimit:  2,
	// 		LockDelta:  30,
	// 		Quantity:   -1,
	// 	},
	// 	Sell: names.SideConfig{
	// 		MustProfit: true,
	// 		RateType:   names.RatePercent,
	// 		RateLimit:  2,
	// 		LockDelta:  30,
	// 		Quantity:   -1,
	// 	},
	// }

	// AMPUSDT := names.TradeConfig{
	// 	Symbol:    "AMPUSDT",
	// 	Side:      names.TradeSideSell,
	// 	IsCyclick: true,
	// 	Buy: names.SideConfig{
	// 		MustProfit: true,
	// 		RateType:   names.RatePercent,
	// 		RateLimit:  2,
	// 		LockDelta:  30,
	// 		Quantity:   -1,
	// 	},
	// 	Sell: names.SideConfig{
	// 		MustProfit: true,
	// 		RateType:   names.RatePercent,
	// 		RateLimit:  2,
	// 		LockDelta:  30,
	// 		Quantity:   -1,
	// 	},
	// }

	GMXUSDT := names.TradeConfig{
		Symbol: "BTCUSDT",
		// Side:      names.TradeSideBuy,
		IsCyclick: true,
		Buy: names.SideConfig{
			MustProfit:    true,
			LimitType:     names.RatePercent,
			StopLimit:     1,
			LockDelta:     0.2,
			Quantity:      names.MAX_QUANTITY,
			DeviationSync: names.DeviationSync{
				// Delta: 1,
			},
		},
		Sell: names.SideConfig{
			MustProfit:    true,
			LimitType:     names.RatePercent,
			StopLimit:     1,
			LockDelta:     0.1,
			Quantity:      names.MAX_QUANTITY,
			DeviationSync: names.DeviationSync{
				// Delta: 1,
			},
		},
	}
	// peggedBestConfigs := []names.TradeConfig{ /*BNBUSDT,XRPUSDT,*/ GMXUSDT /*,AMPUSDT*/}
	// traders.NewStableBestSide(peggedBestConfigs, names.TradeSideBuy, traders.StatusContention, peggedBestConfigs[0]).DoTrade()
	// traders.NewStableBestSide_v1(peggedBestConfigs, names.TradeSideSell, traders.StatusContention, peggedBestConfigs[0]).DoTrade()

	// tradeParam := traders.StableTradeParam{
	// 	QuoteAsset: "USDT",
	// 	BuyStopLimit: 0.001,
	// 	BuyDeviationDelta: 0.001,
	// 	BuyLockDelta: 0.0000001,
	// 	SellStopLimit: 0.001,
	// 	SellDeviationDelta: 0.001,
	// 	SellLockDelta: 0.001,
	// 	BestSide: names.TradeSideSell,
	// 	Status: traders.StatusContention,
	// 	MinPriceChange: 14,
	// 	MaxPriceChange: 24,
	// }
	// traders.NewAutoStableBestSideExample(!true)
	// traders.NewAutoStableExample(!true)
	traders.NewAutoStableBuyHighExample(true)
	// traders.NewAutoStableSplitExample(!true)
	// traders.NewStableLimitExample(!true)

	unused(traders.NewAutoTrade)
	unused(autoConfig)
	unused(v)
	unused(assetGainsConfig)
	unused(BNBUSDT)
	unused(GMXUSDT)
	wg.Wait()
}

func unused(v ...any) {
	_ = v
}
