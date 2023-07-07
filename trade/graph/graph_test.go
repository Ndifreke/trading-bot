package graph

import (
	"testing"
	"trading/binance/kline"

	"github.com/stretchr/testify/assert"
)

var uptrendPullupData = []kline.KlineData{
	{Open: 15.0, Close: 10.0, High: 40.0, Low: 5.0},
}
var downtrendPullupData = []kline.KlineData{
	//the high is in the body of the bar
	{Open: 10.0, Close: 10.0, High: 12.0, Low: 5.0},
	{Open: 10.0, Close: 10.0, High: 12.0, Low: 5.0},
	{Open: 9.0, Close: 5.0, High: 1.0, Low: 5.0},
}

func TestTradeLocker(t *testing.T) {
	upKline := NewGraph(kline.GetKLineMock(uptrendPullupData))
	assert.Equal(t, upKline.GetMarketForcePull().Sentiment, Bull)
	assert.Equal(t, upKline.GetMarketForcePull().BullPull, 25.0)
	assert.Equal(t, upKline.GetMarketForcePull().BearPull, 5.0)

	downKline := NewGraph(kline.GetKLineMock(downtrendPullupData))
	assert.Equal(t, downKline.GetMarketForcePull().Sentiment, Bear)
	assert.Equal(t, downKline.GetMarketForcePull().BullPull, 2.0)
	assert.Equal(t, downKline.GetMarketForcePull().BearPull, 5.0)

	assert.Equal(t, downKline.GetPriceMidpoint(), 17.0)
}
