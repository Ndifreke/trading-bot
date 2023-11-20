package traders

import (
	"fmt"
	"sync"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/trade/deviation"
	"trading/trade/graph"
	"trading/trade/manager"
	"trading/user"
	"trading/utils"

	"github.com/google/uuid"
)

//splits the balance of the trade between all the assets so that every asset takes
// an equal percentage

type autoStableSplit struct {
	initParams         StableTradeParam
	tradingConfigs     names.TradeConfigs
	executorFunc       names.ExecutorFunc
	tradeLockManager   names.LockManagerInterface
	bestSide           names.TradeSide
	contentionChannels []stream.Subscription
	lock               sync.RWMutex
	broadcast          *stream.Broadcaster
}

func createAutoStableSplitTrader(initParams StableTradeParam, tradingConfigs names.TradeConfigs, bestSide names.TradeSide) *autoStableSplit {
	trader := &autoStableSplit{
		initParams:         initParams,
		tradingConfigs:     tradingConfigs,
		broadcast:          stream.NewBroadcast(uuid.New().String()),
		bestSide:           bestSide,
		contentionChannels: []stream.Subscription{},
	}
	return trader
}

func (t *autoStableSplit) Run() {
	for _, tc := range t.tradingConfigs {
		if isValidPeggedBestSideConfig(tc) {
			continue
		}
		go t.Watch(tc)
	}
}

func (t *autoStableSplit) SetExecutor(executorFunc names.ExecutorFunc) names.Trader {
	t.executorFunc = executorFunc
	return t
}

// Remove a config and it associated registeredLocks (subscription and lock)
func (t *autoStableSplit) RemoveConfig(config names.TradeConfig) bool {
	var removed bool
	updatedConfigs := []names.TradeConfig{}
	for _, tc := range t.tradingConfigs {
		if tc.Id == config.Id {
			removed = t.broadcast.Unsubscribe(config)
		} else {
			updatedConfigs = append(updatedConfigs, tc)
		}
	}
	t.tradingConfigs = updatedConfigs
	if removed {
		t.tradeLockManager.RetrieveLock(config).RemoveFromManager()
		}
	return removed
}

// Add a new config to start watching. If this config exist already
// it will be replaced by the added config and the channel and lock assocated with
// them will also be removed
func (t *autoStableSplit) AddConfig(config names.TradeConfig) {
	t.tradingConfigs = append(t.tradingConfigs, config)
	go t.Watch(config)
}

func (tm *autoStableSplit) UstradeTrend(trend graph.TrendType) *autoStableSplit {
	panic("Unsupported action")
}

// TODO Rename to small letter done and remove from interface
func (tm *autoStableSplit) Done(tradedConfig names.TradeConfig, locker names.LockInterface) {
	tm.broadcast.TerminateBroadCast()
	tradedConfig, exist := tm.tradingConfigs.Find(tradedConfig.Id)
	removed := tm.RemoveConfig(tradedConfig)

	if !removed || !exist {
		panic("an error occured, could not find tradedConfig config, ensure the config with id was called by <names.NewIdTradeConfigs>")
	}

	// Generate from init params
	switchedSide := helper.SwitchTradeSide(tradedConfig.Side)

	if switchedSide == tm.bestSide {
		// We need to fullfil this config that has completed a contention
		// So lets switch it to best side and redeem it.
		// Generate a new config using the initParams blueprint and replace it with the traded config
		newConfig := initConfig(tradedConfig.Symbol, tm.initParams)
		newConfig.Side = switchedSide
		newConfig.Id = tradedConfig.Id

		if newConfig.Side.IsBuy() {
			newConfig.Buy.Quantity = names.MAX_QUANTITY
		} else {
			newConfig.Sell.Quantity = names.MAX_QUANTITY
		}
		newConfig = getStableTradeConfigs(names.NewIdTradeConfigs(newConfig))[0]
		tm.AddConfig(newConfig)

	} else {
		// we just completed a best side trade, lets generate a new contention config to replace it
		newConfig := GenerateStableTradeConfigs(tm.initParams)[0]
		newConfig.Side = switchedSide
		contenderCount := 1.0

		for _, cfg := range tm.tradingConfigs {
			if cfg.Side == newConfig.Side {
				contenderCount++
			}
		}
		quoteBalance := user.GetAccount().
			GetBalance(newConfig.Symbol.ParseTradingPair().Quote).Free

		// A contention should not be allow to use all the available balance of the stable quote asset
		// Allocate a portion of stable balance meant for contenters to newConfig
		// in the future we should specify contention allocation either as percentage or fixed value
		allocQty := (quoteBalance / contenderCount)

		if newConfig.Side.IsBuy() {
			newConfig.Buy.Quantity = allocQty
		} else {
			newConfig.Sell.Quantity = allocQty
		}
		// Assign an ID and apply tradeconfigs
		newConfig = getStableTradeConfigs(names.NewIdTradeConfigs(newConfig))[0]
		tm.AddConfig(newConfig)
		return
	}
}

