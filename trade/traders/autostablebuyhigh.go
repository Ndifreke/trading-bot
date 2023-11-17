package traders

import (
	// "fmt"
	"fmt"
	"log"
	"sync"
	"time"
	"trading/binance"
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

type autoStableBuyHigh struct {
	initParams       StableTradeParam
	tradingConfigs   names.TradeConfigs
	executorFunc     names.ExecutorFunc
	tradeLockManager names.LockManagerInterface
	status           status
	fullfillId       string
	broadcast        *stream.Broadcaster
	contentionTime   sync.Map
	mutex            sync.Mutex
}

func createAutoStableBuyHigh(initParams StableTradeParam, tradingConfigs names.TradeConfigs) *autoStableBuyHigh {
	trader := &autoStableBuyHigh{
		fullfillId:     "",
		status:         StatusContention,
		initParams:     initParams,
		tradingConfigs: tradingConfigs,
		broadcast:      stream.NewBroadcast(uuid.New().String()),
		contentionTime: sync.Map{},
		mutex:          sync.Mutex{},
	}
	return trader
}

func (t *autoStableBuyHigh) Run() {
	for _, tc := range t.tradingConfigs {
		if isValidPeggedBestSideConfig(tc) {
			continue
		}
		go t.Watch(tc)
	}
}

func (t *autoStableBuyHigh) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (tm *autoStableBuyHigh) RemoveConfig(config names.TradeConfig) bool {
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
func (t *autoStableBuyHigh) AddConfig(config names.TradeConfig) {
	t.mutex.Lock()
	t.tradingConfigs = append(t.tradingConfigs, config)
	t.mutex.Unlock()

	go t.Watch(config)
}

func (tm *autoStableBuyHigh) UstradeTrend(trend graph.TrendType) *autoStableBuyHigh {
	panic("Unsupported action")
}

func (tm *autoStableBuyHigh) setConfigContentionTime(config names.TradeConfig) {
	tm.contentionTime.Store(config.Id, time.Now())
}

// renit this contention if the time is elapsed a contention can only run for
// 15 minute then we have to destroy it and reinitialise it
func (tm *autoStableBuyHigh) isContentionTimeUp(config names.TradeConfig) bool {

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
func (tm *autoStableBuyHigh) Done(tradedConfig names.TradeConfig, locker names.LockInterface) {

	cfg, exist := tm.tradingConfigs.Find(tradedConfig.Id)
	removed := tm.RemoveConfig(cfg)

	if exist && !removed {
		utils.LogError(fmt.Errorf("removed but still exist tradedConfig config, ensure the config with id %s was called by <names.NewIdTradeConfigs>", tradedConfig.Id), "<autostable>")
		log.Fatal("removed but still exist tradedConfig config, ensure the config with id was called by <names.NewIdTradeConfigs>", tradedConfig.Id)
	}

	if tm.status == StatusContention && tradedConfig.Side == names.TradeSideBuy {
		// change the side only after the config has been generated, this is to
		// ensure that we reinit the config with the opposite side of the params
		tradedConfig = renitTradeConfig(tradedConfig, tm.initParams)
		tradedConfig.Side = names.TradeSideSell

		// we cannot use the same side during contention for redemption there
		// is a chance that the values may be too high and the trade will not be exuted
		// so we switch the side to other side which we presume will be lower and meant for redemption
		// A stright forward solution will be to explicitly set the side to sell since this is an auto stable high trader
		// tradedConfig.Sell = tradedConfig.Buy
		tradedConfig.Sell.Quantity = names.MAX_QUANTITY

		tm.status = StatusFullfilment
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
			withId := names.NewIdTradeConfigs(cfg)[0]
			withId.Side = names.TradeSideSell
			updatedTradeConfigs = append(updatedTradeConfigs, withId)
		}

		tm.tradingConfigs = getStableTradeConfigs(updatedTradeConfigs)
		tm.Run()
	}
}

func (t *autoStableBuyHigh) SetLockManager(lockMan names.LockManagerInterface) names.Trader {
	t.tradeLockManager = lockMan
	return t
}

func (tm *autoStableBuyHigh) Watch(config names.TradeConfig) {

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
			// switch this config to buy so that executor will buy it even
			// though the contention was set to sell
			tradeConfig := state.TradeConfig
			tradeConfig.Side = names.TradeSideBuy
			executor(
				tradeConfig,
				state.Price,
				state.PretradePrice,
				func() {
					tm.Done(tradeConfig, configLocker)
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

func NewAutoStableBuyHighTrader(initParams StableTradeParam) *manager.TradeManager {

	tradeConfigs := names.NewIdTradeConfigs(GenerateStableTradeConfigs(initParams)...)

	if len(tradeConfigs) == 0 {
		return &manager.TradeManager{}
	}

	tradeConfigs = getStableTradeConfigs(tradeConfigs)

	mapFunc := func(cfg names.TradeConfig) names.TradeConfig {
		cfg.Buy, cfg.Sell = cfg.Sell, cfg.Buy
		return cfg
	}
	tradeConfigs = names.NewIdTradeConfigs(tradeConfigs...).Map(mapFunc)
	trader := createAutoStableBuyHigh(initParams, tradeConfigs)
	return manager.NewTradeManager(trader)
}

func NewAutoStableBuyHighExample(run bool) {
	tradeParam := generateStableParams(160, "USDT")
	tradeParam.Side = names.TradeSideSell
	if run {
		NewAutoStableBuyHighTrader(tradeParam).DoTrade()
	}
}
