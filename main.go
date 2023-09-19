package main

import (
	"sync"
	"trading/binance"
	"trading/names"

	// "github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"

	"trading/trade/graph"
	"trading/trade/traders"
)

func init() {
	godotenv.Load()
	binance.LoadExchangeInfo()
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
			LimitType:  names.RatePercent,
			StopLimit:  10,
			LockDelta:  10,
			Quantity:   -3,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  90.5,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}
	// auto.NewAutoTrade([]names.TradeConfig{autoConfig, config4}, 8, "15m").DoTrade()
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
			Quantity:   names.MAX_LIMIT,
		},
	}
	// traders.NewPeggedLimit([]names.TradeConfig{assetGainsConfig}).DoTrade()

	unused(traders.NewBestSideTrade)
	unused(j)

	GMXUSDT := names.TradeConfig{
		Symbol:    "CRVUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  2,
			LockDelta:  50,
			Quantity:   names.MAX_LIMIT,
			DeviationSync: names.DeviationSync{
				Delta: 91,
			},
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			StopLimit:  2,
			LockDelta:  50,
			Quantity:   names.MAX_LIMIT,
			DeviationSync: names.DeviationSync{
				Delta: 91,
			},
		},
	}

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

	peggedBestConfigs := []names.TradeConfig{ /*BNBUSDT,XRPUSDT,*/ GMXUSDT /*,AMPUSDT*/}
	traders.NewPeggedBestSide(peggedBestConfigs, names.TradeSideSell, traders.StatusFullfilment, peggedBestConfigs[0]).DoTrade()

	unused(traders.NewAutoTrade)
	unused(autoConfig)
	unused(v)
	unused(assetGainsConfig)
	unused(BNBUSDT)

	wg.Wait()
}

func unused(v ...any) {
	_ = v
}