func (t *autoStableSplit) SetLockManager(lockMan names.LockManagerInterface) names.Trader {
	t.tradeLockManager = lockMan
	return t
}

func (trader *autoStableSplit) Watch(config names.TradeConfig) {
	executor := trader.executorFunc
	subscription := trader.broadcast.Subscribe(config)

	pretradePrice := binance.GetPriceLatest(config.Symbol.String())
	configLocker := trader.tradeLockManager.AddLock(config, pretradePrice)

	trader.lock.Lock()
	trader.contentionChannels = append(trader.contentionChannels, subscription)
	trader.lock.Unlock()

	configLocker.SetRedemptionCandidateCallback(func(l names.LockInterface) {
		state := l.GetLockState()
		executor(
			state.TradeConfig,
			state.Price,
			state.PretradePrice,
			func() {
				trader.Done(state.TradeConfig, configLocker)
			},
		)
	})

	deviation := deviation.NewDeviationManager(trader, configLocker)

	deviation.PreAddConfig(func(config names.TradeConfig) names.TradeConfig {
		// search this config from the initConfig. Note the initConfig is a blueprint
		// from which other stableTradeConfig can be created from. It represent the users
		// intent in stable asset and not the percentage or fixed value that can be used by a
		// trade locker
		_, exist := trader.tradingConfigs.Find(config.Id)
		if !exist {
			utils.LogError(fmt.Errorf("attempt to post insert config %s %s failed", config.Symbol, config.Id), "<autoStable_v1>")
			return config
		}
		cfg := initConfig(config.Symbol, trader.initParams)
		cfg.Id = config.Id
		cfg.Side = config.Side
		stableConfig := getStableTradeConfigs(names.NewIdTradeConfigs(cfg))
		return stableConfig[0]
	})

	for sub := range subscription.GetChannel() {
		// if trader.status == StatusContention {

		// Deviation is executed selectively, specifically when the status is in contention.
		// This approach is adopted to prevent potential loss of gains while fulfilling.
		// TODO provide configuration to either enable or disable this behaviour
		go deviation.CheckDeviation(&subscription)
		// }
		configLocker.TryLockPrice(sub.Price)
	}
}

// CuncurrentTrades
func NewAutoStableSplitTrader(params StableTradeParam) *manager.TradeManager {
	bestSide := params.BestSide

	if bestSide == "" {
		bestSide = names.TradeSideSell
	}

	tradeConfig := names.NewIdTradeConfigs(GenerateStableTradeConfigs(params)...)

	if len(tradeConfig) == 0 {
		return &manager.TradeManager{}
	}

	tradeConfig = getStableTradeConfigs(tradeConfig)
	trader := createAutoStableSplitTrader(params, tradeConfig, bestSide)
	return manager.NewTradeManager(trader)
}

func NewAutoStableSplitExample(run bool) {
	tradeParam := StableTradeParam{
		QuoteAsset:         "USDT",
		BuyStopLimit:       0.2,
		BuyDeviationDelta:  1,
		BuyLockDelta:       0.03,
		SellStopLimit:      0.1,
		SellDeviationDelta: 0.5,
		SellLockDelta:      0.03,
		BestSide:           names.TradeSideSell,
		Status:             StatusContention,
		MinPriceChange:     3,
		MaxPriceChange:     11,
	}
	if run {
		NewAutoStableSplitTrader(tradeParam).DoTrade()
	}
}
