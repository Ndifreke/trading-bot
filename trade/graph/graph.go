package graph

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"sort"
	"trading/helper"
	"trading/kline"
)

var println = fmt.Println

func unused(v ...interface{}) {
}

type TrendType string

const (
	Uptrend   TrendType = "Uptrend"
	DownTrend           = "DownTrend"
	Dumping             = "Dumping"
	Range               = "Range"
	Reversal            = "Reversal"
	Breakout            = "BreakOut"
	Limit               = "Limit"
	AutoTrend           = "Auto"
)

type Graph struct {
	kLine     kline.KLineInterface
	priceStep float64
}

func NewGraph(kLine kline.KLineInterface) *Graph {
	priceStep := 0.0
	return &Graph{
		kLine,
		priceStep,
	}
}

func (g *Graph) IsBreakOut() bool {
	return g.GetPriceMidpoint() > g.FindAverageEntryPoints().GainHighPrice
}

func (g *Graph) IsDumping() bool {
	return g.GetPriceMidpoint() < g.FindAverageEntryPoints().DipLowPrice
}

func (g *Graph) IsUptrend() bool {
	// We need to set a high and low for uptrend otherwise uptrend will be true
	// when breakout is also true as they both move in upper direction
	entryPoints := g.FindAverageEntryPoints()
	uptrendHighPrice := entryPoints.GainHighPrice
	dipLowPrice := entryPoints.DipLowPrice
	priceRangeOneThird := (uptrendHighPrice - dipLowPrice) / 3

	midPoint := g.GetPriceMidpoint()
	uptrendBasePrice := uptrendHighPrice - priceRangeOneThird

	// Ensure that the midpoint is contained between the uptrendHighPrice and uptrendBasePrice
	return midPoint >= uptrendBasePrice && midPoint <= uptrendHighPrice
}

func (g *Graph) IsDowntrend() bool {
	// We need to set a high and low for downtrend otherwise downtrend will be true
	// when IsDumping is also true as they both move in downward direction
	entryPoints := g.FindAverageEntryPoints()
	uptrendHighPrice := entryPoints.GainHighPrice
	dipLowPrice := entryPoints.DipLowPrice
	priceRangeOneThird := (uptrendHighPrice - dipLowPrice) / 3
	midPoint := g.GetPriceMidpoint()
	downtrendCeiling := dipLowPrice + priceRangeOneThird
	// Ensure that the midpoint is contained between the downTrendCeiling and dipLowPrice
	return midPoint <= downtrendCeiling && midPoint >= dipLowPrice
}

// GetAveragePrice + for up - for down starting from center
func (g *Graph) GetAveragePrice() float64 {
	kLineData := g.getKLineData()
	count, sum := float64(len(kLineData)), 0.0
	for i := 1; i < len(kLineData); i += 2 {
		a, b := kLineData[i-1], kLineData[i]
		sum += a.Close + b.Close
	}
	return sum / count
}

func (g *Graph) getKLineData() []kline.KlineData {
	return g.kLine.KLineData()
}

func (g *Graph) CalculateAveragePriceMovement() float64 {
	kLineData := g.getKLineData()
	var priceStepSum float64
	count := float64(len(kLineData))
	for _, kline := range kLineData {
		max, min := math.Max(kline.High, kline.Low), math.Min(kline.High, kline.Low)
		//Experimental, best result seems to be Open
		priceStepSum += (max - min)

	}

	g.priceStep = priceStepSum / count
	return g.priceStep
}

// func (g *Graph) GetAveragePriceChange() float64 {
// 	kLineData := g.getKLineData()
// 	count := float64(len(kLineData))
//     var sum float64
//     for i, price := range kLineData {
//         if i > 0 {
//             priceChange := price.Close - kLineData[i-1].Close
//             sum += priceChange
//         }
//     }

//     average := sum / float64(count-1)
//     return average
// }

type HighLowData struct{ High, Low float64 }

