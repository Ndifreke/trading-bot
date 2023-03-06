package binance

import (
	"fmt"
	"math"
)

type Graph struct {
	KLines    []KLineJson
	Symbol    string
	interval  string
	priceStep float64
}

func NewGraph(symbol, interval string) Graph {
	KLines := GetKLine(symbol, interval)
	return Graph{
		KLines:   KLines,
		Symbol:   symbol,
		interval: interval,
	}
}

func getKLine() {}

func (g Graph) GetTrend() string {
	return ""
}

// Get the average price of this graph over a period of time
func (g Graph) GetAveragePrice() float64 {
	klns := GetKLine(g.Symbol, g.interval)
	var sum, index float64
	var priceStep float64

	for i, k := range klns {

		if i%2 != 0 {
			diff := math.Max(k.OpenPrice, klns[i-1].OpenPrice) - math.Min(k.OpenPrice, klns[i-1].OpenPrice)
			priceStep += diff
			fmt.Println(priceStep)
		}

		f := k.OpenPrice
		sum += f
		index = float64(i)
	}
	g.priceStep = g.priceStep
	return sum / index
}

// Get average price change for a given duration change
func (g Graph) GetAveragePriceStep() float32 {
	if g.priceStep == 0 {
		g.GetAveragePrice()
	}
	return float32(g.priceStep)
}


func (g Graph) Refresh() Graph {
	g.GetAveragePrice()
	return g
}
