package locker

import (
	"fmt"
	"math"
	"sync"
	"trading/helper"
	"trading/names"
	"trading/utils"
)

// Lock represents a trade lock.
type Lock struct {
	price                       float64 // Only lock when it is up to a percent lock.
	pretradePrice               float64 // Starting price.
	gainsAccrude                float64 // Current gains accrued.
	tradeConfig                 names.TradeConfig
	redemptionIsDue             bool
	lockOwner                   names.TradeLockerInterface
	redemptionDueCallback       func(names.LockInterface)
	redemptionCandidateCallback func(names.LockInterface)
}

// TradeLocker represents a collection of trade locks.
type TradeLocker struct {
	locks     map[names.Symbol]*Lock
	writeLock sync.RWMutex
}

// NewTradeLocker creates a new TradeLocker instance.
func NewTradeLocker() names.TradeLockerInterface {
	return &TradeLocker{
		locks:     make(map[names.Symbol]*Lock),
		writeLock: sync.RWMutex{},
	}
}

// validateLock checks if a lock configuration is valid.
func validateLock(config names.TradeConfig, initialPrice float64) {
	side := config.Side
	priceLimit := helper.CalculateTradePrice(config, initialPrice).Limit
	if side.IsSell() && priceLimit <= initialPrice {
		utils.LogWarn(
			fmt.Sprintf(
				`
		The Configuration of %s is setup for losses
		Initial Price is greater than the stop profit 
		limit for a %s.
		Initial Price = %f, Profit Limit = %f`,
				config.Symbol.String(),
				side.String(),
				initialPrice, priceLimit,
			))
	} else if side.IsBuy() && priceLimit >= initialPrice {
		utils.LogWarn(fmt.Sprintf(
			"\nThe Configuration of %s is setup for losses\n"+
				"Initial Price is less than the stop profit limit for a %s.\n"+
				"Initial Price = %f, Profit Limit = %f",
			config.Symbol.String(),
			side.String(),
			initialPrice,
			priceLimit,
		))
	}
}

// AddLock adds a new lock to the trade locker.
func (l *TradeLocker) AddLock(config names.TradeConfig, initialPrice float64) names.LockInterface {
	l.writeLock.Lock()
	defer l.writeLock.Unlock()

	validateLock(config, initialPrice)
	lock := Lock{price: initialPrice, tradeConfig: config, redemptionIsDue: false, pretradePrice: initialPrice, lockOwner: l, gainsAccrude: initialPrice}
	l.locks[config.Symbol] = &lock
	return &lock
}

// GetMostProfitableLock returns the most profitable lock for each trade side (buy and sell).
func (locker *TradeLocker) GetMostProfitableLock() map[names.TradeSide]names.LockInterface {
	locker.writeLock.Lock()
	defer locker.writeLock.Unlock()

	var trackedSellIncrease, trackedBuyIncrease float64
	highest := make(map[names.TradeSide]names.LockInterface)

	for _, lck := range locker.locks {
		increase := lck.GetPercentIncrease()
		action := lck.tradeConfig.Side

		if action.IsSell() && increase > trackedSellIncrease {
			trackedSellIncrease = increase
			highest[lck.tradeConfig.Side] = lck
		} else if action.IsBuy() && increase < trackedBuyIncrease {
			//assumes buy low fix this
			trackedBuyIncrease = increase
			highest[lck.tradeConfig.Side] = lck
		}
	}

	return highest
}

// GetLossLimit returns the stop loss limit for the lock.
func (lock *Lock) GetLossLimit() float64 {
	return helper.CalculateTradePrice(lock.tradeConfig, lock.pretradePrice).Limit
}

// PretradePrice returns the pre-trade price for the lock.
func (lock *Lock) PretradePrice() float64 {
	return lock.pretradePrice
}

// GetLockedPrice returns the current locked price for the lock.
func (lock *Lock) GetLockedPrice() float64 {
	return lock.gainsAccrude
}

// GetLockState returns the current state of the lock.
func (lock *Lock) GetLockState() names.LockState {
	return names.LockState{
		StopLoss:                    lock.GetLossLimit(),
		LockOwner:                   lock.lockOwner,
		AccrudGains:                 lock.gainsAccrude,
		TradeConfig:                 lock.tradeConfig,
		PretradePrice:               lock.pretradePrice,
		Price:                       lock.price,
		IsRedemptionIsDue:           lock.IsRedemptionDue(),
		IsRedemptionCandidate:       lock.IsRedemptionCandidate(),
		RedemptionDueCallback:       lock.redemptionDueCallback,
		RedemptionCandidateCallback: lock.redemptionCandidateCallback,
	}
}