// Get the highest and lowest price of all the candles in a graph.
func (g *Graph) GetHighLowPrice() HighLowData {
	High, Low := math.Inf(-1), math.Inf(1)
	for _, data := range g.getKLineData() {
		High, Low = math.Max(High, data.High), math.Min(Low, data.Low)
	}
	return HighLowData{High, Low}
}

func (g *Graph) Refresh() *Graph {
	g.kLine.RefreshData()
	return g
}

func (g *Graph) Kline() []kline.KlineData {
	return g.getKLineData()
}

// Calculates the mean of the high and low prices of all the candles in a graph.
// Generally price should be below this point to buy and above to sell
func (g Graph) GetPriceMidpoint() float64 {
	var sumHighAndLow float64
	for _, d := range g.getKLineData() {
		sumHighAndLow = (d.Open + d.Close)
	}
	return sumHighAndLow / 2
}

// func (g Graph) GetPriceMidpoint() float64 {
// 	var highs,lows float64
// 	for _, k := range g.getKLineData() {
// 		highs += k.High;
// 		lows += k.Low
// 	}
// 	return (highs + lows) / 2
// }

/*
*
3m 20 = 1h high volotility
5m 12 = 1h medium volotility
15m 8 2h
*
*/
func NewBinanceGraph(symbol, interval /*1m for optimal*/ string, pointsLimit int /*15 for optimal*/) *Graph {
	return NewGraph(kline.NewKline(symbol, interval, pointsLimit))
}

func (graph Graph) DetermineTrend() TrendType {
	if len(graph.getKLineData()) == 0 {
		return AutoTrend
	}
	if graph.IsBreakOut() {
		return Breakout
	}
	if graph.IsDumping() {
		return Dumping
	}
	if graph.IsDowntrend() {
		return DownTrend
	}
	if graph.IsUptrend() {
		return Uptrend
	}
	return Range
}

func (graph *Graph) DetermineSMAtrend() TrendType {
	data := graph.getKLineData()
	sma := graph.CalculateSMA()
	dataCount := len(data)
	currentPrice := data[dataCount-1].Close

	if currentPrice > sma {
		return Uptrend
	} else if currentPrice < sma {
		return DownTrend
	} else {
		return Range
	}
}

func (graph *Graph) SaveToFile(filename string) error {

	var d = struct {
		TrendName          TrendType          `json:"trendName"`
		PriceMidpoint      float64            `json:"priceMidpoint"`
		EntryPoints        averageEntryPoints `json:"entryPoints"`
		Data               []kline.KlineData  `json:"data"`
		AveragePriceChange float64            `json:"averagePriceChange"`
	}{
		TrendName:          graph.DetermineTrend(),
		Data:               graph.getKLineData(),
		PriceMidpoint:      graph.GetPriceMidpoint(),
		EntryPoints:        graph.FindAverageEntryPoints(),
		AveragePriceChange: graph.CalculateAveragePriceMovement(),
	}

	data, err := json.Marshal(d)

	if err != nil {
		return err
	}

	err = ioutil.WriteFile("./ui/src/dump/kline.json", data, 0644)
	
	if err != nil {
		return err
	}
	return nil
}

type averageEntryPoints struct {
	DipLowerPrice float64 `json:"dipLowerPrice"`
	DipLowPrice   float64 `json:"dipLowPrice"`
	GainLowPrice  float64 `json:"gainLowPrice"`
	GainHighPrice float64 `json:"gainHighPrice"`
}

