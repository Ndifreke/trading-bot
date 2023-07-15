package locker

import (
	"math"
	"trading/helper"
	"trading/names"
)

// Free Lock does not care about who the highest lock is, it immediately reports a lock as due
// When it is mature and has not reduced below the last highest point

// type freeLock Lock

type unbiaseLock struct {
	price                     float64 // Only lock when it is up to a percent lock.
	pretradePrice             float64 // Starting price.
	gainsAccrude              float64 // Current gains accrued.
	tradeConfig               names.TradeConfig
	redemptionIsMature        bool
	lockManager               names.LockManagerInterface
	maturityCallback          func(names.LockInterface)
	maturityCandidateCallback func(names.LockInterface)
}

func UnbiasedLockCreator(price float64, tradeConfig names.TradeConfig, redemptionIsMature bool, pretradePrice float64, lockManager names.LockManagerInterface, gainsAccrude float64) names.LockInterface {
	return &unbiaseLock{
		price:              price,
		tradeConfig:        tradeConfig,
		redemptionIsMature: redemptionIsMature,
		pretradePrice:      pretradePrice,
		lockManager:        lockManager,
		gainsAccrude:       gainsAccrude,
	}
}

// GetTradeLimit returns the stop loss limit for the lock.
func (lock *unbiaseLock) GetTradeLimit() float64 {
	return helper.CalculateTradePrice(lock.tradeConfig, lock.pretradePrice).Limit
}

// PretradePrice returns the pre-trade price for the lock.
func (lock *unbiaseLock) PretradePrice() float64 {
	return lock.pretradePrice
}

// GetLockedPrice returns the current locked price for the lock.
func (lock *unbiaseLock) GetLockedPrice() float64 {
	return lock.gainsAccrude
}

// GetLockState returns the current state of the lock.
func (lock *unbiaseLock) GetLockState() names.LockState {
	return names.LockState{
		StopLoss:                    lock.GetTradeLimit(),
		LockOwner:                   lock.lockManager,
		AccrudGains:                 lock.gainsAccrude,
		TradeConfig:                 lock.tradeConfig,
		PretradePrice:               lock.pretradePrice,
		Price:                       lock.price,
		IsRedemptionIsDue:           lock.IsRedemptionDue(),
		IsRedemptionCandidate:       lock.IsRedemptionDue(),
		RedemptionDueCallback:       lock.maturityCallback,
		RedemptionCandidateCallback: lock.maturityCandidateCallback,
		MinimumLockUnit:             lock.getMinimumLockUnit(),
		AbsoluteGrowth:              lock.AbsoluteGrowthPercent(),
	}
}

// GetLockManager returns the owner of the lock.
func (lock *unbiaseLock) GetLockManager() names.LockManagerInterface {
	return lock.lockManager
}

// AbsoluteGrowthPercent calculates the percentage by which the current price has deviated from the pre-trade price.
// A positive value means the current locked price has gained.
func (lock unbiaseLock) AbsoluteGrowthPercent() float64 {
	return getAbsoluteChangePercent(lock.tradeConfig, lock.price, lock.pretradePrice)
}

// getMinimumLockUnit calculates the minimum amount that the current price
// needs to change in order to lock in a profit. It calculates the value of
// LockUnit as a percentage of stopLossLimit.
func (lock *unbiaseLock) getMinimumLockUnit() float64 {
	// LockUnit is derived as the product of stopLossLimit and the percentage
	// represented by lockDelta
	if lock.tradeConfig.Side == names.TradeSideBuy {
		return lock.pretradePrice * (lock.tradeConfig.Buy.LockDelta / 100)
	}
	return lock.pretradePrice * (lock.tradeConfig.Sell.LockDelta / 100)
}

// TryLockPrice attempts to lock the price. A price will only lock if it is greater or less than minimum gain
func (lock *unbiaseLock) TryLockPrice(price float64) {
	lock.price = price
	config := lock.tradeConfig

	priceChange := price - lock.gainsAccrude
	minimumLock := lock.getMinimumLockUnit()
	stopLoss := lock.GetTradeLimit()

	// update accrud gains everytime price change is greater or less than minimum change
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
	if lock.maturityCallback != nil && lock.IsRedemptionDue() { //LimitReached
		lock.maturityCallback(lock)
	}

	if lock.maturityCandidateCallback != nil && lock.IsRedemptionCandidate() {
		lock.maturityCandidateCallback(lock)
	}
}

// determines if it is the most profitable lock from other locks of similar action
func (lock *unbiaseLock) IsRedemptionCandidate() bool {
	return lock.IsRedemptionDue()
}

func (lock *unbiaseLock) IsRedemptionDue() bool {
	return lock.redemptionIsMature
}

func (l *unbiaseLock) SetRedemptionCandidateCallback(cb func(lock names.LockInterface)) {
	l.maturityCandidateCallback = cb
}

// Total amount of units of price delta that has been locked in
func (lock unbiaseLock) getLockedUnits(config names.TradeConfig) float64 {
	return (lock.gainsAccrude - lock.GetTradeLimit()) / lock.getMinimumLockUnit()
}

func (lock unbiaseLock) TradeSide() names.TradeSide {
	return lock.tradeConfig.Side
}