// GetLockOwner returns the owner of the lock.
func (lock *Lock) GetLockOwner() names.TradeLockerInterface {
	return lock.lockOwner
}

// GetPercentIncrease calculates the percentage by which the current price has deviated from the pre-trade price.
// A positive value means the current locked price has gained.
func (locker Lock) GetPercentIncrease() float64 {
	return helper.GetPercentGrowth(locker.price, locker.pretradePrice)
}

func (l *Lock) isHighestLockAction() bool {
	return l.GetLockOwner().GetMostProfitableLock()[l.tradeConfig.Side] == l
}

// getMinimumLockUnit calculates the minimum amount that the current price
// needs to change in order to lock in a profit. It calculates the value of
// LockUnit as a percentage of stopLossLimit.
func (lock *Lock) getMinimumLockUnit() float64 {
	// LockUnit is derived as the product of stopLossLimit and the percentage
	// represented by lockDelta
	if lock.tradeConfig.Side == names.TradeSideBuy {
		return lock.GetLossLimit() * (lock.tradeConfig.Buy.LockDelta / 100)
	}
	return lock.GetLossLimit() * (lock.tradeConfig.Sell.LockDelta / 100)
}

// TryLockPrice attempts to lock the price. A price will only lock if it is greater or less than minimum gain
func (lock *Lock) TryLockPrice(price float64) {
	lock.price = price
	config := lock.tradeConfig

	priceChange := price - lock.gainsAccrude
	minimumLock := lock.getMinimumLockUnit()
	stopLoss := lock.GetLossLimit()

	// update accrud gains everytime price change is greater or less than minimum change
	if math.Abs(priceChange) >= minimumLock {
		lock.gainsAccrude = price
		gainIncreasedByMinimumLock := priceChange >= minimumLock
		if config.Side.IsSell() {
			// If gains are reduced by at least the minimum lock, and the current price is higher than the stop loss,
			// we sell to avoid further decline in price. There is no need to hold on as this may be a signal of a downturn.
			gainLostPriceIsProfitable := (price > stopLoss) && !gainIncreasedByMinimumLock
			lock.redemptionIsDue = gainLostPriceIsProfitable //priceChange <= -minimumLock
		} else if config.Side.IsBuy() {
			// If gains are increased by at least the minimum lock, and the current price is lower than the stop loss,
			// we initiate redemption as it indicates potential profit taking opportunity.
			lock.redemptionIsDue = (price < stopLoss) && gainIncreasedByMinimumLock
		}
	}
	state := lock.GetLockState()
	log := fmt.Sprintf(
		"\n== Locking %f %s %s ==\n"+
			"Locked Gains          : %f\n"+
			"Stop Loss             : %f\n"+
			"Pretrade Price        : %f\n"+
			"Price                 : %f\n"+
			"Redemption Due        : %t\n"+
			"Redemption Candidate  : %t\n"+
			"Minimum Lock Unit     : %f\n",
		price,
		state.TradeConfig.Side.String(),
		lock.tradeConfig.Symbol,
		state.AccrudGains,
		state.StopLoss,
		state.PretradePrice,
		price,
		state.IsRedemptionIsDue,
		state.IsRedemptionCandidate,
		lock.getMinimumLockUnit(),
	)
	utils.LogInfo(log)

	if lock.redemptionDueCallback != nil && lock.IsRedemptionDue() {
		lock.redemptionDueCallback(lock)
	}

	if lock.redemptionCandidateCallback != nil && lock.IsRedemptionCandidate() {
		lock.redemptionCandidateCallback(lock)
	}

}

// determines if it is the most profitable lock from other locks of similar action
func (lock *Lock) IsRedemptionCandidate() bool {
	return lock.isHighestLockAction() && lock.IsRedemptionDue()
}

func (lock *Lock) IsRedemptionDue() bool {
	return lock.redemptionIsDue
}

func (l *Lock) SetRedemptionDueCallback(cb func(lock names.LockInterface)) {
	l.redemptionDueCallback = cb
}

func (l *Lock) SetRedemptionCandidateCallback(cb func(lock names.LockInterface)) {
	l.redemptionCandidateCallback = cb
}

// Total amount of units of price delta that has been locked in
func (lock Lock) getLockedUnits(config names.TradeConfig) float64 {
	return (lock.gainsAccrude - lock.GetLossLimit()) / lock.getMinimumLockUnit()
}
