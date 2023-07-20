package names

const (
	RatePercent   RateType  = "PERCENT"
	RateFixed     RateType  = "FIXED"
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type RateType string

func (pt RateType) IsPercent() bool {
	return pt == RatePercent
}

func (pt RateType) IsFixed() bool {
	return pt == RateFixed
}

type TradeSide string

func (ta TradeSide) IsBuy() bool {
	return ta == TradeSideBuy
}

func (ta TradeSide) IsSell() bool {
	return ta == TradeSideSell
}

func (ta TradeSide) String() string {
	return string(ta)
}

type SideConfigDeviation struct {
	Threshold  float64 //percent from pretrade price to initiate deviation action
	SwitchSide bool    // switch side when deviation condition is met
}
type SideConfig struct {
	// Determins the position where trade should be executed. Trade may not actually execute here for a number of reasons
	// if locker is still in positive or the fee prevents profit
	RateLimit  float64
	RateType   RateType //PERCENT, FIXED_VALUE
	Quantity   float64
	MustProfit bool
	// determines what percentage change in price to lock positive price movement
	LockDelta float64
	Deviation SideConfigDeviation
}

type TradeConfig struct {
	Sell          SideConfig
	Buy           SideConfig
	Side          TradeSide //BUY or SELL
	StopCondition bool      // a complex condition expression here
	Symbol        Symbol
	IsCyclick     bool // Will run both sell and buy after each other is completed
}

type LockState struct {
	StopLoss                    float64 // limit that price should not go below
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

type LowerHighTrendConfig struct {
}

type LowerLowTrendConfig struct {
}
