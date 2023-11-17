package traders

// Buy order = Quote asset
// Sell order = Base asset
import (
	"encoding/json"
	"fmt"
	"math"
	"trading/binance"
	"trading/helper"
	"trading/names"
	"trading/user"
	"trading/utils"
)

func isValidPeggedBestSideConfig(tc names.TradeConfig) bool {
	if tc.Side.IsBuy() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Buy)), "Buy Side Configuration Error")
	}
	if tc.Side.IsSell() && false {
		utils.LogError(fmt.Errorf("invalid Side Configuration %s", helper.Stringify(tc.Sell)), "Sell Side Configuration Error")
	}
	return false
}

type stableutil struct {
	spotPrice      float64
	config         names.TradeConfig
	calculatedSide names.TradeSide
	commision      float64
	account        user.AccountInterface
}

func newStable(spotPrice float64, config names.TradeConfig, calculateSide names.TradeSide) stableutil {
	return stableutil{spotPrice: spotPrice, config: config, calculatedSide: calculateSide}
}

func (stable *stableutil) UseCommision(commision float64) *stableutil {
	stable.commision = commision
	return stable
}

func (stable *stableutil) UseAccount(account user.AccountInterface) *stableutil {
	stable.account = account
	return stable
}

// Get the config of the side that is currently being updated,
// this is not the same as the trade side config
func (stable *stableutil) getCalculateSideConfig() names.SideConfig {
	if helper.SideIsValid(stable.calculatedSide) {
		if stable.calculatedSide.IsSell() {
			return stable.config.Sell
		}
		return stable.config.Buy
	}

	defaultSide := stable.config.Side
	if helper.SideIsValid(defaultSide) {
		if defaultSide.IsBuy() {
			return stable.config.Buy
		}
		return stable.config.Sell
	}
	return names.SideConfig{}
}

// Calculate the quantity of the calculated side in base value of the trading pairs
func (stable *stableutil) QuantityInBaseValue(account user.AccountInterface) float64 {
	symbol := stable.config.Symbol
	sideConfig := stable.getCalculateSideConfig()
	quantity := sideConfig.Quantity
	baseAsset := symbol.ParseTradingPair().Base
	side := stable.calculatedSide

	if side.IsSell() && quantity < 0 {
		quantity = account.GetBalance(baseAsset).Free
	}

	if side.IsBuy() {
		// in Buy action the quote quontity is always supplied
		quoteAsset := symbol.ParseTradingPair().Quote
		if quantity < 0 {
			quantity = account.GetBalance(quoteAsset).Free
		}
		quantity = (quantity / stable.spotPrice)
	}

	return quantity
}

// Calculate the quote asset value in terms of quantity
func (stable *stableutil) QuantityInQuoteValue(account user.AccountInterface) float64 {
	 baseQty := stable.QuantityInBaseValue(stable.account)
	return baseQty * stable.spotPrice
}

// Get the value of Quotes including the Limit
func (stable *stableutil) GetUpdatedConfig() names.TradeConfig {
	sideConfig := stable.getCalculateSideConfig()
	// sideConfig.LimitType = names.RatePercent
	// sideConfig.StopLimit = fLimit.GetStopLimit()
	// sideConfig.LockDelta = (sideConfig.StopLimit / 100) * sideConfig.LockDelta
	config := stable.config
	if stable.calculatedSide.IsBuy() {
		config.Buy = sideConfig
	} else {
		config.Sell = sideConfig
	}
	return config
}

func (stable *stableutil) CalculateFiatLimit() float64 {
	sideConfig := stable.getCalculateSideConfig()
	stopLimit := sideConfig.StopLimit
	side := stable.calculatedSide

	if stopLimit == 0 || sideConfig.LimitType.IsFixed() {
		// dont change limit if it is fixed. TODO test this
		return 0
	}

	quoteQty := stable.QuantityInQuoteValue(stable.account)
	fiatStopLimit := calculateExitPrice(stable.spotPrice, quoteQty, stopLimit, stable.commision, side)

	//TODO test that deviation actually happens on this fiatExit amount for both buy and sell cases
	stopLimitPercent := helper.CalculatePercentageChange(fiatStopLimit, stable.spotPrice)

	sideConfig.StopLimit = math.Abs(stopLimitPercent)
	configUpdate := stable.config

	if side.IsBuy() {
		configUpdate.Buy = sideConfig
	} else {
		configUpdate.Sell = sideConfig
	}
	stable.config = configUpdate
	return stopLimitPercent
}

