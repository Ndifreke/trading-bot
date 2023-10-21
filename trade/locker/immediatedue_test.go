package locker

import (
	"testing"
	"trading/names"

	"github.com/stretchr/testify/assert"
)

func TestFreeLockSellBuy(t *testing.T) {

	tradeLocker := NewLockManager(ImmediateDueLockCreator)

	tc1 := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			StopLimit: 20,
			LimitType: names.RateFixed,
			Quantity:  2,
			LockDelta: 10,
		},
	}
	tc2 := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			StopLimit: 5,
			LimitType: names.RateFixed,
			Quantity:  2,
			LockDelta: 2,
		},
	}

	lockOne := tradeLocker.AddLock(tc1, 50).(*immediateDueLock)
	lockTwo := tradeLocker.AddLock(tc2, 16).(*immediateDueLock)
	assert.Equal(t, lockOne.GetLockManager(), lockTwo.GetLockManager(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.PretradePrice(), float64(50), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetTradeLimit(), float64(20), "Lock stop limit to be set")

	assert.EqualValues(t, lockOne.getMinimumLockUnit(), 5, "Minimum lock unit to be in percentage of pretrade price")

	lockOne.TryLockPrice(22)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")

	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent buy price is decreases and has not made a rebound by at least minimum lock")

	lockOne.TryLockPrice(15)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next buy price drops increases by minimum lock amount but still profitable")
	lockOne.TryLockPrice(9)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if buy pices is readeamable but next minimum price increase returns it back into the dips")

	assert.Equal(t, lockTwo.PretradePrice(), float64(16))
	assert.Equal(t, lockTwo.GetTradeLimit(), float64(5), "<Lock2> stop limit to be set")
	assert.EqualValues(t, lockTwo.getMinimumLockUnit(), 0.32, "Minimum lock unit to be in percentage of  pretrade price")
	lockTwo.TryLockPrice(11)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "not redeamable if current price is greater than stoplimit")

	lockTwo.TryLockPrice(1)
	lockTwo.TryLockPrice(2)
	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

	lockOne.TryLockPrice(0.1)
	// lockOne.TryLockPrice(5)
	assert.False(t, lockOne.IsRedemptionCandidate(), "should not be redeamable if it has the highest increase but redemption is not due")

	assert.True(t, lockTwo.IsRedemptionDue(), "should be due for redemption after minimum increase as a single lock")
	assert.True(t, lockOne.AbsoluteGrowthPercent() > lockTwo.AbsoluteGrowthPercent(), "should have the highest lock")
	assert.True(t, lockTwo.IsRedemptionCandidate(), "should be redeamable if redemption is due and regardless of highest lock")

	cfg2 := names.TradeConfig{
		Symbol: "BNBBUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			StopLimit: 2,
			LimitType: names.RatePercent,
			Quantity:  0,
			LockDelta: 1,
		},
	}

	lockThree := tradeLocker.AddLock(cfg2, 247).(*immediateDueLock)
	lockThree.TryLockPrice(246)
	assert.False(t, lockThree.IsRedemptionDue())
}
