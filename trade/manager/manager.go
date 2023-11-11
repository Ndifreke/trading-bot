package manager

import (
	"fmt"
	"trading/helper"
	"trading/names"
	"trading/trade/executor"
	"trading/trade/locker"
	"trading/utils"
)

type TradeManager struct {
	trader       names.Trader
	prioritySide names.TradeSide
	lockCreator  names.LockCreatorFunc
}

func NewTradeManager(trader names.Trader) *TradeManager {
	return &TradeManager{
		trader:       trader,
		lockCreator:  locker.PeakHighLockCreator,
		prioritySide: names.TradeSideSell,
	}
}

// set the prioritySide that should by the lock manager to decide which tradeside
// to attempt to sell first when mature default names.TradeSideSell
func (tm *TradeManager) UsePriority(prioritySide names.TradeSide) *TradeManager {
	tm.prioritySide = prioritySide
	return tm
}

// set the lock creator that should be used to create a locker by the lock manager default is locker.CreateYieldHighLock
func (tm *TradeManager) UseLockCreator(lockCreator names.LockCreatorFunc) *TradeManager {
	tm.lockCreator = lockCreator
	return tm
}

func (tm *TradeManager) DoTrade() *TradeManager {

	if tm.trader == nil {
		utils.LogWarn("no trader registered with trade manager")
		return tm
	}

	lockManager := locker.NewLockManager(tm.lockCreator)
	if tm.prioritySide != "" {
		if !helper.SideIsValid(tm.prioritySide) {
			utils.LogError(fmt.Errorf("invalid priority side"), string(tm.prioritySide))
		}
		lockManager.SetPrioritySide(tm.prioritySide)
	}
	utils.LogInfo(fmt.Sprintf(
		"\n=== Trade Manager Summary === \n"+
			"Priority Side     :%s\n",
		tm.prioritySide,
	))
	tm.trader.
		SetLockManager(lockManager).
		SetExecutor(tm.Execute).
		Run()
	return tm
}

func (tm *TradeManager) Execute(
	config names.TradeConfig,
	spot float64,
	basePrice float64,
	done func()) {
	var sold bool

	if config.Side.IsBuy() {
		sold = executor.BuyExecutor(config, spot, basePrice).Execute()
	} else {
		sold = executor.SellExecutor(config, spot, basePrice).Execute()
	}
	if !sold {
		return
	}
	done()
}
