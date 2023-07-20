package locker

import (
	"testing"
	"trading/names"

	"github.com/stretchr/testify/assert"
)

func TestTradeLockerSell(t *testing.T) {

	tradeLocker := NewLockManager(HighLockCreator)

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

	lockOne := tradeLocker.AddLock(tc1, 10).(*highLock)
	lockTwo := tradeLocker.AddLock(tc2, 5).(*highLock)
	assert.Equal(t, lockOne.GetLockManager(), lockTwo.GetLockManager(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.PretradePrice(), float64(10), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetTradeLimit(), float64(50), "Lock stop limit to be set")

	assert.Equal(t, lockOne.getMinimumLockUnit(),
		float64(1), "Minimum lock unit to be in percentage of pretrade price")

	assert.EqualValues(t, lockOne.GetLockedPrice(), 10, "Expect tc1Lock initial locked price to equal initial price")
	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is less than stopLimit")

	lockOne.TryLockPrice(60)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is increasing and not dip below delta")

	lockOne.TryLockPrice(54)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next lock price drops below the minum lock amount but still profitable")
	lockOne.TryLockPrice(59)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if the dips below minimum and increases again")

	assert.Equal(t, lockTwo.PretradePrice(), float64(5))
	assert.Equal(t, lockTwo.GetTradeLimit(), float64(16), "<Lock2> stop limit to be set")
	assert.EqualValues(t, lockTwo.getMinimumLockUnit(),
		0.1, "<Lock2> Minimum lock unit to be in percentage of pretrade price")
	lockTwo.TryLockPrice(4)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if cuurent price is less than stopLimit")

	lockTwo.TryLockPrice(50)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> not redeamable if cuurent price is greater than stopLimit")

	lockTwo.TryLockPrice(49)
	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

	lockOne.TryLockPrice(54)
	assert.False(t, lockOne.IsRedemptionCandidate(), "should not be redeamable if it deeps below limit but still positive and is not the higest candidate with gains")
	assert.True(t, lockOne.IsRedemptionDue(), "<Lock1> As a single redemption it should be due")
	assert.False(t, lockOne == tradeLocker.BestMatureLock())
	assert.False(t, lockOne.IsRedemptionCandidate())
	assert.True(t, lockTwo == tradeLocker.BestMatureLock())
	assert.True(t, lockTwo.IsRedemptionCandidate())

	cfg2 := names.TradeConfig{
		Symbol: "BNBBUSD",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 247.911768,
			RateType:  names.RatePercent,
			Quantity:  0,
			LockDelta: 0.32838497606132194,
		},
	}

	lockThree := tradeLocker.AddLock(cfg2, 247.100000).(*highLock)
	lockThree.TryLockPrice(246.200000)
	assert.False(t, lockThree.IsRedemptionDue())
}

func TestTradeLockerBuy(t *testing.T) {

	tradeLocker := NewLockManager(HighLockCreator)

	configBuyOne := names.TradeConfig{
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

	lockOne := tradeLocker.AddLock(configBuyOne, 50).(*highLock)
	lockTwo := tradeLocker.AddLock(tc2, 16).(*highLock)
	assert.Equal(t, lockOne.GetLockManager(), lockTwo.GetLockManager(), "Expected TradeLocker to be the same")

	assert.Equal(t, lockOne.PretradePrice(), float64(50), "Lock initial price to be 10")
	assert.Equal(t, lockOne.GetTradeLimit(), float64(20), "Lock stop limit to be set")

	assert.EqualValues(t, lockOne.getMinimumLockUnit(),
		5, "Minimum lock unit to be in percentage of pretrade price")

	lockOne.TryLockPrice(22)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")

	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Not redeamable if cuurent buy price is decreases and has not made a rebound by at least minimum lock")

	lockOne.TryLockPrice(15)
	assert.True(t, lockOne.IsRedemptionCandidate(), "Lock should be redeamable if the next buy price drops increases by minimum lock amount but still profitable")
	lockOne.TryLockPrice(10)
	assert.False(t, lockOne.IsRedemptionCandidate(), "Lock should not be readeamable if buy pices is readeamable but next minimum price increase returns it back into the dips")

	assert.Equal(t, lockTwo.PretradePrice(), float64(16))
	assert.Equal(t, lockTwo.GetTradeLimit(), float64(5), "<Lock2> stop limit to be set")
	assert.EqualValues(t, lockTwo.getMinimumLockUnit(), 0.32, "<Lock2> Minimum lock unit to be in percentage of  pretrade price")
	lockTwo.TryLockPrice(11)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "<Lock2> not redeamable if current price is greater than stoplimit")

	lockTwo.TryLockPrice(3)
	lockTwo.TryLockPrice(4)
	assert.True(t, lockTwo.IsRedemptionCandidate(), "<Lock2> redeamable if current price reduces by minLimit but greater than stopLimit")

	lockOne.TryLockPrice(19)
	lockOne.TryLockPrice(13)
	assert.False(t, lockOne.IsRedemptionCandidate(), "should not be redeamable if it has the highest increase but redemption is not due")

	assert.True(t, lockTwo.IsRedemptionDue(), "should be due for redemption after minimum increase as a single lock")
	lockOne.TryLockPrice(5)
	lockOne.TryLockPrice(10)
	assert.False(t, lockTwo.IsRedemptionCandidate(), "should not be redeamable if redemption is due but is not the highest lock")

	assert.True(t, lockOne == tradeLocker.BestMatureLock())
	assert.False(t, lockTwo == tradeLocker.BestMatureLock())

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

	lockThree := tradeLocker.AddLock(cfg2, 247).(*highLock)
	lockThree.TryLockPrice(246)
	assert.False(t, lockThree.IsRedemptionDue())
}

