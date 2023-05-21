package locker

import (
	"trading/trade"
)

type lock struct {
	lossLimit                   float32 // limit that price should not go below
	price                       float32 //only lock when it is upto a percent lock
	basePrice                   float32 //starting price
	gainLimit                   float32 //current gains accrude
	tradeConfig                 trade.TradeConfig
	redemptionIsDue             bool
	lockOwner                   *TradeLocker
	redemptionDueCallback       func(lock)
	redemptionCandidateCallback func(lock)
}

type TradeLocker struct {
	locks map[trade.Symbol]*lock
}

func NewTradeLocker() *TradeLocker {
	return &TradeLocker{
		locks: make(map[trade.Symbol]*lock),
	}
}

func (l *TradeLocker) AddLock(config trade.TradeConfig, initialPrice, stopLimit float32) *lock {
	lock := lock{lossLimit: stopLimit, price: initialPrice, tradeConfig: config, redemptionIsDue: false, basePrice: initialPrice, lockOwner: l, gainLimit: initialPrice}
	l.locks[config.Symbol] = &lock
	return &lock
}

func (locker *TradeLocker) GetMostProfitableLock() map[trade.TradeAction]*lock {

	var trackedSellIncrease, trackedBuyIncrease float32
	highest := make(map[trade.TradeAction]*lock)

	for _, lck := range locker.locks {
		lockIncrease := lck.GetPercentIncrease()
		action := lck.tradeConfig.Action

		if action.IsSell() && lockIncrease > trackedSellIncrease {
			trackedSellIncrease = lockIncrease
			highest[lck.tradeConfig.Action] = lck
		} else if action.IsBuy() && lockIncrease < trackedBuyIncrease {
			//assumes buy low fix this
			trackedBuyIncrease = lockIncrease
			highest[lck.tradeConfig.Action] = lck
		}
	}
	return highest
}

func (lock *lock) GetLossLimit() float32 {
	return lock.lossLimit
}
func (lock *lock) GetBasePrice() float32 {
	return lock.basePrice
}
func (lock *lock) GetLockedPrice() float32 {
	return lock.gainLimit
}
func (lock *lock) GetLockOwner() *TradeLocker {
	return lock.lockOwner
}

func (locker *lock) GetPercentIncrease() float32 {
	return (locker.price - locker.basePrice) / locker.basePrice * 100
}
func (l *lock) isHighestLockAction() bool {
	return l.GetLockOwner().GetMostProfitableLock()[l.tradeConfig.Action] == l
}

// Get the minimum amount that the current price has to change in order to Lock in as profit
// it value of delta calculated as a percentage of stopLimit
func (lock *lock) getMinimumLockUnit() float32 {
	//LockUnit is derived as the percentage of lockDelta on stopLoseLimt
	return lock.lossLimit * (lock.tradeConfig.Price.Sell.LockDelta / 100)
}

func (lock *lock) TryLockPrice(price float32) {
	lock.price = price
	minimumLock := lock.getMinimumLockUnit()

	priceLimitChange := (price - lock.gainLimit) / minimumLock

	if priceLimitChange >= 1 || priceLimitChange <= -1 {
		lock.gainLimit = price
		lock.redemptionIsDue = priceLimitChange <= -1
	}

	//notify listener on this lock when redemption is due
	if lock.redemptionDueCallback != nil && lock.IsRedemptionDue() {
		lock.redemptionDueCallback(*lock)
	}

	//notify listener on this lock when this is a redemption candidate
	if lock.redemptionCandidateCallback != nil && lock.IsRedemptionCandidate() {
		lock.redemptionCandidateCallback(*lock)
	}
}

// determines if it is the most profitable lock from other locks of similar action
func (lock *lock) IsRedemptionCandidate() bool {
	return lock.isHighestLockAction() && lock.IsRedemptionDue()
}

func (lock *lock) IsRedemptionDue() bool {
	// does not ake into consideration the action
	// only redeam on the merit of profitability
	currentPrice := lock.price
	redemptionIsDue := lock.redemptionIsDue

	if lock.lossLimit > currentPrice {
		return false
	}

	if lock.tradeConfig.Action.IsBuy() {
		return false
	}
	return redemptionIsDue
}

func (l *lock) SetRedemptionDueCallback(cb func(lock)) {
	l.redemptionDueCallback = cb
}

func (l *lock) SetRedemptionCandidateCallback(cb func(lock)) {
	l.redemptionCandidateCallback = cb
}

// Total amount of units of price delta that has been locked in
func (lock lock) getLockedUnits(config trade.TradeConfig) float32 {
	return (lock.gainLimit - lock.lossLimit) / lock.getMinimumLockUnit()
}