// calculates the exit price required to achieve a desired profit after accounting for trading fees.
// All values are in fiat or stable currency.
//
// Parameters:
//
//	initialPrice: The initial asset price.
//	initialInvestment: The initial investment amount.
//	desiredProfit: The desired profit amount.
//	tradingFeeRate: The trading fee rate (e.g., 0.001 for 0.1%).
//
// Returns:
//
//	The exit price needed to achieve the desired profit.
func calculateExitPrice(spotPrice, quoteQuantity, desiredFiatProfit, tradingFeeRate float64, side names.TradeSide) float64 {
	tradePrice := quoteQuantity / spotPrice
	tradingFee := tradePrice * spotPrice * tradingFeeRate
	totalRequiredAmount := quoteQuantity + desiredFiatProfit + tradingFee

	if side.IsBuy() {
		totalRequiredAmount = quoteQuantity - desiredFiatProfit - tradingFee
	}
	exitPrice := totalRequiredAmount / tradePrice
	return exitPrice
}

func deviationTriggerPrice(initialPrice, initialInvestment, desiredProfit, tradingFeeRate float64, side names.TradeSide) float64 {
	tradePrice := initialInvestment / initialPrice
	tradingFee := tradePrice * initialPrice * tradingFeeRate
	totalRequiredAmount := initialInvestment + desiredProfit + tradingFee

	if side.IsSell() {
		//trigger if it goes below this point
		totalRequiredAmount = initialInvestment - desiredProfit - tradingFee
	}
	exitPrice := totalRequiredAmount / tradePrice
	return exitPrice
}

// Calculates what price will lead to the users balance change
// by said deviation fiat instead of price deviation in price like other deviatiom
func (fiat *stableutil) CalculateFiatDeviation() float64 {
	sideConfig := fiat.getCalculateSideConfig()
	deviation := sideConfig.DeviationSync
	side := fiat.calculatedSide

	if deviation.Delta == 0 {
		return 0
	}

	quoteQty := fiat.QuantityInQuoteValue(fiat.account)
	baseQty := fiat.QuantityInBaseValue(fiat.account)
	deviationTriggerPrice := deviationTriggerPrice(fiat.spotPrice, quoteQty, deviation.Delta, fiat.commision, side)

	//TODO test that deviation actually happens on this fiatExit amount for both buy and sell cases
	deviationInPercent := helper.CalculatePercentageChange(deviationTriggerPrice, fiat.spotPrice)

	formattedString := fmt.Sprintf(
		"\n== %s deviation summary %s ===\n"+
			"Spot Price      :    %s\n"+
			"Trigger Point   :    %s\n"+
			"Spot Percentage :    %f%%\n"+
			"Base Quantity   :    %s\n"+
			"Quote Quantity  :    %s\n"+
			"Trading Fee     :    %f",
		fiat.config.Symbol.FormatQuotePrice(deviation.Delta),
		side,
		fiat.config.Symbol.FormatQuotePrice(fiat.spotPrice),
		fiat.config.Symbol.FormatQuotePrice(deviationTriggerPrice),
		deviationInPercent,
		fiat.config.Symbol.FormatBasePrice(baseQty),
		fiat.config.Symbol.FormatQuotePrice(quoteQty),
		fiat.commision,
	)
	_ = formattedString
	// utils.LogInfo(formattedString)
	sideConfig.DeviationSync.Delta = math.Abs(deviationInPercent)
	configUpdate := fiat.config

	if side.IsBuy() {
		configUpdate.Buy = sideConfig
	} else {
		configUpdate.Sell = sideConfig
	}
	fiat.config = configUpdate
	return deviationInPercent
}

// Function updates the stop price of the trading config based on a specified peggedcoin limit.
// Ensures the stop price aligns with the peggedcoin limit, allowing realization of the limit value from trading the balance.
// Takes into account available trading asset balance, current quote price, commission percentage, and peggedcoin limit.
// Returns the updated trade configuration.

func calculateConfigPeggedLimit(configs []names.TradeConfig, spotPriceList map[string]float64, takersFees map[string]float64, account user.AccountInterface) []names.TradeConfig {
	updatedConfigs := []names.TradeConfig{}
	// configsIncrease := make(map[string]map[names.TradeSide]float64)
	sides := []names.TradeSide{names.TradeSideBuy, names.TradeSideSell}

	for _, config := range configs {
		var updatedConfig = config

		//for each side of each config buy and sell calculate config pegged limit
		for _, calculateSide := range sides {
			symbol := updatedConfig.Symbol
			spotPrice := spotPriceList[symbol.String()]

			fiatLimit := newStable(spotPrice, updatedConfig, calculateSide)

			fiatLimit.UseAccount(account).UseCommision(takersFees[symbol.String()])
			fiatLimit.CalculateFiatDeviation()
			fiatLimit.CalculateFiatLimit()
			fiatLimit.CalculateFiatLockDelta()

			updatedConfig = fiatLimit.GetUpdatedConfig()

		}
		updatedConfigs = append(updatedConfigs, updatedConfig)
	}
	return updatedConfigs
}

