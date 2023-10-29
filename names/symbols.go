package names

import (
	"fmt"
	binLib "github.com/adshao/go-binance/v2"
	"math"
	// "strconv"
	"strings"
)

type Symbol string

type TradingPair struct {
	Base  string
	Quote string
}

func (tp TradingPair) String() Symbol {
	return Symbol(tp.Base + tp.Quote)
}

// ParseTradingPair parses a trading pair string into a TradingPair struct.
func (s Symbol) ParseTradingPair() TradingPair {
	return GetStoredInfo().ToPair(s.String())
}

func (s Symbol) String() string {
	return string(s)
}

func (s Symbol) Gains(pretradeValue, spotValue float64, side TradeSide) string {
	sellGains := spotValue - pretradeValue
	if side == TradeSideBuy && sellGains <= math.MinInt64 {
		sellGains = math.Copysign(sellGains, 1)
	}
	return s.FormatQuotePrice(sellGains)
}

// FormatBasePrice formats a price as a string with the base currency symbol.
func (s Symbol) FormatBasePrice(price float64) string {
	baseSymbol := s.ParseTradingPair().Base
	return fmt.Sprintf("%f%s", price, baseSymbol)
}

// FormatQuotePrice formats a price as a string with the quote currency symbol.
func (s Symbol) FormatQuotePrice(price float64) string {
	quoteSymbol := s.ParseTradingPair().Quote
	return fmt.Sprintf("%f%s", price, quoteSymbol)
}

func (s Symbol) Info() binLib.Symbol {
	for _, symbol := range GetStoredInfo().symbols {
		if s.String() == symbol.Symbol {
			return symbol
		}
	}
	return binLib.Symbol{}
}

// func (s Symbol) Quantity(quantity float64) float64 {
// 	stepSize, tickSize := s.Filter().StepSize(), s.Filter().TickSize()

// 	formatValue := func(val float64, stepSize string) float64 {
// 		precision := strings.Index(stepSize, "1") - 1
// 		if precision > 0 {
// 			formatString := fmt.Sprintf("%%.%df", precision)
// 			formattedValue, _ := strconv.ParseFloat(fmt.Sprintf(formatString, val), 64)
// 			return formattedValue
// 		}
// 		return math.Floor(val)
// 	}
// 	formated := formatValue(quantity-tickSize, stepSize)
// 	utils.LogError(fmt.Errorf("ok"), fmt.Sprintf("Qty=%f\nStep Size=%s\nFormated=%f", quantity, stepSize, formated))
// 	return formated
// }

func (s Symbol) Quantity(quantity float64) float64 {
	stepSize := s.Filter().StepSize()
	precision := strings.Index(stepSize, "1") - 1
	formated := math.Floor(quantity)

	if precision > 0 {
		power := math.Pow(10, float64(precision))
		return float64(int(quantity*power)) / power
	}
	return formated
}

func (s Symbol) Price(price float64) float64 {
	tickSize := s.Filter().TickSize()
	precision := strings.Index(tickSize, "1") - 1
	formated := price

	if precision > 0 {
		power := math.Pow(10, float64(precision))
		return float64(int(price*power)) / power
	}
	return formated
}

type filter []map[string]interface{}

func (s Symbol) Filter() filter {
	return filter(s.Info().Filters)
}

func (f filter) StepSize() string {
	for _, ff := range f {
		if ff["filterType"] == "LOT_SIZE" {
			return ff["stepSize"].(string)
		}
	}
	return ""
}

func (f filter) TickSize() string {
	for _, ff := range f {
		if ff["filterType"] == "PRICE_FILTER" {
			if tickSizeStr, ok := ff["tickSize"].(string); ok {
				return tickSizeStr
			}
		}
	}
	return ""
}
