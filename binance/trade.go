package binance

import (
	"context"
	"encoding/json"
	"trading/utils"
	"net/http"
	"sort"
	"github.com/adshao/go-binance/v2"
)

func GetTradeHistories(symbol string) []*binance.Trade {
	s, _ := GetClient().NewHistoricalTradesService().Symbol(symbol).Do(context.Background())
	return s
}

type PriceChangeStats struct {
	Symbol             string  `json:"symbol"`
	PriceChange        string  `json:"priceChange"`
	PriceChangePercent float64 `json:"priceChangePercent,string"` // Note the ",string" tag
	WeightedAvgPrice   string  `json:"weightedAvgPrice"`
	PrevClosePrice     string  `json:"prevClosePrice"`
	LastPrice          string  `json:"lastPrice"`
	LastQty            string  `json:"lastQty"`
	BidPrice           string  `json:"bidPrice"`
	BidQty             string  `json:"bidQty"`
	AskPrice           string  `json:"askPrice"`
	AskQty             string  `json:"askQty"`
	OpenPrice          string  `json:"openPrice"`
	HighPrice          string  `json:"highPrice"`
	LowPrice           string  `json:"lowPrice"`
	Volume             string  `json:"volume"`
	QuoteVolume        string  `json:"quoteVolume"`
	OpenTime           int64   `json:"openTime"`
	CloseTime          int64   `json:"closeTime"`
	FristID            int64   `json:"firstId"`
	LastID             int64   `json:"lastId"`
	Count              int64   `json:"count"`
}

func GetSymbolStats() []PriceChangeStats {
	req, err := http.Get("https://api.binance.com/api/v3/ticker/24hr")

	if err != nil {
		utils.LogError(err, "<GetSymbolStats> request error")
		return []PriceChangeStats{}
	}

	defer req.Body.Close()

	decoder := json.NewDecoder(req.Body)
	var data []PriceChangeStats
	if err := decoder.Decode(&data); err != nil {
		return nil
	}
	sort.Slice(data, func(a, b int) bool { return data[a].PriceChangePercent > data[b].PriceChangePercent })

	return data
}
