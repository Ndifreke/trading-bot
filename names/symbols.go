package names

import (
	"fmt"
	"math"
	"strconv"

	binLib "github.com/adshao/go-binance/v2"
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

func (s Symbol) Quantity(quantity float64) float64 {
	stepSize := 0.0
	for _, f := range s.Info().Filters {
		if f["filterType"] == "LOT_SIZE" {
			stepSize, _ = strconv.ParseFloat(f["stepSize"].(string), 64)
			if stepSize == 0 {
				return quantity
			}
			break
		}
	}
	return math.Floor(quantity/stepSize) * stepSize
}