func getUpdateWithPeggedLimit(configs []names.TradeConfig) []names.TradeConfig {
	account := user.GetAccount()
	symbolList := names.TradeConfigs(configs).ListSymbol()
	fees := names.GetTradeFees(symbolList)

	takersFees := make(map[string]float64)
	for symbol, fee := range fees {
		takersFees[symbol] = fee.TakerCommission
	}

	// convertDeltaStop to percentage implementation, remove old implementation

	spotPrices, err := binance.GetSymbolPrices(symbolList)
	if err != nil {
		panic("ERROR>>>>")
	}
	updatedConfig := calculateConfigPeggedLimit(configs, spotPrices, takersFees, account)
	return updatedConfig
}

func (fiat *stableutil) CalculateFiatLockDelta() float64 {
	sideConfig := fiat.getCalculateSideConfig()
	lockDelta := sideConfig.LockDelta
	side := fiat.calculatedSide

	if lockDelta == 0 {
		//dont change limit if it is fixed. TODO test this
		return lockDelta
	}

	quoteQty := fiat.QuantityInQuoteValue(fiat.account)
	fiatLockDelta := (quoteQty + lockDelta) / (quoteQty / fiat.spotPrice)

	//TODO test that deviation actually happens on this fiatExit amount for both buy and sell cases
	lockDeltaFiatPercent := helper.CalculatePercentageChange(fiatLockDelta, fiat.spotPrice)

	sideConfig.LockDelta = math.Abs(lockDeltaFiatPercent)
	configUpdate := fiat.config

	if side.IsBuy() {
		configUpdate.Buy = sideConfig
	} else {
		configUpdate.Sell = sideConfig
	}
	fiat.config = configUpdate
	return lockDeltaFiatPercent
}

func getStableTradeConfigs(configs []names.TradeConfig) []names.TradeConfig {
	account := user.GetAccount()
	symbolList := names.TradeConfigs(configs).ListSymbol()
	fees := names.GetTradeFees(symbolList)

	takersFees := make(map[string]float64)
	for symbol, fee := range fees {
		takersFees[symbol] = fee.TakerCommission
	}

	// convertDeltaStop to percentage implementation, remove old implementation

	spotPrices, err := binance.GetSymbolPrices(symbolList)
	if err != nil {
		panic("ERROR>>>>" )
	}
	updatedConfig := calculateConfigPeggedLimit(configs, spotPrices, takersFees, account)
	return updatedConfig
}

type StableTradeParam struct {
	QuoteAsset         string          `json:"quoteAsset"`
	SellDeviationDelta float64         `json:"sellDeviationDelta"`
	SellStopLimit      float64         `json:"sellStopLimit"`
	SellLockDelta      float64         `json:"sellLockDelta"`
	BuyDeviationDelta  float64         `json:"buyDeviationDelta"`
	BuyStopLimit       float64         `json:"buyStopLimit"`
	BuyLockDelta       float64         `json:"buyLockDelta"`
	BestSide           names.TradeSide `json:"bestSide"`
	Status             status          `json:"status"`
	MinPriceChange     float64         `json:"minPriceChange"`
	MaxPriceChange     float64         `json:"maxPriceChange"`
	Side               names.TradeSide `json:"side"`
}

// Fetch a list of assets and decorate them
func GenerateStableTradeConfigs(params StableTradeParam) []names.TradeConfig {
	var symbols []names.Symbol
	var tradingConfigs []names.TradeConfig

	if utils.Env().IsMock() {
		symbols = []names.Symbol{"BTCUSDT", "BNBUSDT"}
	} else {
		stats := binance.GetSymbolStats()
		// We select asset with at most 2% and increase, this
		// group of increase always indicate entry bull and have not peaked yet
		// meaning there is still room for growth

		minimumPrice, maximumPrice := 20.0, 31.0

		if params.MaxPriceChange != 0 {
			maximumPrice = params.MaxPriceChange
		}
		if params.MinPriceChange != 0 {
			minimumPrice = params.MinPriceChange
		}

		if len(stats) == 0 {
			return tradingConfigs
		}

		for _, s := range stats[0:90] {
			symbol := names.Symbol(s.Symbol)
			change := s.PriceChangePercent
			if symbol.ParseTradingPair().Quote == params.QuoteAsset && change > minimumPrice && change < maximumPrice {
				symbols = append(symbols, symbol)
			}
		}

		info := names.GetNewInfo()
		symbols = info.FilterSpotable(symbols)

		for _, s := range symbols {
			fmt.Println(s)
		}
	}

	for _, Symbol := range symbols {
		tradingConfigs = append(tradingConfigs, initConfig(Symbol, params))
	}
	return tradingConfigs
}

