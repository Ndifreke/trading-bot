package main

import (
	"fmt"
	"io"
	"log"
	"net"

	// "net/http"
	"os"
	"trading/api"
	"trading/trade"
	"trading/trade/limit"

	"sync"
	"time"
	"trading/binance"

	"github.com/joho/godotenv"
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
	fmt.Println("Will Sleep")
	// defer wg.Done()
}

func a(ch chan int) {
	do[float32](1.9998)

	ch <- 9
	fmt.Println("First execution")
	time.Sleep(time.Second * 2)
	// fmt.Print(wg)
	// defer wg.Done()
}

func b(ch chan int, wg *sync.WaitGroup) {
	fmt.Println("Second execution")
	fmt.Println(<-ch)
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
	binance.CreateOrder("BTCUSDT",9,9,"MARKET", "BUY")
	// s.ReadMessage(j)
	// s.Connect()
	// fmt.Println(d.GetAveragePrice())
	// request.NewSocket("")
	// net.Dial()
	config := trade.TradeConfig{
		Price: struct {
			Sell trade.Price
			Buy  trade.Price
		}{
			Sell: trade.Price{
				PriceRate:     0,
				PriceRateType: trade.RatePercent,
				Quantity:      1,
				MustProfit:    true,
			},
			Buy: trade.Price{
				PriceRate:     0.01,
				PriceRateType: trade.RatePercent,
				Quantity:      1,
				MustProfit:    true,
			},
		},
		Symbol: "BNBBUSD",
		Side:   trade.TradeSideBuy,
		// IsCyclick: true,
	}
	config2 := trade.TradeConfig{
		Price: struct {
			Sell trade.Price
			Buy  trade.Price
		}{
			Sell: trade.Price{
				PriceRate:     0,
				PriceRateType: trade.RatePercent,
				Quantity:      200,
			},
			Buy: trade.Price{
				PriceRate:     99,
				PriceRateType: trade.RatePercent,
				Quantity:      1,
			},
		},
		Symbol: "BTCBUSD",
		Side:   trade.TradeSideBuy,
		// IsCyclick: true,
	}
	_ = config2
	_ = config
	// f := trade.GetTradeFee(config,"BUSDUSDT")
	j := []trade.TradeConfig{config}
	wg.Add(1)
	// limit.NewLimitTradeManager(j...).Run()
	unused(j)
	unused(limit.BuyRun)
	wg.Wait()

}

func receiver(ch chan int) {
	time.Sleep(20 * time.Second)
	ch <- 2
}

func connect() {

	listener, err := net.Listen("tcp", SERVER+":"+PORT)
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {

	defer conn.Close()
	for {
		_, err := io.WriteString(conn, "\nHello My friends")
		if err != nil {
			return
		}
		time.Sleep(time.Second)
	}
}

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

func unused(v any) {
	_ = v
}
