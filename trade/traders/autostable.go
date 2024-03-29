package traders

import (
	// "fmt"
	"fmt"
	"sync"
	"time"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/trade/deviation"
	"trading/trade/graph"
	"trading/trade/manager"
	"trading/utils"

	// "trading/utils"

	"github.com/google/uuid"
)

//splits the balance of the trade between all the assets so that every asset takes
// an equal percentage

type autoStable struct {
	initParams       StableTradeParam
	tradingConfigs   names.TradeConfigs
	executorFunc     names.ExecutorFunc
	tradeLockManager names.LockManagerInterface
	bestSide         names.TradeSide
	status           status
	fullfillId       string
	broadcast        *stream.Broadcaster
	contentionTime   sync.Map
	mutex            sync.Mutex
}

func createAutoStable(initParams StableTradeParam, tradingConfigs names.TradeConfigs, staus status, bestSide names.TradeSide) *autoStable {
	trader := &autoStable{
		fullfillId:     "",
		status:         staus,
		initParams:     initParams,
		tradingConfigs: tradingConfigs,
		broadcast:      stream.NewBroadcast(uuid.New().String()),
		bestSide:       bestSide,
		contentionTime: sync.Map{},
		mutex:          sync.Mutex{},
	}
	return trader
}

func (t *autoStable) Run() {
	for _, tc := range t.tradingConfigs {
		if isValidPeggedBestSideConfig(tc) {
			continue
		}
		go t.Watch(tc)
	}
}

func (t *autoStable) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (tm *autoStable) RemoveConfig(config names.TradeConfig) bool {
	tm.mutex.Lock()
	defer tm.mutex.Unlock()
	var removed bool
	updatedConfigs := []names.TradeConfig{}
	for _, tc := range tm.tradingConfigs {
		if tc.Id == config.Id {
			removed = tm.broadcast.Unsubscribe(config)
		} else {
			updatedConfigs = append(updatedConfigs, tc)
		}
	}
	tm.tradingConfigs = updatedConfigs
	if removed {
		lock := tm.tradeLockManager.RetrieveLock(config)
		if lock != nil {
			lock.RemoveFromManager()
		}
		tm.contentionTime.Delete(config.Id)
	}
	return removed
}

// Add a new config to start watching. If this config exist already
// it will be replaced by the added config and the channel and lock assocated with
// them will also be removed
func (t *autoStable) AddConfig(config names.TradeConfig) {
	t.mutex.Lock()
	t.tradingConfigs = append(t.tradingConfigs, config)
	t.mutex.Unlock()

	go t.Watch(config)
}

func (tm *autoStable) UstradeTrend(trend graph.TrendType) *autoStable {
	panic("Unsupported action")
}

func (tm *autoStable) setConfigContentionTime(config names.TradeConfig) {
	tm.contentionTime.Store(config.Id, time.Now())
}

// renit this contention if the time is elapsed a contention can only run for
// 15 minute then we have to destroy it and reinitialise it
func (tm *autoStable) isContentionTimeUp(config names.TradeConfig) bool {

	// we need to be sure it is not full and fullfil status
	if value, exist := tm.contentionTime.Load(config.Id); exist && config.Id != tm.fullfillId {
		cTime := value.(time.Time)

		//todo we need to be sure this config does not exist already
		if cTime.Minute()%16 < 14 {
			//15 minute has not elapsed since this contention started
			return false
		}

		generatedConfigs := GenerateStableTradeConfigs(tm.initParams)
		newConfig := config

		for _, gc := range generatedConfigs {
			isGcTrading := true
			for _, tc := range tm.tradingConfigs {
				//Lets try to see if there is another config that is not this config or any of the current trading configs
				isGcTrading = tc.Symbol == gc.Symbol
			}
			if !isGcTrading {
				//Found a new config that is not trading exit and use it
				newConfig = gc
				break
			}
		}
		newConfig.Side = config.Side
		newConfig = renitTradeConfig(newConfig, tm.initParams)

		//remove old config and insert new one
		if tm.RemoveConfig(config) {
			tm.AddConfig(newConfig)
			return true
		}
	}
	return false
}

