package kline

import (
	"context"
	// "encoding/json"
	// "io/ioutil"
	"strconv"
	"trading/binance"
	"trading/utils"
)

type KlineData struct {
	// *binance.Kline
	Open                     float64 `json:"open"`
	High                     float64 `json:"high"`
	Low                      float64 `json:"low"`
	Close                    float64 `json:"close"`
	OpenTime                 int64   `json:"openTime"`
	CloseTime                int64   `json:"closeTime"`
	Volume                   float64 `json:"volume"`
	TradeNum                 int64   `json:"tradeNum"`
	TakerBuyBaseAssetVolume  float64 `json:"takerBuyBaseAssetVolume"`
	TakerBuyQuoteAssetVolume float64 `json:"takerBuyQuoteAssetVolume"`
	QuoteAssetVolume         float64 `json:"quoteAssetVolume"`
}

type KLine struct {
	symbol      string
	interval    string
	pointsLimit int
	data        []KlineData
}

type KLineInterface interface {
	KLineData() []KlineData
	RefreshData() []KlineData
	GetPointLimit() int
}

func NewKline(symbol, interval string, limit int) KLineInterface {
	//#MOCK
	if utils.Env().IsMock() {
		return GetKLineMock(MockKLineData)
	}
	if interval == "" {
		interval = "15m"
	}
	if limit == 0 {
		limit = 18
	}
	return &KLine{symbol: symbol, interval: interval, pointsLimit: limit}
}

func (kline *KLine) KLineData() []KlineData {
	if len(kline.data) == 0 {
		kline.data = GetKLineData(kline.symbol, kline.interval, kline.pointsLimit)
	}
	return kline.data
}

func GetKLineData(symbol, interval string, pointLimit int) []KlineData {
	kline, err := binance.GetClient().NewKlinesService().Interval(interval).Symbol(symbol).Limit(pointLimit).Do(context.Background())
	if err != nil {
		utils.LogError(err, "GetKLineData function call")
	}

	var klines []KlineData
	for _, k := range kline {
		Open, _ := strconv.ParseFloat(k.Open, 64)
		High, _ := strconv.ParseFloat(k.High, 64)
		Low, _ := strconv.ParseFloat(k.Low, 64)
		Close, _ := strconv.ParseFloat(k.Close, 64)
		OpenTime := k.OpenTime
		Volume, _ := strconv.ParseFloat(k.Volume, 64)
		TakerBuyBaseAssetVolume, _ := strconv.ParseFloat(k.TakerBuyBaseAssetVolume, 64)
		TakerBuyQuoteAssetVolume, _ := strconv.ParseFloat(k.TakerBuyQuoteAssetVolume, 64)
		QuoteAssetVolume, _ := strconv.ParseFloat(k.QuoteAssetVolume, 64)
		klines = append(klines, KlineData{
			TradeNum:                 k.TradeNum,
			TakerBuyBaseAssetVolume:  TakerBuyBaseAssetVolume,
			TakerBuyQuoteAssetVolume: TakerBuyQuoteAssetVolume,
			Open:                     Open,
			High:                     High,
			Low:                      Low,
			Close:                    Close,
			OpenTime:                 OpenTime,
			Volume:                   Volume,
			QuoteAssetVolume:         QuoteAssetVolume,
			CloseTime:                k.CloseTime,
		})
	}
	return klines
}

func (kline *KLine) RefreshData() []KlineData {
	data := GetKLineData(kline.symbol, kline.interval, kline.pointsLimit)
	kline.data = data
	return data
}

func (kline *KLine) GetPointLimit() int {
	return kline.pointsLimit
}
