package binance

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	"trading/api"
)

type ServerTimeJson struct {
	ServerTime int64 `json:"serverTime"`
}

type PriceJson struct {
	Mins  int     `json:"mins"`
	Price float32 `json:"price,string,omitempty"`
}

func GetPriceAverage(symbol string) api.RequestResponse[PriceJson] {
	var bn = New[PriceJson](apiArg{Api: Endpoints.PriceAverage})
	params := map[string]string{"symbol": strings.ToUpper(symbol)}
	return bn.RequestWithQuery(params)
}

func GetPriceLatest(symbol string) api.RequestResponse[PriceJson] {
	var bn = New[PriceJson](apiArg{Api: Endpoints.PriceLatest})
	params := map[string]string{"symbol": strings.ToUpper(symbol)}
	return bn.RequestWithQuery(params)
}


type KLineJson struct {
	OpenTime          int64
	OpenPrice         float64
	HighestPrice      string
	LowestPrice       string
	ClosePrice        string
	CloseTime         int64
	TradingPairCount  float64
	TotalTradingValue float32
	TradeCount        int
}

func GetKLine(symbol, interval string) []KLineJson {
	var bn = New[[][]any](apiArg{Api: Endpoints.KLines})
	param := map[string]string{
		"symbol":   symbol,
		"interval": interval,
	}
	res := bn.RequestWithQuery(param)
	if res.Error != nil {
		panic(res.Error)
	}
	var d []KLineJson
	for _, kLine := range res.Body {
		OpenPrice, _ := strconv.ParseFloat(string(kLine[1].(string)), 64)

		k := KLineJson{
			OpenTime:         int64(kLine[0].(float64)),
			OpenPrice:        OpenPrice,
			HighestPrice:     string(kLine[2].(string)),
			LowestPrice:      string(kLine[3].(string)),
			ClosePrice:       string(kLine[4].(string)),
			CloseTime:        int64(kLine[6].(float64)),
			TradingPairCount: float64(kLine[8].(float64)),
			// TotalTradingValue: 0,
			// TradeCount:        0,
		}
		d = append(d, k)
	}
	return d
}

type OrderJson struct {
	Symbol              string  `json:"symbol"`
	OrderId             int     `json:"orderId"`
	OrderListId         int     `json:"orderListId"`
	ClientOrderId       string  `json:"clientOrderId"`
	Price               string  `json:"price"`
	OrigQty             string  `json:"origQty"`
	ExecutedQuantity    string  `json:"executedQty"`
	CummulativeQuoteQty string  `json:"cummulativeQuoteQty"`
	Status              string  `json:"status"`
	TimeInForce         string  `json:"timeInForce"`
	Type                string  `json:"type"`
	Side                string  `json:"side"`
	StopPrice           string  `json:"stopPrice"`
	IcebergQty          string  `json:"icebergQty"`
	Time                float64 `json:"time"`
	UpdateTime          int64   `json:"updateTime"`
	IsWorking           bool    `json:"isWorking"`
	OrigQuoteOrderQty   string  `json:"origQuoteOrderQty"`
}

func GetOpenOrders() {
	var bn = New[OrderJson](apiArg{Api: Endpoints.OpenOrders})
	var timestamp = fmt.Sprintf("%d", time.Now().Unix()*1000)
	r := bn.RequestWithQuery(map[string]string{
		"signature":  getSignature(timestamp),
		"timestamp":  fmt.Sprint(timestamp),
		"recvWindow": "60000",
	})
	_ = r
	// fmt.Print(r.Ok,r.Body,r.Response)
}
