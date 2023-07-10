package locker

import (
	"testing"
	"trading/names"

	"github.com/stretchr/testify/assert"
)

const (
	sellMinimum = 5
)

func TestTradeLockerSell(t *testing.T) {

	tradeLocker := NewTradeLocker()

	tc1 := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 50,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 10,
		},
	}
	tc2 := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 16,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 2,
		},
	}

	lockOne := tradeLocker.AddLock(tc1, 10).(*Lock)
	lockTwo := tradeLocker.AddLock(tc2, 5).(*Lock)
	assert.Equal(t, lockOne.GetLockOwner(), lockTwo.GetLockOwner(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.PretradePrice(), float64(10), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetLossLimit(), float64(50), "Lock stop limit to be set")

	assert.Equal(t, lockOne.getMinimumLockUnit(),
		float64(sellMinimum), "Minimum lock unit to be in percentage of stopLimit")

	// assert.Equal(t, lockOne.LockedPrice(), float32(10), "Expect tc1Lock initial locked price to equal initial price")
	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is less than stopLimit")

	lockOne.TryLockPrice(60)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is increasing and not dip below delta")

	lockOne.TryLockPrice(54)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next lock price drops below the minum lock amount but still profitable")
	lockOne.TryLockPrice(59)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if the dips below minimum and increases again")

	assert.Equal(t, lockTwo.PretradePrice(), float64(5))
	assert.Equal(t, lockTwo.GetLossLimit(), float64(16), "<Lock2> stop limit to be set")
	assert.Equal(t, lockTwo.getMinimumLockUnit(),
		float64(0.32), "<Lock2> Minimum lock unit to be in percentage of stopLimit")
	lockTwo.TryLockPrice(4)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if cuurent price is less than stopLimit")

	lockTwo.TryLockPrice(50)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> not redeamable if cuurent price is greater than stopLimit")

	lockTwo.TryLockPrice(49)
	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

	lockOne.TryLockPrice(54)
	assert.False(t, lockOne.IsRedemptionCandidate(), "should not be redeamable if it deeps below limit but still positive and is not the higest candidate with gains")
	assert.True(t, lockOne.IsRedemptionDue(), "<Lock1> As a single redemption it should be due")
	assert.False(t, lockOne.isHighestLockAction())
	assert.True(t, lockTwo.isHighestLockAction())

	cfg2 := names.TradeConfig{
		Symbol: "BNBBUSD",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit:  247.911768,
			RateType:  names.RatePercent,
			Quantity:  0,
			LockDelta: 0.32838497606132194,
		},
	}

	lockThree := tradeLocker.AddLock(cfg2, 247.100000).(*Lock)
	lockThree.TryLockPrice(246.200000)
	assert.False(t, lockThree.IsRedemptionDue())
}

func TestTradeLockerBuy(t *testing.T) {

	tradeLocker := NewTradeLocker()

	tc1 := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 20,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 10,
		},
	}
	tc2 := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 5,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 2,
		},
	}

	lockOne := tradeLocker.AddLock(tc1, 50).(*Lock)
	lockTwo := tradeLocker.AddLock(tc2, 16).(*Lock)
	assert.Equal(t, lockOne.GetLockOwner(), lockTwo.GetLockOwner(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.PretradePrice(), float64(50), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetLossLimit(), float64(20), "Lock stop limit to be set")

	assert.Equal(t, lockOne.getMinimumLockUnit(),
		float64(2), "Minimum lock unit to be in percentage of stopLimit")

	lockOne.TryLockPrice(22)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")

	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent buy price is decreases and has not made a rebound by at least minimum lock")

	lockOne.TryLockPrice(15)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next buy price drops increases by minimum lock amount but still profitable")
	lockOne.TryLockPrice(13)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if buy pices is readeamable but next minimum price increase returns it back into the dips")

	assert.Equal(t, lockTwo.PretradePrice(), float64(16))
	assert.Equal(t, lockTwo.GetLossLimit(), float64(5), "<Lock2> stop limit to be set")
	assert.Equal(t, lockTwo.getMinimumLockUnit(),
		float64(0.1), "<Lock2> Minimum lock unit to be in percentage of stopLimit")
	lockTwo.TryLockPrice(11)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> not redeamable if current price is greater than stoplimit")

	lockTwo.TryLockPrice(3)
	lockTwo.TryLockPrice(3.1)
	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

	lockOne.TryLockPrice(0.1)
	// lockOne.TryLockPrice(5)
	assert.False(t, lockOne.IsRedemptionCandidate(), "should not be redeamable if it has the highest increase but redemption is not due")

	assert.True(t, lockTwo.IsRedemptionDue(), "should be due for redemption after minimum increase as a single lock")
	assert.False(t, lockTwo.IsRedemptionCandidate(), "should not be redeamable if redemption is not due but is not the highest lock")

	assert.True(t, lockOne.isHighestLockAction())
	assert.False(t, lockTwo.isHighestLockAction())

	cfg2 := names.TradeConfig{
		Symbol: "BNBBUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 2,
			RateType:  names.RatePercent,
			Quantity:  0,
			LockDelta: 1,
		},
	}

	lockThree := tradeLocker.AddLock(cfg2, 247).(*Lock)
	lockThree.TryLockPrice(246)
	assert.False(t, lockThree.IsRedemptionDue())
}

