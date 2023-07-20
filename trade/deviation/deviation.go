package deviation

import (
	"fmt"
	"math"
	"trading/helper"
	"trading/names"
	"trading/stream"
	"trading/utils"
)

type DeviationManager struct {
	trader    names.Trader
	tradeLock names.LockInterface
}

func NewDeviationManager(trader names.Trader, configLocker names.LockInterface) *DeviationManager {
	return &DeviationManager{
		trader:    trader,
		tradeLock: configLocker,
	}
}

func (dev *DeviationManager) CheckDeviation(subscription *stream.Subscription) {

	originalConfig := subscription.State().TradingConfig
	config := subscription.State().TradingConfig

	if config.Side == names.TradeSideBuy {
		deviation := config.Buy.Deviation
		if deviation.Threshold == 0 {
			return
		}
	}
	if config.Side == names.TradeSideSell {
		deviation := config.Sell.Deviation
		if deviation.Threshold == 0 {
			return
		}
	}

	if !helper.SideIsValid(config.Side) {
		utils.LogError(fmt.Errorf("invalid configuration side"), fmt.Sprintf("deviation Failed because of invalid side %s", config.Side))
	}

	priceChange := dev.tradeLock.AbsoluteGrowthPercent()

	//TODO   REMOVE
	if config.Side == names.TradeSideBuy {
		priceChange = 9
	}
	if config.Side == names.TradeSideSell {
		priceChange = -9
	}

	if config.Side == names.TradeSideBuy {

		deviation := config.Buy.Deviation

		// positive deviation for buy is an indication that the price has grown
		// above the pretrade price, negative value is an indication that the
		// price is decreasing towards the buy position.
		// NOTE: We assume that a buy position will always be below the pretrade price
		if priceChange >= deviation.Threshold {

			if deviation.SwitchSide {
				config.Side = names.TradeSideSell
				if config.Sell.RateLimit == 0 {
					//TODO
					// there is no config on the side we are switching to
					// lets use the template config provided
					// build up config either from graph or template that the user
					// will provde for this side
				}
			}
			if dev.trader.RemoveConfig(originalConfig) {
				dev.trader.AddConfig(config)
			}
		}
	}

	if config.Side == names.TradeSideSell {
		deviation := config.Sell.Deviation

		// negative deviation for sell is an indication that the price has dipped
		// below the pretrade price, positive value is an indication that the
		// price is increasing towards the sell position.
		// NOTE: We assume that a sell position will always be above the pretrade price
		if priceChange < 0  && math.Abs(priceChange) >= deviation.Threshold {
			if deviation.SwitchSide {

				config.Side = names.TradeSideBuy
				if config.Buy.RateLimit == 0 {
					//TODO
					// there is no config on the side we are switching to
					// lets use the template config provided
					// build up config either from graph or template that the user
					// will provde for this side
				}

			}
			fmt.Println(originalConfig.Side, config.Side)
			if dev.trader.RemoveConfig(originalConfig) {
				dev.trader.AddConfig(config)
			}
		}
	}
}
