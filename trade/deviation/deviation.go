package deviation

// CheckDeviation monitors a configuration for any deviations from the latest updates.
// If a configuration goes out of sync by the defined deviation, CheckDeviation will re-watch the config.
// By design, CheckDeviation should be non-disruptive to other configurations, especially if there are due locks.
// To achieve this, we ensure that the traders' RemoveConfig and AddConfig methods do not destroy the state of other configurations.

import (
	"fmt"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/utils"
)

type DeviationManager struct {
	trader      names.Trader
	tradeLock   names.LockInterface
	postAddFunc func(config names.TradeConfig) names.TradeConfig
}

func NewDeviationManager(trader names.Trader, configLocker names.LockInterface) *DeviationManager {
	return &DeviationManager{
		trader:    trader,
		tradeLock: configLocker,
	}
}

// If provided this function will called with the config
// before it is inserted back again into the trade runner
func (dev *DeviationManager) PreAddConfig(postAddFunc func(config names.TradeConfig) names.TradeConfig) {
	dev.postAddFunc = postAddFunc
}

func (dev *DeviationManager) CheckDeviation(subscription *stream.Subscription) {

	originalConfig := subscription.State().TradingConfig
	config := originalConfig
	var deviation names.DeviationSync

	if config.Side.IsBuy() && config.Buy.DeviationSync.Delta > 0 {
		deviation = config.Buy.DeviationSync
	}
	if config.Side.IsSell() && config.Sell.DeviationSync.Delta > 0 {
		deviation = config.Sell.DeviationSync
	}

	if deviation.Delta == 0 {
		// No deviation is defined for this config
		// we should implement global deviation
		// that determines what side to apply deviation to
		// and if this side matches it, we should allow deviation to run
		// but only on the accepted side
		return
	}

	if !helper.SideIsValid(config.Side) {
		utils.LogError(fmt.Errorf("invalid configuration side"), fmt.Sprintf("deviation Failed because of invalid side %s", config.Side))
	}

	spotPrice, pretradePrice := dev.tradeLock.GetLockState().Price, dev.tradeLock.GetLockState().PretradePrice
	deviationPrice := GetDeviationTriggerPrice(pretradePrice, config)

	if config.Side.IsBuy() && spotPrice >= deviationPrice {
		if deviation.FlipSide {
			config.Side = names.TradeSideSell
			if config.Sell.StopLimit == 0 {
				//TODO
				// there is no config on the side we are switching to
				// lets use the template config provided
				// build up config either from graph or template that the user
				// will provde for this side
			}
		}
		if dev.trader.RemoveConfig(originalConfig) {
			//TODO might want to use gorutine for postAdd and AddConfig
			if dev.postAddFunc != nil {
				config = dev.postAddFunc(config)
			}
			dev.trader.AddConfig(config)
		}
	}

	if config.Side.IsSell() && spotPrice <= deviationPrice {
		if deviation.FlipSide {
			config.Side = names.TradeSideBuy
			if config.Buy.StopLimit == 0 {
				//TODO
				// there is no config on the side we are switching to
				// lets use the template config provided
				// build up config either from graph or template that the user
				// will provde for this side
			}
		}
		if dev.trader.RemoveConfig(originalConfig) {
			if dev.postAddFunc != nil {
				config = dev.postAddFunc(config)
			}
			dev.trader.AddConfig(config)
		}
	}
}

func GetDeviationTriggerPrice(pretradePrice float64, config names.TradeConfig) float64 {
	side := config.Side
	deviationSpotLimit := 0.0
	if side.IsBuy() && config.Buy.DeviationSync.Delta != 0 {
		devationValue := helper.CalculateValueOfPercentage(pretradePrice, config.Buy.DeviationSync.Delta)
		// In Buy if the price increase to this point
		deviationSpotLimit = pretradePrice + devationValue
	}
	if side.IsSell() && config.Sell.DeviationSync.Delta != 0 {
		devationValue := helper.CalculateValueOfPercentage(pretradePrice, config.Sell.DeviationSync.Delta)
		// In Buy if the price increase to this point
		deviationSpotLimit = pretradePrice - devationValue
	}
	return deviationSpotLimit
}
