package main

import (
	"fmt"
	"io"
	"log"
	"net"

	// "net/http"
	"os"
	"trading/api"
	// "trading/helper"
	"trading/names"
	"trading/trade/limit"

	"sync"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/joho/godotenv"

	// "trading/binance"
	detection "trading/trade/graph"
	"trading/trade/manager"
)

type args = api.ApiArg

const (
	SERVER = "localhost"
	PORT   = "8080"
)

func init() {
	godotenv.Load()
}

func do[K float32](j K) {
	time.Sleep(0)
	// defer wg.Done()
}

func a(ch chan int) {
	do[float32](1.9998)

	ch <- 9
	time.Sleep(time.Second * 2)
	// fmt.Print(wg)
	// defer wg.Done()
}

func b(ch chan int, wg *sync.WaitGroup) {
	// fmt.Print(wg)
	defer wg.Done()

}

var (
	count int = 0
	c         = 300
	me    sync.Mutex
)

func counting(wg *sync.WaitGroup, i int, ch chan int) {
	fmt.Print(i, "This is I")
	for i := 0; i < c; i++ {
		fmt.Println(i)
		count++
	}
	defer wg.Done()
}

type Person struct {
	Name string
	any  interface{}
}

func (p *Person) setName(n string) {
	p.Name = n
}

func (p Person) String() {
	fmt.Print(p.Name)
}

func main() {
	var wg sync.WaitGroup

	_ = wg
	// bn := binance.New[binance.PriceJson](args{
	// 	Api: binance.Endpoints.PriceLatest,
	// })
	params := map[string]string{"symbol": "USDTAPT"}
	_ = params
	// r := bn.RequestWithQuery(params)
	// fmt.Print(r.Body.Price)
	// fmt.Print(binance.GetKLine("BTCBUSD","1m"))
	// d := binance.NewGraph("BNBBUSD","15m")
	// var j = func (conn *websocket.Conn, d request.MiniTickerData){
	// 	fmt.Println(d.StreamName)
	// 	// conn.Close()
	// }
	// s := request.PriceStream([]string{"cfxusdt","btcusdt"})
	// o := binance.GetOrderHistories("PEPEUSDT").ListSell().Latest()

	// fmt.Println(time.Unix(0, o.Time*int64(time.Millisecond)).Format(time.RFC1123), o.Side)
	// d, _ := binance.CreateOrder("BTCUSDT", 9, 9, "MARKET", "BUY")
	// unused(d)
	config := names.TradeConfig{
		Price: struct {
			Sell names.Price
			Buy  names.Price
		}{
			Sell: names.Price{
				RateLimit:  0,
				RateType:   names.RatePercent,
				Quantity:   1,
				MustProfit: true,
			},
			Buy: names.Price{
				RateLimit:  1,
				RateType:   names.RatePercent,
				Quantity:   1,
				MustProfit: true,
			},
		},
		Symbol: "BTCUSDT",
		Side:   names.TradeSideSell,
		// IsCyclick: true,
	}
	config2 := names.TradeConfig{
		Price: struct {
			Sell names.Price
			Buy  names.Price
		}{
			Sell: names.Price{
				RateLimit: 0,
				RateType:  names.RatePercent,
				Quantity:  200,
			},
			Buy: names.Price{
				RateLimit: 99,
				RateType:  names.RatePercent,
				Quantity:  1,
			},
		},
		Symbol: "BTCBUSD",
		Side:   names.TradeSideBuy,
	}
	a, b := names.Symbol("BNBBUSD"), names.Symbol("BTCUSDT")
	unused(b, a)
	v := a
	g := detection.NewBinanceGraph(v.String(), "15m", 18)
	fmt.Println(g.CalculateAveragePriceMovement(), "AVERAGE MOVEMENT")
	// fmt.Println(g.GetHighLowPrice(), "HIGH LOW")
	// fmt.Println(g.GetPriceMidpoint(), "MIDPOINT")
	g.SaveToFile("")

	fmt.Println(g.GetAveragePrice(), "Average")
	spew.Dump(g.GetTrendPullForce())
	// fmt.Println(g.GetPriceAvgDifference(), "aStep")

	// s.ReadMessage(j)
	// s.Connect()
	// fmt.Println(d.GetAveragePrice())
	// request.NewSocket("")
	// net.Dial()

	config3 := names.TradeConfig{
		Symbol: v,
		Side:   names.TradeSideSell,
		Price: struct{Sell names.Price; Buy names.Price}{},
	}
	_ = config2
	_ = config
	j := []names.TradeConfig{config3} //config

	wg.Add(1)
	// limit.NewLimitTradeManager(j...).Run()
	// tradeConfigs := helper.GenerateTradeConfigs(helper.TradeSymbolList)
	manager.NewTradeManager(j...).SetGraphParam("15m", 18).DoTrade()
	// manager.NewTradeManager(tradeConfigs...).SetGraphParam("15m", 18).DoTrade()
	unused(j)
	// unused(manager.NewTradeManager)
	unused(limit.BuyRun)
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




	

