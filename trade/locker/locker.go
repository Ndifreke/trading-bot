package locker

import (
	"fmt"
	"math"
	"os"
	"trading/helper"
	"trading/names"
	"trading/utils"
)

// Lock represents a trade lock.
type Lock struct {
	stopLoss                    float64 // Accepted level that price has to go above or below to sell or buy this trade respectively. The locker may decide not to sell at this point if the price keeps going in the positive direction of the sell action, and vice versa.
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
	locks map[names.Symbol]*Lock
}

// NewTradeLocker creates a new TradeLocker instance.
func NewTradeLocker() names.TradeLockerInterface {
	return &TradeLocker{
		locks: make(map[names.Symbol]*Lock),
	}
}

// validateLock checks if a lock configuration is valid.
func validateLock(config names.TradeConfig, initialPrice, stopLoss float64) {
	side := config.Side
	if side.IsSell() && stopLoss <= initialPrice {
		utils.LogWarn(
			fmt.Sprintf(
				`
		The Configuration of %s is setup for losses
		Initial Price is greater than the stop profit 
		limit for a %s.
		Initial Price = %f, Profit Limit = %f`,
				config.Symbol.String(),
				side.String(),
				initialPrice, stopLoss,
			))
	} else if side.IsBuy() && stopLoss >= initialPrice {
		utils.LogWarn(fmt.Sprintf(
			"\nThe Configuration of %s is setup for losses\n"+
				"Initial Price is less than the stop profit limit for a %s.\n"+
				"Initial Price = %f, Profit Limit = %f",
			config.Symbol.String(),
			side.String(),
			initialPrice,
			stopLoss,
		))
	}
	if stopLoss == 0 {
		utils.LogError(fmt.Errorf("stop Loss for %s %s is zero", side.String(), config.Symbol.String()), "Locker")
		os.Exit(1)
	}
}

// AddLock adds a new lock to the trade locker.
func (l *TradeLocker) AddLock(config names.TradeConfig, initialPrice, stopLoss float64) names.LockInterface {
	validateLock(config, initialPrice, stopLoss)
	lock := Lock{stopLoss: stopLoss, price: initialPrice, tradeConfig: config, redemptionIsDue: false, pretradePrice: initialPrice, lockOwner: l, gainsAccrude: initialPrice}
	l.locks[config.Symbol] = &lock
	return &lock
}

// GetMostProfitableLock returns the most profitable lock for each trade side (buy and sell).
func (locker *TradeLocker) GetMostProfitableLock() map[names.TradeSide]names.LockInterface {

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
	return lock.stopLoss
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
		StopLoss:                    lock.stopLoss,
		LockOwner:                   lock.lockOwner,
		AccrudGains:                 lock.gainsAccrude,
		TradeConfig:                 lock.tradeConfig,
		PretradePrice:               lock.pretradePrice,
		Price:                       lock.price,
		RedemptionIsDue:             lock.IsRedemptionDue(),
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
	return helper.GetPercentChange(locker.price, locker.pretradePrice)
	// (locker.price - locker.basePrice) / locker.basePrice * 100
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
		return lock.stopLoss * (lock.tradeConfig.Price.Buy.LockDelta / 100)
	}
	return lock.stopLoss * (lock.tradeConfig.Price.Sell.LockDelta / 100)
}

// TryLockPrice attempts to lock the price. A price will only lock if it is greater or less than minimum gain
func (lock *Lock) TryLockPrice(price float64) {
	lock.price = price
	config := lock.tradeConfig

	priceChange := price - lock.gainsAccrude
	minimumLock := lock.getMinimumLockUnit()

	// update accrud gains everytime price change is greater or less than minimum change
	if math.Abs(priceChange) >= minimumLock {
		lock.gainsAccrude = price
		gainIncreasedByMinimumLock := priceChange >= minimumLock
		if config.Side.IsSell() {
			// If gains are reduced by at least the minimum lock, and the current price is higher than the stop loss,
			// we sell to avoid further decline in price. There is no need to hold on as this may be a signal of a downturn.
			gainLostPriceIsProfitable := (price > lock.stopLoss) && !gainIncreasedByMinimumLock
			lock.redemptionIsDue = gainLostPriceIsProfitable //priceChange <= -minimumLock
		} else if config.Side.IsBuy() {
			// If gains are increased by at least the minimum lock, and the current price is lower than the stop loss,
			// we initiate redemption as it indicates potential profit taking opportunity.
			lock.redemptionIsDue = (price < lock.stopLoss) && gainIncreasedByMinimumLock
		}
	}

	if lock.redemptionDueCallback != nil && lock.IsRedemptionDue() {
		lock.redemptionDueCallback(lock)
	}

	if lock.redemptionCandidateCallback != nil && lock.IsRedemptionCandidate() {
		lock.redemptionCandidateCallback(lock)
	}

	state := lock.GetLockState()
	log := fmt.Sprintf(
		"\n== Locking %f for %s ==\n"+
			"Locked Gains        : %f\n"+
			"Stop Loss           : %f\n"+
			"Pretrade Price      : %f\n"+
			"Price               : %f\n"+
			"Due                 : %t\n"+
			"Minimum Lock Unit   : %f\n",
		price, lock.tradeConfig.Symbol,
		state.AccrudGains,
		state.StopLoss,
		state.PretradePrice,
		price,
		state.RedemptionIsDue,
		lock.getMinimumLockUnit(),
	)
	utils.LogInfo(log)
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
	return (lock.gainsAccrude - lock.stopLoss) / lock.getMinimumLockUnit()
}
