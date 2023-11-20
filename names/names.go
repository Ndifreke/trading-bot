package names

import (
	"fmt"

	"github.com/google/uuid"
)

const (
	RatePercent StopLimit = "PERCENT"
	// RateUSDT   RateType  = "USDT"
	RateFixed     StopLimit = "FIXED"
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type QuantityType float64

var MAX_QUANTITY float64 = -1

func (limit QuantityType) IsMax() bool {
	return limit == -1
}

type StopLimit string

func (pt StopLimit) IsPercent() bool {
	return pt == RatePercent
}

func (pt StopLimit) IsFixed() bool {
	return pt == RateFixed
}

type TradeSide string

func (ta TradeSide) IsBuy() bool {
	return ta == TradeSideBuy
}

func (ta TradeSide) IsEmpty() bool {
	return ta == ""
}

func (ta TradeSide) IsSell() bool {
	return ta == TradeSideSell
}

func (ta TradeSide) String() string {
	return string(ta)
}

type DeviationSync struct {
	//percent from pretrade price to initiate deviation action
	Delta float64
	// switch side when deviation condition is met
	FlipSide bool
}

type SideConfig struct {
	// Determins the position where trade should be executed. Trade may not actually execute here for a number of reasons
	// if locker is still in positive or the fee prevents profit
	StopLimit  float64
	LimitType  StopLimit //PERCENT, FIXED_VALUE
	Quantity   float64
	MustProfit bool
	//determines what percentage change in price to lock positive price movement
	LockDelta     float64
	DeviationSync DeviationSync
}

type TradeConfig struct {
	Id            string
	Sell          SideConfig
	Buy           SideConfig
	Side          TradeSide //BUY or SELL
	StopCondition bool      // a complex condition expression here
	Symbol        Symbol
	IsCyclick     bool // Will run both sell and buy after each other is completed
}

type TradeConfigs []TradeConfig

// NewIdTradeConfigs creates a new collection of trade configurations with generated IDs.
func NewIdTradeConfigs(configs ...TradeConfig) TradeConfigs {
	items := []TradeConfig{}
	for i, config := range configs {
		if config.Id == "" {
			config.Id = fmt.Sprintf("%d_%s", i, uuid.New().String())
		}
		items = append(items, config)
	}
	tcg := TradeConfigs(items)
	return tcg
}

func (configs TradeConfigs) ListSymbol() []string {
	var s []string
	for _, config := range configs {
		s = append(s, config.Symbol.String())
	}
	return s
}

func (cfg TradeConfigs) Configs() []TradeConfig {
	return cfg
}

func (cfgs TradeConfigs) Find(id string) (TradeConfig, bool) {
	for _, cfg := range cfgs {
		if cfg.Id != "" && cfg.Id == id {
			return cfg, true
		}
	}
	return TradeConfig{}, false
}

func (cfgs TradeConfigs) Map(callback func(TradeConfig)TradeConfig) TradeConfigs {
	updates := []TradeConfig{}
	for _, cfg := range cfgs {
		updates = append(updates, callback(cfg))
	}
	return updates
}

func (cfgs TradeConfigs) ForEach(callback func(TradeConfig))  {
	for _, cfg := range cfgs {
		callback(cfg)
	}
}

type LockState struct {
	StopLimit                   float64 // limit that price should not go below
	Price                       float64 //only lock when it is upto a percent lock
	PretradePrice               float64 //starting price
	AccrudGains                 float64 //current gains accrude
	TradeConfig                 TradeConfig
	IsRedemptionIsDue           bool
	IsRedemptionCandidate       bool
	LockOwner                   LockManagerInterface
	RedemptionDueCallback       func(LockInterface)
	RedemptionCandidateCallback func(LockInterface)
	MinimumLockUnit             float64
	AbsoluteGrowth              float64
}

type LockInterface interface {
	SetRedemptionCandidateCallback(func(LockInterface))
	GetLockState() LockState
	AbsoluteGrowthPercent() float64
	RelativeGrowthPercent() float64
	TryLockPrice(price float64)
	SetVerbose(verbose bool)
	TradeSide() TradeSide
	IsRedemptionDue() bool
	GetLockManager() LockManagerInterface
	RemoveFromManager() bool
}

// type LockCreatorFunc func(price float64, tradeConfig TradeConfig, redemptionIsMature bool, pretradePrice float64, lockManager  TradeLockManagerInterface, gainsAccrude float64) LockInterface
type LockCreatorFunc func(price float64, tradeConfig TradeConfig, redemptionIsMature bool, pretradePrice float64, lockManager LockManagerInterface, gainsAccrude float64) LockInterface

type LockManagerInterface interface {
	AddLock(config TradeConfig, pretradePrice float64) LockInterface
	BestMatureLock() LockInterface
	SetPrioritySide(prioritySide TradeSide)
	SetLockCreator(LockCreatorFunc)
	RetrieveLock(config TradeConfig) LockInterface
	RemoveLock(lock LockInterface) bool
	RemoveLocks() bool
	RetrieveLocks() map[Symbol]LockInterface
}

type ExecutorFunc = func(
	config TradeConfig,
	currentPrice float64,
	basePrice float64,
	done func(),
)

type Trader interface {
	Run()
	Done(confg TradeConfig, locker LockInterface)
	SetExecutor(tradeExecutor ExecutorFunc) Trader
	SetLockManager(tradeLocker LockManagerInterface) Trader
	AddConfig(confg TradeConfig)
	RemoveConfig(config TradeConfig) bool
}

type TradeManagerFunc func(t ...TradeConfig) *Trader

type TrendConfig struct {
	Support    float64
	Resistance float64
	SpikeCount int
}
