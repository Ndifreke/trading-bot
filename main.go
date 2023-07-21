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
			RateLimit:  0,
			RateType:   names.RatePercent,
			Quantity:   1,
			MustProfit: true,
		},
		Buy: names.SideConfig{
			RateLimit:  1,
			RateType:   names.RatePercent,
			Quantity:   1,
			MustProfit: true,
		},
		Symbol: "BTCUSDT",
		Side:   names.TradeSideSell,
		// IsCyclick: true,
	}
	config2 := names.TradeConfig{
		Sell: names.SideConfig{
			RateLimit: 0,
			RateType:  names.RatePercent,
			Quantity:  200,
		},
		Buy: names.SideConfig{
			RateLimit: 99,
			RateType:  names.RatePercent,
			Quantity:  1,
		},
		Symbol: "BTCBUSD",
		Side:   names.TradeSideBuy,
	}
	a, b := names.Symbol("BNBBUSD"), names.Symbol("BTCUSDT")
	unused(b, a)
	v := a
	g := graph.NewBinanceGraph(v.String(), "15m", 8)

	g.SaveToFile("")

	config3 := names.TradeConfig{
		Symbol:    "DIAUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  2,
			LockDelta:  0.4,
			Quantity:   -1,
			Deviation: names.SideConfigDeviation{
				SwitchSide: true,
				Threshold:  0.00034,
			},
		},
		Buy: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  2,
			LockDelta:  0.4,
			Quantity:   -1,
			Deviation: names.SideConfigDeviation{
				SwitchSide: true,
				Threshold:  0.00034,
			},
		},
	}

	config4 := names.TradeConfig{
		IsCyclick: true,

		Symbol: "BTCBUSD",
		Side:   names.TradeSideBuy,
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RateFixed,
			RateLimit:  50,
			LockDelta:  0.1,
			Quantity:   0.54,
		},
		Buy: names.SideConfig{
			MustProfit: true,
			RateType:   names.RateFixed,
			RateLimit:  50,
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
			RateType:   names.RatePercent,
			RateLimit:  10,
			LockDelta:  10,
			Quantity:   -3,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  90.5,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}
	// auto.NewAutoTrade([]names.TradeConfig{autoConfig, config4}, 8, "15m").DoTrade()
	bestConfigs := []names.TradeConfig{config3, config}

	//  bestside.NewBestSideTrade(bestConfigs, 12, "15m", names.TradeSideSell, bestside.StatusContention, names.TradeConfig{}).DoTrade()
	j := []names.TradeConfig{config3, config4, config2, autoConfig}
	x := []names.TradeConfig{
		// config3,
		config4,
	}
	// y := []names.TradeConfig{config4}
	// limit.NewLimitTrade(j).DoTrade()
	traders.NewLimitTrade(x).DoTrade()
	autoBestConfig := []string{"BNBUSDT", "XRPUSDT", "SOLUSDT", "TROYUSDT", "ETHUSDT", "BTCUSDT", "SOLUSDT", "AVAXUSDT"}
	traders.NewAutoBestSide(autoBestConfig, 12, "15m", names.TradeSideSell, traders.StatusContention, autoBestConfig[0]).DoTrade()
	unused(bestConfigs)
	unused(autoBestConfig)
	unused(traders.NewBestSideTrade)
	unused(j)

	unused(traders.NewAutoTrade)
	unused(autoConfig)

	wg.Wait()
}

func unused(v ...any) {
	_ = v
}