func TestTradeLockerBuyAndSell(t *testing.T) {

	tradeLocker := NewLockManager(HighLockCreator)

	buy := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 50,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}
	sell := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 150,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}

	buyConfig := tradeLocker.AddLock(buy, 100).(*highLock)
	sellConfig := tradeLocker.AddLock(sell, 100).(*highLock)

	assert.False(t, buyConfig.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")

	buyConfig.TryLockPrice(39)
	buyConfig.TryLockPrice(40)
	sellConfig.TryLockPrice(152)
	sellConfig.TryLockPrice(151)
	assert.True(t, buyConfig.IsRedemptionCandidate(), "buy should be the redemption candidate from having maximum change")
	assert.False(t, sellConfig.IsRedemptionCandidate(), "buy not be the redemption candidate if there a candidate with higher increase")
	assert.True(t, sellConfig.IsRedemptionDue(), "buy be redeemable even if it is not the best candidate")

	sellConfig.TryLockPrice(180)
	sellConfig.TryLockPrice(179)

	assert.False(t, buyConfig.IsRedemptionCandidate(), "buy should be the redemption candidate from having maximum change")
	assert.True(t, sellConfig.IsRedemptionCandidate(), "buy not be the redemption candidate if there a candidate with higher increase")
	assert.True(t, buyConfig.IsRedemptionDue(), "buy be redeemable even if it is not the best candidate")

}

func TestTradeLockerPriority(t *testing.T) {
	buy := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 50,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}
	sell := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 101,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}

	tradeLockManager := NewLockManager(HighLockCreator)
	tradeLockManager.SetPrioritySide(names.TradeSideSell)

	buyConfig := tradeLockManager.AddLock(buy, 100).(*highLock)
	sellConfig := tradeLockManager.AddLock(sell, 100).(*highLock)

	assert.False(t, buyConfig.IsRedemptionCandidate(), "Not redeamable if cuurent price is greater than buy stopLimit")
	buyConfig.TryLockPrice(10)
	buyConfig.TryLockPrice(11)
	sellConfig.TryLockPrice(110)
	sellConfig.TryLockPrice(109)
	assert.False(t, buyConfig.IsRedemptionCandidate(), "should not be the redemption candidate even with maximum change if it is not priority side")
	assert.True(t, sellConfig.IsRedemptionCandidate(), "should be the redemption candidate even with least change if it priority side")
	assert.True(t, sellConfig.IsRedemptionDue(), "buy be redeemable even if it is not the best candidate")
	assert.True(t, buyConfig.AbsoluteGrowthPercent() > sellConfig.AbsoluteGrowthPercent())

	manager := NewLockManager(HighLockCreator)
	manager.SetPrioritySide(names.TradeSideBuy)
	sellNew := names.TradeConfig{
		Symbol: "BTCUSD",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 100,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}
	sellOne := manager.AddLock(sell, 100).(*highLock)
	sellTwo := manager.AddLock(sellNew, 99).(*highLock)

	sellOne.TryLockPrice(150)
	sellOne.TryLockPrice(149)
	sellTwo.TryLockPrice(200)
	sellTwo.TryLockPrice(199)
	assert.True(t, sellTwo.IsRedemptionCandidate(), "should be the redemption candidate in absense of priority side")
}

func TestTradeLockerGrowth(t *testing.T) {

	buy := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideBuy,
		Buy: names.SideConfig{
			RateLimit: 50,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}

	tradeLockManager := NewLockManager(HighLockCreator)
	buyConfig := tradeLockManager.AddLock(buy, 100).(*highLock)
	buyConfig.TryLockPrice(91)
	assert.EqualValues(t, buyConfig.RelativeGrowthPercent(), -9,"")
	buyConfig.TryLockPrice(109)
	assert.EqualValues(t, buyConfig.RelativeGrowthPercent(), 9,"")



	sell := names.TradeConfig{
		Symbol: "ETHUSDC",
		Side:   names.TradeSideSell,
		Sell: names.SideConfig{
			RateLimit: 101,
			RateType:  names.RateFixed,
			Quantity:  2,
			LockDelta: 1,
		},
	}

	sellLockManager := NewLockManager(HighLockCreator)
	sellConfig := sellLockManager.AddLock(sell, 100).(*highLock)

	sellConfig.TryLockPrice(120)
	assert.EqualValues(t, sellConfig.RelativeGrowthPercent(), 20,"")
	sellConfig.TryLockPrice(80)
	assert.EqualValues(t, sellConfig.RelativeGrowthPercent(), -20,"")
}