// func TestTradeLockerSpecialCases(t *testing.T) {

// 	tradeLocker := NewTradeLocker()

// 	buyLockConfig := names.TradeConfig{
// 		Symbol: "BTCUSD",
// 		Side:   names.TradeSideBuy,
// 		Buy: names.SideConfig{
// 			RateLimit: 0,
// 			RateType:  names.RatePercent,
// 			Quantity:  10,
// 			LockDelta: 0,
// 		},
// 	}
// 	sellLockConfig := names.TradeConfig{
// 		Symbol: "ETHUSDC",
// 		Side:   names.TradeSideBuy,
// 		Sell: names.SideConfig{
// 			RateLimit: 0,
// 			RateType:  names.RatePercent,
// 			Quantity:  20,
// 			LockDelta: 0,
// 		},
// 	}

// 	buyLock := tradeLocker.AddLock(buyLockConfig, 50, 20).(*Lock)
// 	lockTwo := tradeLocker.AddLock(sellLockConfig, 16, 5).(*Lock)
// 	assert.Equal(t, buyLock.GetLockOwner(), lockTwo.GetLockOwner(), "Expected TradeLocker to be the same")

// 	assert.Equal(t, buyLock.PretradePrice(), float64(50), "Lock initial price to be 10")
// 	assert.Equal(t, buyLock.GetLossLimit(), float64(20), "Lock stop limit to be set")

// 	assert.Equal(t, buyLock.getMinimumLockUnit(),
// 		float64(2), "Minimum lock unit to be in percentage of stopLimit")

// 	buyLock.TryLockPrice(22)
// 	assert.False(t, buyLock.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")

// 	buyLock.TryLockPrice(10)
// 	assert.False(t, buyLock.IsRedemptionCandidate(), "Not redeamable if cuurent buy price is decreases and has not made a rebound by at least minimum lock")

// 	buyLock.TryLockPrice(15)
// 	assert.True(t, buyLock.IsRedemptionCandidate(), "Lock should be redeamable if the next buy price drops increases by minimum lock amount but still profitable")
// 	buyLock.TryLockPrice(13)
// 	assert.False(t, buyLock.IsRedemptionCandidate(), "Lock should not be readeamable if buy pices is readeamable but next minimum price increase returns it back into the dips")

// 	assert.Equal(t, lockTwo.PretradePrice(), float64(16))
// 	assert.Equal(t, lockTwo.GetLossLimit(), float64(5), "<Lock2> stop limit to be set")
// 	assert.Equal(t, lockTwo.getMinimumLockUnit(),
// 		float64(0.1), "<Lock2> Minimum lock unit to be in percentage of stopLimit")
// 	lockTwo.TryLockPrice(11)
// 	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> not redeamable if current price is greater than stoplimit")

// 	lockTwo.TryLockPrice(3)
// 	lockTwo.TryLockPrice(3.1)
// 	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

// 	buyLock.TryLockPrice(0.1)
// 	// lockOne.TryLockPrice(5)
// 	assert.False(t, buyLock.IsRedemptionCandidate(), "should not be redeamable if it has the highest increase but redemption is not due")

// 	assert.True(t, lockTwo.IsRedemptionDue(), "should be due for redemption after minimum increase as a single lock")
// 	assert.False(t, lockTwo.IsRedemptionCandidate(), "should not be redeamable if redemption is not due but is not the highest lock")

// 	assert.True(t, buyLock.isHighestLockAction())
// 	assert.False(t, lockTwo.isHighestLockAction())

// 	cfg2 := names.TradeConfig{
// 		Symbol: "BNBBUSD",
// 		Side:   names.TradeSideBuy,
// 		Sell: names.SideConfig{
// 			RateLimit: 0,
// 			RateType:  names.RatePercent,
// 			Quantity:  0,
// 			LockDelta: 0.32838497606132194,
// 		},
// 	}

// 	lockThree := tradeLocker.AddLock(cfg2, 247.100000, 247.911768).(*Lock)
// 	lockThree.TryLockPrice(246.200000)
// 	assert.False(t, lockThree.IsRedemptionDue())
// }
