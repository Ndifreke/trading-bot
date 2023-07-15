package locker

import (
	"fmt"
	"math"
	"sync"
	"trading/helper"
	"trading/names"
	"trading/utils"
)

// type LockCreatorFunc func(price float64, tradeConfig names.TradeConfig, redemptionIsMature bool, pretradePrice float64, lockManager *LockManager, gainsAccrude float64) names.LockInterface

// LockManager represents a collection of trade locks.
type LockManager struct {
	locks        map[names.Symbol]names.LockInterface
	writeLock    sync.RWMutex
	lockCreator  names.LockCreatorFunc
	prioritySide names.TradeSide
}

// NewLockManager creates a new TradeLocker instance.
func NewLockManager(lockCreator names.LockCreatorFunc) names.LockManagerInterface {
	return &LockManager{
		locks:       make(map[names.Symbol]names.LockInterface),
		writeLock:   sync.RWMutex{},
		lockCreator: lockCreator,
	}
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetPrioritySide(prioritySide names.TradeSide) {
	m.prioritySide = prioritySide
}

// Set the function that will be used to create a new kind of lock
// the lock creator must return a pointer to that lock
func (m *LockManager) SetLockCreator(creator names.LockCreatorFunc) {
	m.lockCreator = creator
}

func (m *LockManager) BestMatureLock() names.LockInterface {
	var topSellDrift, topBuyDrift float64
	highest := make(map[names.TradeSide]names.LockInterface)

	for _, lock := range m.locks {
		absoluteChange := lock.AbsoluteGrowthPercent()
		side := lock.TradeSide()

		if lock.IsRedemptionDue() {
			m.writeLock.Lock()
			if side.IsSell() && absoluteChange >= topSellDrift {
				topSellDrift = absoluteChange
				highest[lock.TradeSide()] = lock
			} else if side.IsBuy() && absoluteChange >= topBuyDrift {
				//assumes buy low fix this
				topBuyDrift = absoluteChange
				highest[lock.TradeSide()] = lock
			}
			m.writeLock.Unlock()
		}
	}
	if m.prioritySide != "" {
		priority, exist := highest[m.prioritySide]
		if exist {
			return priority
		}
		//We could not find lock on the priorty side, let fall back
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

// AddLock adds a new lock to the trade locker.
func (l *LockManager) AddLock(config names.TradeConfig, initialPrice float64) names.LockInterface {

	validateLock(config, initialPrice)
	newLock := l.lockCreator(initialPrice, config, false, initialPrice, l, initialPrice)
	l.writeLock.Lock()
	l.locks[config.Symbol] = newLock
	l.writeLock.Unlock()

	return newLock
}

func getAbsoluteChangePercent(config names.TradeConfig, price, pretradePrice float64) float64 {
	return math.Abs(helper.GrowthPercent(price, pretradePrice))
}

func logLock(lock names.LockInterface) {
	state := lock.GetLockState()
	log := fmt.Sprintf(
		"\n== Locking %f %s %s ==\n"+
			"Locked Gains          : %f\n"+
			"Stop Loss             : %f\n"+
			"Pretrade Price        : %f\n"+
			"Price                 : %f\n"+
			"Redemption Due        : %t\n"+
			"Redemption Candidate  : %t\n"+
			"Minimum Lock Unit     : %f\n"+
			"Price change          : %f\n",
		state.Price,
		state.TradeConfig.Side.String(),
		state.TradeConfig.Symbol,
		state.AccrudGains,
		state.StopLoss,
		state.PretradePrice,
		state.Price,
		state.IsRedemptionIsDue,
		state.IsRedemptionCandidate,
		state.MinimumLockUnit,
		state.AbsoluteGrowth,
	)
	utils.LogInfo(log)
}
