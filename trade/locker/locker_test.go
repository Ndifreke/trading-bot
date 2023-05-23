package locker

import (
	"testing"
	"trading/trade"

	"github.com/stretchr/testify/assert"
)

const (
	sellMinimum = 5
)

func TestTradeLocker(t *testing.T) {

	tradeLocker := NewTradeLocker()

	tc1 := trade.TradeConfig{
		Symbol: "BTCUSD",
		Side:   trade.TradeSideSell,
		Price: struct {
			Sell trade.Price
			Buy  trade.Price
		}{
			Buy: trade.Price{
				PriceRate:     1,
				PriceRateType: trade.RatePercent,
				Quantity:      2,
			},
			Sell: trade.Price{
				PriceRate:     1,
				PriceRateType: trade.RatePercent,
				Quantity:      2,
				LockDelta:     10,
			},
		},
	}
	tc2 := trade.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   trade.TradeSideSell,
		Price: struct {
			Sell trade.Price
			Buy  trade.Price
		}{
			Buy: trade.Price{
				PriceRate:     1,
				PriceRateType: trade.RatePercent,
				Quantity:      2,
			},
			Sell: trade.Price{
				PriceRate:     1,
				PriceRateType: trade.RatePercent,
				Quantity:      2,
				LockDelta:     2,
			},
		},
	}

	lockOne := tradeLocker.AddLock(tc1, 10, 50)
	lockTwo := tradeLocker.AddLock(tc2, 5, 16)
	assert.Equal(t, lockOne.GetLockOwner(), lockTwo.GetLockOwner(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.GetBasePrice(), float32(10), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetLossLimit(), float32(50), "Lock stop limit to be set")

	assert.Equal(t, lockOne.getMinimumLockUnit(),
		float32(sellMinimum), "Minimum lock unit to be in percentage of stopLimit")

	// assert.Equal(t, lockOne.LockedPrice(), float32(10), "Expect tc1Lock initial locked price to equal initial price")
	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is less than stopLimit")

	lockOne.TryLockPrice(60)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is increasing and not dip below delta")

	lockOne.TryLockPrice(54)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next lock price drops below the minum lock amount but still profitable")
	lockOne.TryLockPrice(59)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if the dips below minimum and increases again")

	assert.Equal(t, lockTwo.GetBasePrice(), float32(5))
	assert.Equal(t, lockTwo.GetLossLimit(), float32(16), "<Lock2> stop limit to be set")
	assert.Equal(t, lockTwo.getMinimumLockUnit(),
		float32(0.32), "<Lock2> Minimum lock unit to be in percentage of stopLimit")
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
}
