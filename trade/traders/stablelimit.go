package traders

import (
	"trading/names"
	"trading/trade/manager"
)

func NewStableLimit(configs []names.TradeConfig) *manager.TradeManager {

	updatedConfig := getStableTradeConfigs(configs)
	return NewLimitTrade(updatedConfig)
}

func NewStableLimitExample(run bool) {
	config := initConfig("MEMEUSDT", generateStableParams(160, "USDT"))
	if run {
		NewStableLimit(names.NewIdTradeConfigs(config)).DoTrade()
	}
}
