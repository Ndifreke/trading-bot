package graph

// import (
// 	"testing"
// 	"trading/kline"

// 	"github.com/stretchr/testify/assert"
// )



// func TestTradeLocker(t *testing.T) {

// 	upGraph := NewGraph(kline.GetKLineMock(kline.MockKLinePullUpData))
// 	assert.Equal(t, upGraph.GetCandleSpikePull().Sentiment, Bull)
// 	assert.Equal(t, upGraph.GetCandleSpikePull().BullPull, 25.0)
// 	assert.Equal(t, upGraph.GetCandleSpikePull().BearPull, 5.0)

// 	downGraph := NewGraph(kline.GetKLineMock(kline.MockKLineData))
// 	assert.Equal(t, downGraph.GetCandleSpikePull().Sentiment, Bear)
// 	assert.Equal(t, downGraph.GetCandleSpikePull().BullPull, 2.0)
// 	assert.Equal(t, downGraph.GetCandleSpikePull().BearPull, 5.0)

// 	assert.Equal(t, downGraph.GetPriceMidpoint(), 17.0)
// }
