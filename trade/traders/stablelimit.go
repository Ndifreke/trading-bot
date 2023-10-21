package traders

import (
	// "fmt"
	// "strconv"
	// "sync"
	// "math"

	// "trading/helper"
	"trading/names"
	"trading/trade/manager"
	// "trading/utils"
)

// RateLimit will considered as USDT equivalent if the spot price
// is 250USDT given the available balance or quantity of asset to trade, the rateLimit
// will be used to determine how much 250 has to grow to realise this USDT limit including fees incured

// LockDelta is treated as percentage of RateLimit if rate Limit = 5USDT, Spot = 250 and delta = 10
//  RateLimitDelta = (RateLimit/100) * LockDelta = 5USDT/100 * 10 = 0.5
// minimum lock unit (Spot/100) * RateLimitDelta = 250/100

// When you are buying, you use the Quote asset to buy the base asset,
// We you sell, you are selling the Base asset into the Quote Value
// BTCUSDT = Sell = BTC => USDT, BUY = USDT => BTC
// All trades asset quantity are completed in BASE asset
func NewStableLimit(configs []names.TradeConfig) *manager.TradeManager {

	updatedConfig := getStableTradeConfigs(configs)
	return NewLimitTrade(updatedConfig)
}
