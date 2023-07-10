package graph

import (
	"testing"
	"trading/kline"

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

	upGraph := NewGraph(kline.GetKLineMock(uptrendPullupData))
	assert.Equal(t, upGraph.GetCandleSpikePull().Sentiment, Bull)
	assert.Equal(t, upGraph.GetCandleSpikePull().BullPull, 25.0)
	assert.Equal(t, upGraph.GetCandleSpikePull().BearPull, 5.0)

	downGraph := NewGraph(kline.GetKLineMock(downtrendPullupData))
	assert.Equal(t, downGraph.GetCandleSpikePull().Sentiment, Bear)
	assert.Equal(t, downGraph.GetCandleSpikePull().BullPull, 2.0)
	assert.Equal(t, downGraph.GetCandleSpikePull().BearPull, 5.0)

	assert.Equal(t, downGraph.GetPriceMidpoint(), 17.0)
}
