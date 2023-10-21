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
	locks       map[names.Symbol]names.LockInterface
	wrLock      sync.RWMutex
	lockCreator names.LockCreatorFunc

	// Priority side when set will be used to prioritise the trade side to be executed
	// what this means is that if any config on the side of the priority side is due
	// regardless of if it is the highest lock, any other config on the opposing side
	// of the priority will not be executed
	prioritySide names.TradeSide
}

// NewLockManager creates a new TradeLocker instance.
func NewLockManager(lockCreator names.LockCreatorFunc) names.LockManagerInterface {
	return &LockManager{
		locks:       make(map[names.Symbol]names.LockInterface),
		wrLock:      sync.RWMutex{},
		lockCreator: lockCreator,
	}
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetPrioritySide(prioritySide names.TradeSide) {
	m.prioritySide = prioritySide
}

func (m *LockManager) RetrieveLock(config names.TradeConfig) names.LockInterface {
	return m.locks[config.Symbol]
}

// Remove this lock from the manager
func (m *LockManager) RemoveLock(lock names.LockInterface) bool {
	var exist bool
	for _, l := range m.locks {
		if l == lock {
			exist = true
			break
		}
	}
	if exist {
		delete(m.locks, lock.GetLockState().TradeConfig.Symbol)
	}
	return exist
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetLockCreator(creator names.LockCreatorFunc) {
	m.lockCreator = creator
}

func (m *LockManager) BestMatureLock() names.LockInterface {
	var topSellDrift, topBuyDrift float64
	highest := make(map[names.TradeSide]names.LockInterface)

	m.wrLock.RLock()
	for _, lock := range m.locks {
		absoluteChange := lock.AbsoluteGrowthPercent()
		side := lock.TradeSide()
		if lock.IsRedemptionDue() {
			if side.IsSell() && absoluteChange >= topSellDrift {
				topSellDrift = absoluteChange
				highest[lock.TradeSide()] = lock
			} else if side.IsBuy() && absoluteChange >= topBuyDrift {
				//assumes buy low fix this
				topBuyDrift = absoluteChange
				highest[lock.TradeSide()] = lock
			}
		}
	}
	m.wrLock.RUnlock()

	if !m.prioritySide.IsEmpty() {
		priority, exist := highest[m.prioritySide]
		if exist {
			return priority
		}
		// We could not find lock on the priorty side, let fall back
		// to the less priority side in case there is a mature due lock
		side := helper.SwitchTradeSide(m.prioritySide)
		return highest[side]
	}
	buyLock, sellLock := highest[names.TradeSideBuy], highest[names.TradeSideSell]
	if sellLock == nil {
		return buyLock
	}
	if buyLock == nil {
		return sellLock
	}
	if sellLock.AbsoluteGrowthPercent() > buyLock.AbsoluteGrowthPercent() {
		return sellLock
	}
	return buyLock
}

func validateLock(config names.TradeConfig, initialPrice float64) {
	side := config.Side
	priceLimit := helper.CalculateTradePrice(config, initialPrice).Limit
	if side.IsSell() && priceLimit <= initialPrice {
		utils.LogWarn(
			fmt.Sprintf(
				"\nThe Configuration of %s is setup for losses\n"+
					"Initial Price is greater than the stop profit limit for a %s\n"+
					"Initial Price = %f, Profit Limit = %f",
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

	if !helper.SideIsValid(side) {
		utils.LogError(fmt.Errorf("no configuration tradeside"),
			fmt.Sprintf(
				"\nconfiguration of %s is setup without a side\n"+
					"%s\n",
				config.Symbol.String(),
				helper.Stringify(config),
			))
	}
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
	l.wrLock.Lock()
	l.locks[config.Symbol] = newLock
	l.wrLock.Unlock()

	return newLock
}

func tradePricePercentChange(config names.TradeConfig, price, pretradePrice float64) float64 {
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
