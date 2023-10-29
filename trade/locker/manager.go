package locker

import (
	"fmt"
	"sync"
	"trading/helper"
	"trading/names"
	"trading/trade/deviation"
	"trading/utils"
)

// LockManager represents a collection of trade locks.
type LockManager struct {
	locks       sync.Map
	lockCreator names.LockCreatorFunc
	prioritySide names.TradeSide
}

// NewLockManager creates a new TradeLocker instance.
func NewLockManager(lockCreator names.LockCreatorFunc) names.LockManagerInterface {
	return &LockManager{
		locks:       sync.Map{},
		lockCreator: lockCreator,
	}
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetPrioritySide(prioritySide names.TradeSide) {
	m.prioritySide = prioritySide
}

func (m *LockManager) RetrieveLock(config names.TradeConfig) names.LockInterface {
	lock, _ := m.locks.Load(config.Symbol)
	if lock != nil {
		return lock.(names.LockInterface)
	}
	return nil
}

// Retrieve all locks from this lock manager
func (m *LockManager) RetrieveLocks() map[names.Symbol]names.LockInterface {
	locks := make(map[names.Symbol]names.LockInterface)
	m.locks.Range(func(key, value interface{}) bool {
		locks[key.(names.Symbol)] = value.(names.LockInterface)
		return true
	})
	return locks
}

// Remove this lock from the manager
func (m *LockManager) RemoveLock(lock names.LockInterface) bool {
	var deleted bool
	m.locks.Range(func(key, value interface{}) bool {
		if value == lock {
			m.locks.Delete(key)
			deleted = true
		}
		return true
	})
	return deleted
}

// Remove All locks on this lock manager
func (m *LockManager) RemoveLocks() bool {
	m.locks = sync.Map{}
	return true
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetLockCreator(creator names.LockCreatorFunc) {
	m.lockCreator = creator
}

func (m *LockManager) BestMatureLock() names.LockInterface {
	var topSellDrift, topBuyDrift float64
	var highestLockSell, highestLockBuy names.LockInterface

	m.locks.Range(func(key, value interface{}) bool {
		lock := value.(names.LockInterface)
		absoluteChange := lock.AbsoluteGrowthPercent()
		side := lock.TradeSide()
		if lock.IsRedemptionDue() {
			if side.IsSell() && absoluteChange >= topSellDrift {
				topSellDrift = absoluteChange
				highestLockSell = lock
			} else if side.IsBuy() && absoluteChange >= topBuyDrift {
				topBuyDrift = absoluteChange
				highestLockBuy = lock
			}
		}
		return true
	})

	if !m.prioritySide.IsEmpty() {
		if m.prioritySide.IsSell() && highestLockSell != nil {
			return highestLockSell
		} else if m.prioritySide.IsBuy() && highestLockBuy != nil {
			return highestLockBuy
		}
	}

	if highestLockSell != nil && highestLockBuy != nil {
		if highestLockSell.AbsoluteGrowthPercent() > highestLockBuy.AbsoluteGrowthPercent() {
			return highestLockSell
		}
		return highestLockBuy
	}

	if highestLockSell != nil {
		return highestLockSell
	}
	return highestLockBuy
}

func validateLock(config names.TradeConfig, initialPrice float64) {
	// Validation logic for locks
}

func validateConfig(config names.TradeConfig) bool {
	validate := func(cfg names.SideConfig, side names.TradeSide) bool {
		if cfg.LimitType != names.RateFixed && cfg.LimitType != names.RatePercent {
			utils.LogError(
				fmt.Errorf("configuration should explicitely specify limit type as percentage or fixed"),
				fmt.Sprintf("but found '%s' on side %s. This will treat limit as absolute value %f",
					cfg.LimitType,
					side,
					cfg.StopLimit),
			)
			return false
		}
		return true
	}
	return validate(config.Buy, names.TradeSideBuy) && validate(config.Sell, names.TradeSideSell)
}

// AddLock adds a new lock to the trade locker.
func (l *LockManager) AddLock(config names.TradeConfig, initialPrice float64) names.LockInterface {
	validateLock(config, initialPrice)
	validateConfig(config)

	newLock := l.lockCreator(initialPrice, config, false, initialPrice, l, initialPrice)
	l.locks.Store(config.Symbol, newLock)
	return newLock
}

func tradePricePercentChange(config names.TradeConfig, price, pretradePrice float64) float64 {
	// Calculate price change
	return helper.CalculatePercentageChange(price, pretradePrice)
}


func logLock(lock names.LockInterface) {
	state := lock.GetLockState()
	config := state.TradeConfig
	symbol := config.Symbol
	pretradePrice := state.PretradePrice
	spotPrice := state.Price
	side := config.Side
	log := fmt.Sprintf(
		"\n== Locking %f %s %s ==\n"+
			"Locked Gains          : %f\n"+
			"Stop Loss             : %f\n"+
			"Pretrade Price        : %f\n"+
			// "Price                 : %f\n"+
			"Redemption Due        : %t\n"+
			"Redemption Candidate  : %t\n"+
			"Minimum Lock Unit     : %f\n"+
			"Price Change          : %f\n"+
			"Gains                 : %s\n",
		spotPrice,
		symbol.String(),
		side.String(),
		state.AccrudGains,
		state.StopLimit,
		pretradePrice,
		// spotPrice,
		state.IsRedemptionIsDue,
		state.IsRedemptionCandidate,
		state.MinimumLockUnit,
		state.AbsoluteGrowth,
		symbol.Gains(pretradePrice, spotPrice, side),
	)
	deviationSpotLimit := deviation.GetDeviationTriggerPrice(pretradePrice, config) //0.0
	// if side == names.TradeSideBuy && config.Buy.DeviationSync.Delta != 0 {
	// 	devationValue := helper.CalculateValueOfPercentage(pretradePrice, config.Buy.DeviationSync.Delta)
	// 	// In Buy if the price increase to this point
	// 	deviationSpotLimit = pretradePrice + devationValue
	// }
	// if side == names.TradeSideSell && config.Sell.DeviationSync.Delta != 0 {
	// 	devationValue := helper.CalculateValueOfPercentage(pretradePrice, config.Sell.DeviationSync.Delta)
	// 	// In Buy if the price increase to this point
	// 	deviationSpotLimit = pretradePrice - devationValue
	// }
	if deviationSpotLimit != 0 {
		log = log + fmt.Sprintf("Deviation Trigger     : %s\n", symbol.FormatQuotePrice(deviationSpotLimit))
	}
	utils.LogInfo(log)
}