// TODO Rename to small letter done and remove from interface
func (tm *autoStable) Done(tradedConfig names.TradeConfig, locker names.LockInterface) {
	// tm.mutex.Lock()
	// defer tm.mutex.Unlock()

	removed := tm.RemoveConfig(tradedConfig)
	_, exist := tm.tradingConfigs.Find(tradedConfig.Id)

	if removed && exist {
		utils.LogError(fmt.Errorf("removed but still exist tradedConfig config, ensure the config with id %s was called by <names.NewIdTradeConfigs>", tradedConfig.Id), "<autostable>")
		// panic("an error occured, could not find tradedConfig config, ensure the config with id was called by <names.NewIdTradeConfigs>")
	}

	// Generate from init params
	nextSide := helper.SwitchTradeSide(tradedConfig.Side)

	if tm.status == StatusContention && nextSide == tm.bestSide {
		// We need to fulfill this configuration that has completed a contention.
		// So, let's switch it to the best side and redeem it.
		// Generate a new configuration using the initParams blueprint

		// newConfig := initConfig(tradedConfig.Symbol, tm.initParams)
		// newConfig.Side = switchedSide
		// newConfig.Id = tradedConfig.Id

		// if newConfig.Side.IsBuy() {
		// 	newConfig.Buy.Quantity = names.MAX_QUANTITY
		// } else {
		// 	newConfig.Sell.Quantity = names.MAX_QUANTITY
		// }
		// newConfig = getStableTradeConfigs(names.NewIdTradeConfigs(newConfig))[0]

		//
		tradedConfig.Side = nextSide
		if tradedConfig.Side.IsBuy() {
			tradedConfig.Buy.Quantity = names.MAX_QUANTITY
		} else {
			tradedConfig.Sell.Quantity = names.MAX_QUANTITY
		}
		tradedConfig = renitTradeConfig(tradedConfig, tm.initParams)
		//

		tm.status = StatusFullfilment
		// tm.fullfillId = newConfig.Id
		tm.fullfillId = tradedConfig.Id
		tm.AddConfig(tradedConfig)
	} else {
		// we just completed a best side trade, lets generate a new contention config to replace it
		tm.status = StatusContention
		tm.fullfillId = ""

		bestLock := tm.tradeLockManager.BestMatureLock()
		if bestLock != nil && bestLock.IsRedemptionDue() {
			// Check if any of the contention is due then do nothing,
			// if there is a due lock it will most likely be traded into best side
			// Hopefully the executor will handle it execution
			return
		}

		// We could not find any lock that was in due state
		// lets terminate all of them and start a new process
		tm.broadcast.TerminateBroadCast()
		tm.broadcast = stream.NewBroadcast(uuid.New().String())
		tm.tradeLockManager.RemoveLocks()

		newConfigs := GenerateStableTradeConfigs(tm.initParams)
		updatedTradeConfigs := []names.TradeConfig{}

		for _, cfg := range newConfigs {
			idConfig := names.NewIdTradeConfigs(cfg)[0]
			idConfig.Side = nextSide
			updatedTradeConfigs = append(updatedTradeConfigs, idConfig)
		}

		tm.tradingConfigs = getStableTradeConfigs(updatedTradeConfigs)
		tm.Run()
	}
}

func (t *autoStable) SetLockManager(lockMan names.LockManagerInterface) names.Trader {
	t.tradeLockManager = lockMan
	return t
}

func (tm *autoStable) Watch(config names.TradeConfig) {

	tm.setConfigContentionTime(config)
	executor := tm.executorFunc
	subscription := tm.broadcast.Subscribe(config)
	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := tm.tradeLockManager.AddLock(config, pretradePrice)

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()

		if tm.status == StatusFullfilment {
			//supress executing none fullfill trader when status is full
			if tm.fullfillId == state.TradeConfig.Id {
				executor(
					state.TradeConfig,
					state.Price,
					state.PretradePrice,
					func() {
						tm.Done(state.TradeConfig, configLocker)
					},
				)
			}
		} else {
			executor(
				state.TradeConfig,
				state.Price,
				state.PretradePrice,
				func() {
					tm.Done(state.TradeConfig, configLocker)
				},
			)
		}
	})

	deviation := deviation.NewDeviationManager(tm, configLocker)

	deviation.PreAddConfig(func(config names.TradeConfig) names.TradeConfig {
		return renitTradeConfig(config, tm.initParams)
	})

	for sub := range subscription.GetChannel() {

		if tm.isContentionTimeUp(config) {
			continue
		}

		if tm.status == StatusFullfilment && tm.fullfillId != config.Id {
			// during fullfilment don't allow none fullfuling config to deviate else
			// if balance changes by fullfuling config the resulting config from deviation
			// will get into a state that is not wanted
		} else {
			deviation.CheckDeviation(&subscription)
		}

		configLocker.TryLockPrice(sub.Price)

		if tm.status == StatusFullfilment && tm.fullfillId != subscription.State().TradingConfig.Id {
			configLocker.SetVerbose(false)
		} else {
			configLocker.SetVerbose(true)
		}

	}
}

// CuncurrentTrades
func NewAutoStableTrader(initParams StableTradeParam) *manager.TradeManager {
	bestSide := initParams.BestSide

	if bestSide == "" {
		bestSide = names.TradeSideSell
	}

	tradeConfigs := names.NewIdTradeConfigs(GenerateStableTradeConfigs(initParams)...)

	if len(tradeConfigs) == 0 {
		return &manager.TradeManager{}
	}

	tradeConfigs, _ = configsSideToContention(tradeConfigs, bestSide, names.TradeConfig{}, StatusContention)
	tradeConfigs = getStableTradeConfigs(tradeConfigs)
	trader := createAutoStable(initParams, tradeConfigs, StatusContention, bestSide)
	return manager.NewTradeManager(trader)
}

func NewAutoStableExample(run bool) {
	tradeParam := generateStableParams(160, "USDT")
	if run {
		NewAutoStableTrader(tradeParam).DoTrade()
	}
}
