package helper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"trading/names"
)

func GetSideConfig(config names.TradeConfig) (names.SideConfig, error) {
	if config.Side.IsBuy() {
		return config.Buy, nil
	}
	if config.Side.IsSell(){
		return config.Sell, nil
	}
	return names.SideConfig{}, fmt.Errorf("side Config is invalid")
}

func IsSameConfig(a, b names.TradeConfig) bool {
	return a.Side == b.Side && a.Symbol == b.Symbol
}

func SideIsValid(side names.TradeSide) bool {
	return side.IsSell() || side.IsBuy() 
}

func Stringify(data interface{}) string {
	jsonBytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err.Error()
	}
	return string(jsonBytes)
}

// calculate what percent of Price is pricePercent in the buy context
// returns the value of price at pricePercent (10, 100) = 90
func calculateBuyPriceFromPercent(pricePercent, price float64) float64 {
	discount := pricePercent / 100.0
	return price * (1 - discount)
}

// Calculates the selling price based on a given percentage increase.
// It takes the percentage and the original price as input and returns the sell price. E,g (10, 100) = 110
func calculateSellPriceFromPercent(percentage, price float64) float64 {
	increase := percentage / 100.0
	return price * (1 + increase)
}

func CalculateTradeBuyFixedPrice(price names.SideConfig, priceRate float64) float64 {
	if price.LimitType.IsPercent() {
		return calculateBuyPriceFromPercent(price.StopLimit, priceRate)
	}
	return price.StopLimit
}

// calculates the sell price for the Price symbol based on the given price type and rate.
// If the price type is trade.RatePercent, the sell price will be the PriceRate percentage of the last trade price.
// If the price type is 'fixed', the sell price will be a fixed amount, the PriceRate.
// returns the calculated sell price for the given Price.
func CalculateTradeFixedSellPrice(price names.SideConfig, priceRate float64) float64 {
	if price.LimitType.IsPercent() {
		return calculateSellPriceFromPercent(price.StopLimit, priceRate)
	}
	return price.StopLimit
}

// Calculate both the fixed and percentage trade price of this config
func CalculateTradePrice(trade names.TradeConfig, priceRate float64) struct {
	Limit   float64
	Percent float64
} {
	// if trade.Side != names.TradeSideBuy || trade.Side != names.TradeSideSell {
	// 	panic(fmt.Sprintf("Unknown Trade Action %s", trade.Side))
	// }
	var fixedPrice, percent float64
	if trade.Side.IsBuy() {
		fixedPrice = CalculateTradeBuyFixedPrice(trade.Buy, priceRate)

		percent = trade.Buy.StopLimit
		if trade.Buy.LimitType == names.RateFixed {
			//assumes buy position is always lower than current position
			percent = CalculatePercentageChange(fixedPrice, priceRate)
		}
	} else {
		fixedPrice = CalculateTradeFixedSellPrice(trade.Sell, priceRate)

		percent = trade.Sell.StopLimit
		if trade.Sell.LimitType == names.RateFixed {
			//assumes sell position is always higher than current position
			percent = CalculatePercentageChange(fixedPrice, priceRate)
		}
	}
	return struct {
		Limit   float64
		Percent float64
	}{Limit: fixedPrice, Percent: math.Abs(percent)} //GetUnitPercentageOfPrice(fixedPrice-priceRate, priceDifference)

}

type TradeFee struct {
	Value  float64
	String string
}

func GetTradeFee(trade names.TradeConfig, currentPrice float64) TradeFee {
	//TODO IMPLEMENT FOR BUY
	quanity := trade.Sell.Quantity

	fee := (float64(quanity) * currentPrice) * 0.001
	return TradeFee{
		Value:  fee,
		String: trade.Symbol.FormatQuotePrice(fee),
	}
}

// calculates the percentage growth from an old value to a new value.
// E.g newValue = 10, oldValue = 5, newValue value has grown by 100% i.e 5 of oldValue value
//
// Parameters:
// newValue (float64): The new value.
// oldValue (float64): The old value.
//
// Returns:
// The percentage by which the oldValue has grown to become the newValue.
func CalculatePercentageChange(newValue, oldValue float64) float64 {
	// divident := oldValue
	// if oldValue < newValue {
	// 	divident = newValue
	// }
	// max, min := math.Max(newValue, oldValue), math.Min(newValue, oldValue)
	// percentageIncrease := ((max - min) / min) * 100
	percentageIncrease := ((newValue - oldValue) / oldValue) * 100
	return percentageIncrease
}

// CalculateValueOfPercentage calculates the amount that represents the given percentage of the value.
// It takes a value and a percentage as input and returns the calculated amount.
// The percentage should be between 0 and 100 (inclusive).
func CalculateValueOfPercentage(value, percentage float64) float64 {
	if percentage <= 0 || percentage > 100 {
		return 0
	}
	return (value * percentage) / 100
}

// CalculatePercentageOfValue calculates the percentage that 'value' represents in comparison to 'baseValue'.
// Parameters:
//
//	value: The value for which you want to calculate the percentage.
//	baseValue: The reference value against which the percentage is calculated.
//
// Returns:
//
//	The calculated percentage that 'value' represents in relation to 'baseValue'.
func CalculatePercentageOfValue(price, priceUnit float64) float64 {
	return (priceUnit / price) * 100.0
}

func WriteStringToFile(filename, content string) error {
	file, err := os.OpenFile(fmt.Sprintf("./logs/%s", filename), os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	existingContent, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	contentToWrite := append([]byte(content), existingContent...)

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	if _, err := file.Write(contentToWrite); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func SwitchTradeSide(current names.TradeSide) names.TradeSide {
	if current.IsBuy() {
		return names.TradeSideSell
	}
	return names.TradeSideBuy
}
