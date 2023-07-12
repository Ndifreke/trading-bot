package main

import (
	"io"
	"log"
	"net"

	// "net/http"
	"os"
	// "trading/helper"
	"sync"
	"trading/names"

	// "github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"

	// "trading/user"
	detection "trading/trade/graph"
	"trading/trade/manager"
)

const (
	SERVER = "localhost"
	PORT   = "8080"
)

func init() {
	godotenv.Load()
	names.GetSymbols()
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


	_ = wg
	// bn := binance.New[binance.PriceJson](args{
	// 	Api: binance.Endpoints.PriceLatest,
	// })
	params := map[string]string{"symbol": "USDTAPT"}
	_ = params
	
	config := names.TradeConfig{
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
	g := detection.NewBinanceGraph(v.String(), "15m", 8)

	g.SaveToFile("")


	config3 := names.TradeConfig{
		Symbol:    "FLMUSDT",
		Side:      names.TradeSideSell,
		IsCyclick: true,
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  2,
			LockDelta:  0.4,
			Quantity:   -1,
		},
		Buy: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  1.5,
			LockDelta:  0.4,
			Quantity:  -1,
		},
	}

	config5 := names.TradeConfig{
		Symbol:    "BNBUSDT",
		Side:      names.TradeSideBuy,
		IsCyclick: !true,
		Buy: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  10,
			LockDelta:  10,
			Quantity:   54,
		},
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  0.5,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}

	config4 := names.TradeConfig{
		Symbol:    "BTCBUSD",
		Side:      names.TradeSideBuy,
		IsCyclick: !true,
		Sell: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  0.9,
			LockDelta:  0.1,
			Quantity:   0.54,
		},
		Buy: names.SideConfig{
			MustProfit: true,
			RateType:   names.RatePercent,
			RateLimit:  0.5,
			LockDelta:  0.1,
			Quantity:   100,
		},
	}
	_ = config2
	_ = config4
	_ = config5
	_ = config
	// j := []names.TradeConfig{config4, config3, config5 } //config
	j := []names.TradeConfig{config3}
	wg.Add(1)

	// x := user.CreateUser().GetAccount().GetBalance("USDT")
	// spew.Dump(x.Free)
	manager.NewTradeManager(j...).UstradeTrend(detection.Limit).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(j...).UstradeTrend(detection.Limit).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(j...).UstradeTrend(detection.Limit).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(j...).UstradeTrend(detection.Limit).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(j...).UstradeTrend(detection.Limit).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(tradeConfigs...).SetGraphParam("15m", 18).DoTrade()
	unused(j)
	// unused(manager.NewTradeManager)

	wg.Wait()
}

// func receiver(ch chan int) {
// 	time.Sleep(20 * time.Second)
// 	ch <- 2
// }

// func connect() {

// 	listener, err := net.Listen("tcp", SERVER+":"+PORT)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for {
// 		conn, err := listener.Accept()
// 		if err != nil {
// 			log.Fatal(err)
// 			continue
// 		}
// 		go handleConnection(conn)
// 	}
// }

func client() {
	clientConn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		log.Fatal(err)
	}
	defer clientConn.Close()
	mustCopy(os.Stdout, clientConn)
}

func mustCopy(dst io.Writer, src io.Reader) {
	if _, err := io.Copy(dst, src); err != nil {
		log.Fatal(err)
	}
}

func unused(v ...any) {
	_ = v
}