// Create a blueprint tradeConfig for this symbol using the params
// Note this function does not assign a side to the newly created config
func initConfig(symbol names.Symbol, params StableTradeParam) names.TradeConfig {
	side := names.TradeSideBuy
	if helper.SideIsValid(params.Side) {
		side = params.Side
	}
	config := names.TradeConfig{
		Symbol:    symbol,
		IsCyclick: true,
		Side:      side,
		Buy: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			LockDelta:  params.BuyLockDelta,
			Quantity:   names.MAX_QUANTITY,
			StopLimit:  params.BuyStopLimit,
			DeviationSync: names.DeviationSync{
				Delta: params.BuyDeviationDelta,
			},
		},
		Sell: names.SideConfig{
			MustProfit: true,
			LimitType:  names.RatePercent,
			LockDelta:  params.SellLockDelta,
			Quantity:   names.MAX_QUANTITY,
			StopLimit:  params.SellStopLimit,
			DeviationSync: names.DeviationSync{
				Delta: params.SellDeviationDelta,
			},
		} ,
	}
	return config
}

// re-initialise this trade config as stable using the init param maintining the configs, Id, Side, symbol
func renitTradeConfig(config names.TradeConfig, initParams StableTradeParam) names.TradeConfig {
	cfg := initConfig(config.Symbol, initParams)
	cfg.Id = config.Id
	cfg.Side = config.Side
	stableConfig := getStableTradeConfigs(names.NewIdTradeConfigs(cfg))
	return stableConfig[0]
}

func generateStableParams(quoteAmount float64, quoteAsset string) StableTradeParam {
	baseAmount := 900.0
	refParam := StableTradeParam{
		QuoteAsset:         quoteAsset,
		BuyStopLimit:       30,
		BuyDeviationDelta:  10,
		BuyLockDelta:       0.5,
		SellStopLimit:      8,
		SellDeviationDelta: 20,
		SellLockDelta:      0.02,
		BestSide:           names.TradeSideSell,
		Status:             StatusContention,
		MinPriceChange:     2,
		MaxPriceChange:     20,
	} 
	refParam = StableTradeParam{
		QuoteAsset: quoteAsset,
		BuyStopLimit:       8,
		BuyDeviationDelta:  5,
		BuyLockDelta:       0.5,
		SellStopLimit:      4,
		SellDeviationDelta: 10,
		SellLockDelta:      0.02,
		BestSide:           names.TradeSideSell,
		Status:             StatusContention,
		MinPriceChange:     2,
		MaxPriceChange:     20,
	}
	increase := (quoteAmount / baseAmount)
	params := StableTradeParam{
		QuoteAsset:         quoteAsset,
		BuyStopLimit:       refParam.BuyStopLimit * increase,
		BuyDeviationDelta:  refParam.BuyDeviationDelta * increase,
		SellLockDelta:      refParam.SellLockDelta * increase,
		BuyLockDelta:       refParam.BuyLockDelta * increase,
		SellStopLimit:      refParam.SellStopLimit * increase,
		SellDeviationDelta: refParam.SellDeviationDelta * increase,
		BestSide:           names.TradeSideSell,
		Status:             StatusContention,
		MinPriceChange:     refParam.MinPriceChange,
		MaxPriceChange:     refParam.MaxPriceChange,
	}
	// params = StableTradeParam{
	// 	QuoteAsset: quoteAsset,
	// 	BuyStopLimit:       8,
	// 	BuyDeviationDelta:  5,
	// 	BuyLockDelta:       0.5,
	// 	SellStopLimit:      4,
	// 	SellDeviationDelta: 10,
	// 	SellLockDelta:      0.02,
	// 	BestSide:           names.TradeSideSell,
	// 	Status:             StatusContention,
	// 	MinPriceChange:     2,
	// 	MaxPriceChange:     20,
	// }

	r, e := json.MarshalIndent(params, "", "    ")
	fmt.Println(string(r), e)
	return params
}
