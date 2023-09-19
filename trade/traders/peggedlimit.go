package traders

import (
	// "fmt"
	// "strconv"
	// "sync"
	// "math"
	"trading/binance"
	"trading/helper"

	// "trading/helper"
	"trading/names"
	"trading/trade/manager"
	"trading/user"
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
func NewPeggedLimit(configs []names.TradeConfig) *manager.TradeManager {
	// account := user.GetAccount()
	// symbolList := names.TradeConfigs(configs).ListSymbol()
	// fees := binance.GetTradeFees(symbolList)

	// takersFees := make(map[string]float64)
	// for symbol, fee := range fees {
	// 	takersFees[symbol] = fee.TakerCommission
	// }

	// spotPrices, err := binance.GetSymbolPrices(symbolList)
	// if err != nil {
	// 	panic("ERROR>>>>")
	// }

	// updatedConfig, _ := calculateConfigpeggedLimit(configs, spotPrices, takersFees, account)
	// fmt.Println("Account Takers fee=", helper.Stringify(updatedConfig[0]))

	updatedConfig := getUpdateWithPeggedLimit(configs)
	return NewLimitTrade(updatedConfig)
}

func getUpdateWithPeggedLimit(configs []names.TradeConfig) []names.TradeConfig {
	account := user.GetAccount()
	symbolList := names.TradeConfigs(configs).ListSymbol()
	fees := binance.GetTradeFees(symbolList)

	takersFees := make(map[string]float64)
	for symbol, fee := range fees {
		takersFees[symbol] = fee.TakerCommission
	}

	spotPrices, err := binance.GetSymbolPrices(symbolList)
	if err != nil {
		panic("ERROR>>>>")
	}
	updatedConfig, _ := calculateConfigPeggedLimit(configs, spotPrices, takersFees, account)
	return updatedConfig
}

// Function updates the stop price of the trading config based on a specified peggedcoin limit.
// Ensures the stop price aligns with the peggedcoin limit, allowing realization of the limit value from trading the balance.
// Takes into account available trading asset balance, current quote price, commission percentage, and peggedcoin limit.
// Returns the updated trade configuration.

func calculateConfigPeggedLimit(configs []names.TradeConfig, spotPriceList map[string]float64, takersFees map[string]float64, account user.Accounter) ([]names.TradeConfig, map[string]map[names.TradeSide]float64) {
	updatedConfigs := []names.TradeConfig{}
	configsIncrease := make(map[string]map[names.TradeSide]float64)
	sides := []names.TradeSide{names.TradeSideBuy, names.TradeSideSell}

	for _, config := range configs {
		var updatedConfig = config
		for _, side := range sides {
			symbol := updatedConfig.Symbol
			spotPrice := spotPriceList[symbol.String()]
			fiatLimit := getFiatLimit(spotPrice, updatedConfig, side)

			fiatLimit.SetAccount(account).SetCommision(takersFees[symbol.String()])
			fiatLimit.ApplySyncDeviation()

			updatedConfig = fiatLimit.GetUpdatedConfig()

			if _, ok := configsIncrease[symbol.String()]; !ok {
				configsIncrease[symbol.String()] = make(map[names.TradeSide]float64)
			}
			configsIncrease[symbol.String()][side] = fiatLimit.RaisedQuotLimitWithFees()

			// updatedConfig = fiatLimit.GetUpdatedConfig()
		}
		updatedConfigs = append(updatedConfigs, updatedConfig)
	}
	return updatedConfigs, configsIncrease
}

// // Function updates the stop price of the trading config based on a specified peggedcoin limit.
// // Ensures the stop price aligns with the peggedcoin limit, allowing realization of the limit value from trading the balance.
// // Takes into account available trading as
// func calculateConfigPeggedLimit2(configs []names.TradeConfig, spotPriceList map[string]float64, takersFees map[string]float64, account user.Accounter) ([]names.TradeConfig, map[string]map[names.TradeSide]float64) {
// 	updatedConfigs := []names.TradeConfig{}
// 	configsIncrease := make(map[string]map[names.TradeSide]float64)
// 	sides := []names.TradeSide{names.TradeSideBuy, names.TradeSideSell}

// 	for _, config := range configs {

// 		for _, side := range sides {
// 			sideConfig := config.Buy
// 			symbol := config.Symbol

// 			if side == names.TradeSideSell {
// 				sideConfig = config.Sell
// 			}

// 			tradedBaseQuantity := sideConfig.Quantity
// 			spotPrice := spotPriceList[symbol.String()]

// 			//cost of one spot value in base value
// 			spotToBaseRatio := 1 / spotPrice

// 			if tradedBaseQuantity < 0 {
// 				tradedAsset := symbol.ParseTradingPair().Base
// 				tradedBaseQuantity = account.GetBalance(tradedAsset).Locked

// 				if config.Side == names.TradeSideBuy {
// 					tradedAsset = symbol.ParseTradingPair().Quote
// 					tradedBaseQuantity = account.GetBalance(tradedAsset).Locked / spotPrice
// 				}
// 			}
// 			commission := takersFees[symbol.String()]
// 			// assetSpotValue := tradedBaseQuantity * spotPrice
// 			covertLimitInBase(spotPrice, sideConfig.StopLimit)
// 			requiredBaseRaise := spotToBaseRatio * sideConfig.StopLimit
// 			updatedBaseStopValue := (requiredBaseRaise + tradedBaseQuantity)
// 			updatedQuoteStopValue := updatedBaseStopValue * spotPrice
// 			tradingFees := commission * updatedBaseStopValue // (assetSpotValue /*+ math.Abs(spotPrice - finalRaiseStop) */) //tradedAssetVolume
// 			updatedQuoteWithFees := updatedQuoteStopValue + tradingFees + (tradingFees * spotPrice)
// 			sideConfig.LimitType = names.RatePercent
// 			sideConfig.StopLimit = helper.GrowthPercent((updatedQuoteWithFees / tradedBaseQuantity), spotPrice)
// 			rateLimitDelta := (sideConfig.StopLimit / 100) * sideConfig.LockDelta
// 			sideConfig.LockDelta = rateLimitDelta
// 			if _, ok := configsIncrease[symbol.String()]; !ok {
// 				configsIncrease[symbol.String()] = make(map[names.TradeSide]float64)
// 			}

// 			fiatLimit := getFiatLimit(spotPrice, config, side)
// 			fiatLimit.SetAccount(account).SetCommision(takersFees[symbol.String()])
// 			fiatLimit.GetUpdatedConfig()

// 			configsIncrease[symbol.String()][side] = updatedQuoteWithFees
// 			if side == names.TradeSideBuy {
// 				config.Buy = sideConfig
// 			} else {
// 				config.Sell = sideConfig
// 			}
// 		}
// 		updatedConfigs = append(updatedConfigs, config)
// 	}
// 	return updatedConfigs, configsIncrease
// }

type fiatLimit struct {
	spotPrice  float64
	config     names.TradeConfig
	configSide names.TradeSide
	commision  float64
	account    user.Accounter
}

func getFiatLimit(spotPrice float64, config names.TradeConfig, configSide names.TradeSide) fiatLimit {
	return fiatLimit{spotPrice: spotPrice, config: config, configSide: configSide}
}

func (fLimit *fiatLimit) SetCommision(commision float64) *fiatLimit {
	fLimit.commision = commision
	return fLimit
}

func (fLimit *fiatLimit) SetAccount(account user.Accounter) *fiatLimit {
	fLimit.account = account
	return fLimit
}

// Convert this configside stopLimit to equivalent of the base value.
// it assumes that the limit is a value of the same magnitude as the spot value
func (fLimit *fiatLimit) LimitInBase() float64 {
	spotToBaseRatio := 1 / fLimit.spotPrice
	stopLimit := fLimit.getSideConfig().StopLimit
	requiredBaseRaise := spotToBaseRatio * stopLimit
	return requiredBaseRaise
}

// Convert the stopLimit to equivalent of the base value.
// it assumes that the limit is a value of the same magnitude as the spot value
func (fLimit *fiatLimit) LimitToBase(stopLimit float64) float64 {
	spotToBaseRatio := 1 / fLimit.spotPrice
	requiredBaseRaise := spotToBaseRatio * stopLimit
	return requiredBaseRaise
}

func (fLimit *fiatLimit) getSideConfig() names.SideConfig {
	if helper.SideIsValid(fLimit.configSide) {
		if fLimit.configSide == names.TradeSideSell {
			return fLimit.config.Sell
		}
		return fLimit.config.Buy
	}

	defaultSide := fLimit.config.Side
	if helper.SideIsValid(defaultSide) {
		if defaultSide == names.TradeSideBuy {
			return fLimit.config.Buy
		}
		return fLimit.config.Sell
	}
	return names.SideConfig{}
}

func (fLimit *fiatLimit) GetBaseQuantity(account user.Accounter) float64 {
	quantity := fLimit.getSideConfig().Quantity
	symbol := fLimit.config.Symbol
	config := fLimit.config
	if quantity < 0 {
		tradedAsset := symbol.ParseTradingPair().Base
		quantity = account.GetBalance(tradedAsset).Locked

		if config.Side == names.TradeSideBuy {
			tradedAsset = symbol.ParseTradingPair().Quote
			quantity = account.GetBalance(tradedAsset).Locked / fLimit.spotPrice
		}
	}
	return quantity
}

// Get the value of Base including the Limit note that limit
// is converted to the base value and added to the base value
func (fLimit *fiatLimit) RaisedBaseLimit() float64 {
	return fLimit.LimitInBase() + fLimit.GetBaseQuantity(fLimit.account)
}

// Get the value of Quotes including the Limit
func (fLimit *fiatLimit) RaisedQuoteLimit() float64 {
	baseQuantity := fLimit.RaisedBaseLimit()
	return baseQuantity * fLimit.spotPrice
}

func (fLimit *fiatLimit) totalFeesInBase() float64 {
	return fLimit.commision * fLimit.RaisedBaseLimit()
}

// Get the value of Quotes including the Limit
func (fLimit *fiatLimit) RaisedQuotLimitWithFees() float64 {
	baseFees := fLimit.totalFeesInBase()
	raisedQuoteLimit := fLimit.RaisedQuoteLimit()
	return raisedQuoteLimit + baseFees + (baseFees * fLimit.spotPrice)
}

// Get the value of Quotes including the Limit
func (fLimit *fiatLimit) GetStopLimit() float64 {
	updatedQuote := fLimit.RaisedQuotLimitWithFees()
	baseQuantity := fLimit.GetBaseQuantity(fLimit.account)
	if fLimit.getSideConfig().LimitType == names.RateFixed {
		return updatedQuote
	}
	return helper.GrowthPercent((updatedQuote / baseQuantity), fLimit.spotPrice)
}

// Get the value of Quotes including the Limit
func (fLimit *fiatLimit) GetUpdatedConfig() names.TradeConfig {
	sideConfig := fLimit.getSideConfig()
	sideConfig.LimitType = names.RatePercent
	sideConfig.StopLimit = fLimit.GetStopLimit()
	sideConfig.LockDelta = (sideConfig.StopLimit / 100) * sideConfig.LockDelta
	config := fLimit.config
	if fLimit.configSide == names.TradeSideBuy {
		config.Buy = sideConfig
	} else {
		config.Sell = sideConfig
	}
	return config
}

// Get the value of Quotes including the Limit
func (fLimit *fiatLimit) PercentOfStopLimitOnSpot(stopLimitPercent float64) float64 {
	LimitSideConfig := fLimit.getSideConfig()
	stopLimit := LimitSideConfig.StopLimit
	if stopLimitPercent == 0 || stopLimit == 0 {
		return 0
	}
	limitConfig := fLimit.config
	//Will get stopLimitPercent (percentage) of stopLimit on spotPrice
	LimitSideConfig.StopLimit = (stopLimit / 100) * stopLimitPercent
	LimitSideConfig.LimitType = names.RatePercent

	if fLimit.configSide == names.TradeSideBuy {
		limitConfig.Buy = LimitSideConfig
	} else {
		limitConfig.Sell = LimitSideConfig
	}
	percentFiatLimit := getFiatLimit(fLimit.spotPrice, limitConfig, fLimit.configSide)
	percentStopLimit := percentFiatLimit.GetStopLimit()

	return percentStopLimit
}

// Get the value of Quotes including the Limit
// func (fLimit *fiatLimit) ApplySyncDeviation() float64 {
// 	deviation := fLimit.getSideConfig().DeviationSync
// 	if deviation.Delta == 0 {
// 		return 0
// 	}
// 	// deviationConfig := fLimit.config
// 	// important lets set the stop Limit to the deviation delta so we can calculate
// 	// deviation as a percentage of Limit on the spotValue instead of a percentage of spotValue
// 	side := fLimit.configSide
// 	//failing because of account 232
// 	deviationDelta := fLimit.PercentOfStopLimitOnSpot(deviation.Delta)

// 	// lets return back to this current config and update deviation
// 	sideConfig := fLimit.getSideConfig()
// 	sideConfig.DeviationSync.Delta = deviationDelta
// 	configUpdate := fLimit.config

// 	if side == names.TradeSideBuy {
// 		configUpdate.Buy = sideConfig
// 	} else {
// 		configUpdate.Sell = sideConfig
// 	}
// 	fLimit.config = configUpdate
// 	return deviationDelta
// }

func (fLimit *fiatLimit) ApplySyncDeviation() float64 {
	deviation := fLimit.getSideConfig().DeviationSync
	if deviation.Delta == 0 {
		return 0
	}
	sideConfig := fLimit.getSideConfig()

	// Deviation is calculated as a percentage of the Fiat value on the spot price. 
    // E.g what percentage of spot price is deviation delta
	deviationLimit := helper.GetUnitPercentageOfPrice(fLimit.spotPrice, deviation.Delta) //(stopLimit/100) * deviation.Delta
	side := fLimit.configSide
	sideConfig.DeviationSync.Delta = deviationLimit

	configUpdate := fLimit.config

	if side == names.TradeSideBuy {
		configUpdate.Buy = sideConfig
	} else {
		configUpdate.Sell = sideConfig
	}
	fLimit.config = configUpdate
	return deviationLimit
}

// Convert the limit to equivalent of the base value.
// it assumes that the limit is a value of the same magnitude as the spot value
func covertLimitInBase(spotPrice, stopLimit float64) float64 {
	spotToBaseRatio := 1 / spotPrice
	requiredBaseRaise := spotToBaseRatio * stopLimit
	return requiredBaseRaise
}
