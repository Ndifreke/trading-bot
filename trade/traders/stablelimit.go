package traders

import (
	"trading/names"
	"trading/trade/manager"
)

func NewStableLimit(configs []names.TradeConfig) *manager.TradeManager {

	updatedConfig := getStableTradeConfigs(configs)
	return NewLimitTrade(updatedConfig)
}
