package locker

import (
	"math"
	"trading/helper"
	"trading/names"
)

// Allows only one MatureLock which is the higest Lock
// if priority is supplied the the manager, the priority side
// will be preffered in case of maturity before other side

type highLock struct {
	price                     float64 // Only lock when it is up to a percent lock.
	pretradePrice             float64 // Starting price.
	gainsAccrude              float64 // Current gains accrued.
	tradeConfig               names.TradeConfig
	redemptionIsMature        bool
	lockManager               names.LockManagerInterface
	maturityCallback          func(names.LockInterface)
	maturityCandidateCallback func(names.LockInterface)
}

var HighLockCreator = func(price float64, tradeConfig names.TradeConfig, redemptionIsMature bool, pretradePrice float64, lockManager names.LockManagerInterface, gainsAccrude float64) names.LockInterface {
	return &highLock{
		price:              price,
		tradeConfig:        tradeConfig,
		redemptionIsMature: redemptionIsMature,
		pretradePrice:      pretradePrice,
		lockManager:        lockManager,
		gainsAccrude:       gainsAccrude,
	}
}

// GetTradeLimit returns the stop loss limit for the lock.
func (lock *highLock) RemoveFromManager() bool {
	return lock.lockManager.RemoveLock(lock)
}

// GetTradeLimit returns the stop loss limit for the lock.
func (lock *highLock) GetTradeLimit() float64 {
	return helper.CalculateTradePrice(lock.tradeConfig, lock.pretradePrice).Limit
}

// PretradePrice returns the pre-trade price for the lock.
func (lock *highLock) PretradePrice() float64 {
	return lock.pretradePrice
}

// GetLockedPrice returns the current locked price for the lock.
func (lock *highLock) GetLockedPrice() float64 {
	return lock.gainsAccrude
}

// GetLockState returns the current state of the lock.
func (lock *highLock) GetLockState() names.LockState {
	return names.LockState{
		StopLoss:                    lock.GetTradeLimit(),
		LockOwner:                   lock.lockManager,
		AccrudGains:                 lock.gainsAccrude,
		TradeConfig:                 lock.tradeConfig,
		PretradePrice:               lock.pretradePrice,
		Price:                       lock.price,
		IsRedemptionIsDue:           lock.IsRedemptionDue(),
		IsRedemptionCandidate:       lock.IsRedemptionCandidate(),
		RedemptionDueCallback:       lock.maturityCallback,
		RedemptionCandidateCallback: lock.maturityCandidateCallback,
		MinimumLockUnit:             lock.getMinimumLockUnit(),
		AbsoluteGrowth:              lock.AbsoluteGrowthPercent(),
	}
}

// GetLockManager returns the owner of the lock.
func (lock *highLock) GetLockManager() names.LockManagerInterface {
	return lock.lockManager
}

// AbsoluteGrowthPercent calculates the percentage by which the current price has deviated from the pre-trade price.
// A positive value means the current locked price has gained.
func (lock highLock) AbsoluteGrowthPercent() float64 {
	percentChange := tradePricePercentChange(lock.tradeConfig, lock.price, lock.pretradePrice)
	return math.Abs(percentChange)
}

func (lock highLock) RelativeGrowthPercent() float64 {
	return tradePricePercentChange(lock.tradeConfig, lock.price, lock.pretradePrice)
}

// getMinimumLockUnit calculates the minimum amount that the current price
// needs to change in order to lock in a profit. It calculates the value of
// LockUnit as a percentage of stopLossLimit.
func (lock *highLock) getMinimumLockUnit() float64 {
	// LockUnit is derived as the product of stopLossLimit and the percentage
	// represented by lockDelta
	if lock.tradeConfig.Side == names.TradeSideBuy {
		return lock.pretradePrice * (lock.tradeConfig.Buy.LockDelta / 100)
	}
	return lock.pretradePrice * (lock.tradeConfig.Sell.LockDelta / 100)
}

// TryLockPrice attempts to lock the price. A price will only lock if it is greater or less than minimum gain
func (lock *highLock) TryLockPrice(price float64) {
	lock.price = price
	config := lock.tradeConfig

	priceChange := price - lock.gainsAccrude
	minimumLock := lock.getMinimumLockUnit()
	stopLoss := lock.GetTradeLimit()

	// // update accrud gains everytime price change is greater or less than minimum change
	if math.Abs(priceChange) >= minimumLock {
		lock.gainsAccrude = price
		gainIncreasedByMinimumLock := priceChange >= minimumLock
		if config.Side.IsSell() {

			// If gains are reduced by at least the minimum lock, and the current price is higher than the stop loss,
			// we sell to avoid further decline in price. There is no need to hold on as this may be a signal of a downturn.
			gainLostPriceIsProfitable := (price > stopLoss) && !gainIncreasedByMinimumLock
			lock.redemptionIsMature = gainLostPriceIsProfitable //priceChange <= -minimumLock

		} else if config.Side.IsBuy() {

			// If gains are increased by at least the minimum lock, and the current price is lower than the stop loss,
			// we initiate redemption as it indicates potential profit taking opportunity.
			lock.redemptionIsMature = (price < stopLoss) && gainIncreasedByMinimumLock
		}
	}

	logLock(lock)

	if lock.maturityCallback != nil && lock.IsRedemptionDue() {
		lock.maturityCallback(lock)
	}

	if lock.maturityCandidateCallback != nil && lock.IsRedemptionCandidate() {
		lock.maturityCandidateCallback(lock)
	}
}

// best Mature Lock attempts to retrieve the best lock from the list managed
// by lock manager. The sentinels are based on is it maturity, is it prioritySide, is it highest percentage increase
func (l *highLock) bestMatureLock() names.LockInterface {
	return l.GetLockManager().BestMatureLock() //3

}

// determines if it is the most profitable lock from other locks of similar action
func (lock *highLock) IsRedemptionCandidate() bool {
	return lock.bestMatureLock() == lock //2
}

// Checks if this lock is mature on it own
func (lock *highLock) IsRedemptionDue() bool {
	return lock.redemptionIsMature
}

func (l *highLock) SetRedemptionDueCallback(cb func(lock names.LockInterface)) { //maturecallback
	l.maturityCallback = cb
}

func (l *highLock) SetRedemptionCandidateCallback(cb func(lock names.LockInterface)) {
	l.maturityCandidateCallback = cb
}

// Total amount of units of price delta that has been locked in
func (lock highLock) getLockedUnits(config names.TradeConfig) float64 {
	return (lock.gainsAccrude - lock.GetTradeLimit()) / lock.getMinimumLockUnit()
}

func (lock highLock) TradeSide() names.TradeSide {
	return lock.tradeConfig.Side
}