// Find average entry point base on no sentiment
func (graph *Graph) FindAverageEntryPoints() averageEntryPoints {

	data := graph.getKLineData()
	itemCount := len(data)
	var lows, highs []float64

	for _, d := range data {
		lows, highs = append(lows, d.Low), append(highs, d.High)
	}
	// reverse the sorted data to get values that is closer to the mean  by using the lowest highes and highest lows to determine the value
	// the closer the value to the mean the easier it will be to sell/buy

	// More profit, but difficult to sell, sorting uses low dips Lowest-Lows
	// Calculate the low and lower dips

	sort.Float64s(lows)
	var dipLowerSum float64
	for _, d := range lows[0 : itemCount/2] {
		dipLowerSum += d
	}
	DipLowerPrice := dipLowerSum / float64(itemCount/2)

	var dipLowSum float64
	for _, d := range lows {
		dipLowSum += d
	}
	DipLowPrice := dipLowSum / float64(itemCount)

	// Calculate the High and Higher gains
	sort.Float64s(highs)
	var gainLowSum float64
	for _, d := range highs[0 : itemCount/2] {
		gainLowSum += d
	}
	GainLowPrice := gainLowSum / float64(itemCount/2)

	var gainHighSum float64
	for _, d := range highs[itemCount/2:] {
		gainHighSum += d
	}
	GainHighPrice := gainHighSum / float64(itemCount/2)

	return averageEntryPoints{
		DipLowerPrice,
		DipLowPrice,
		GainLowPrice,
		GainHighPrice,
	}
}

// Use the sentiment of the trend to determine the entry points
func (graph *Graph) DetermineEntryPointsFromTrend() {

	klines := graph.getKLineData()
	marketPull := graph.GetCandleSpikePull()
	trend := graph.DetermineTrend()
	itemCount := len(klines)
	switch trend {

	case Dumping:
		//RANGE TODO
		//find the last two price that is lower than the current price and
		//find their difference as th entry point. if it is a range graph
		unused(itemCount, marketPull)
	default:
	}
}

func GetBestTradingSymbol() {
	//list of trading pairs and get the one with either breakout or
	//upper trendpair graph
}

type marketSentiment string

const (
	Bear marketSentiment = "Bear"
	Bull marketSentiment = "Bull"
)

type MarketPullForce struct {
	BullPull  float64
	BearPull  float64
	Sentiment marketSentiment
}

type TrendPullForce struct {
	SentimentPercent float64
	Sentiment        marketSentiment
	BullPull         float64
	BearPull         float64
}

/*
The trend force similar to Market pull force
SentimentPercent how much percentage of the opposite of sentiment is the sentiment
*/
func (graph *Graph) GetTrendPullForce() TrendPullForce {
	Sentiment := Bear
	kline := graph.getKLineData()
	var Bulltrend, Beartrend float64
	for _, k := range kline {
		Bulltrend += k.High - k.Open
		Beartrend += k.Open - k.Low
	}

	if Bulltrend > Beartrend {
		Sentiment = Bull
	}
	max, min := math.Max(Bulltrend, Beartrend), math.Min(Bulltrend, Beartrend)
	return TrendPullForce{
		BullPull:         Bulltrend,
		BearPull:         Beartrend,
		Sentiment:        Sentiment,
		SentimentPercent: helper.GrowthPercent(max, min),
	}
}

/*
Computes the sentiment of candles to determine between the bull and the bear
who is the most greedy. It uses disance between (open, close) to high or low.
This is usefull when you want to predict the short term of if the current candle
will grow or shrink based on it previous spikes
*/
func (graph *Graph) GetCandleSpikePull() MarketPullForce {
	sentiment := Bear

	var closePull, openPull float64
	for _, k := range graph.getKLineData() {
		distanceToHigh := k.High - math.Max(k.Open, k.Close)
		openPull += distanceToHigh

		distanceToLow := math.Min(k.Open, k.Close) - k.Low
		closePull += distanceToLow
	}

	if closePull == openPull {
		// the pull direction is the same lets use the trend to conclude
		if graph.IsBreakOut() || graph.IsUptrend() {
			sentiment = Bull
		}
	}
	if openPull > closePull {
		sentiment = Bull
	}

	return MarketPullForce{
		Sentiment: sentiment,
		BullPull:  openPull,
		BearPull:  closePull,
	}
}

func (graph *Graph) CalculateSMA() float64 {

	data := graph.getKLineData()
	dataCount := len(data)
	windowSize := graph.kLine.GetPointLimit()
	if dataCount < windowSize {
		return 0.0
	}
	sum := 0.0
	for i := dataCount - windowSize; i < dataCount; i++ {
		sum += data[i].Close
	}
	return sum / float64(windowSize)
}
