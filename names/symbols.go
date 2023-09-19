package names

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"trading/binance"

	binanceLib "github.com/adshao/go-binance/v2"
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
	return GetSymbols().ToPair(s.String())
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
	return fmt.Sprintf("%f %s", price, baseSymbol)
}

// FormatQuotePrice formats a price as a string with the quote currency symbol.
func (s Symbol) FormatQuotePrice(price float64) string {
	quoteSymbol := s.ParseTradingPair().Quote
	return fmt.Sprintf("%f %s", price, quoteSymbol)
}

type symbols struct {
	symbols []binanceLib.Symbol
}

func GetSymbols() symbols {
	return symbols{
		symbols: binance.LoadExchangeInfo().Symbols,
	}
}

type SymbolPrecision struct {
	Quote int
	Base  int
}

func (smb symbols) PreciseValue(symbol string, value float64) float64 {
	countDecimalPlaces := func(number float64) int {
		parts := strings.Split(strconv.FormatFloat(number, 'f', -1, 64), ".")
		if len(parts) > 1 {
			return len(parts[1])
		}
		return 0
	}
	var stepSize float64
	for _, s := range smb.symbols {
		if s.Symbol == symbol {
			stepSize, _ = strconv.ParseFloat(s.Filters[1]["stepSize"].(string), 64)
			break
		}
	}
	// value - stepSize Dont sell exactly what is availble to reduce error caused by price shift
	toPrecision := fmt.Sprintf("%.*f", countDecimalPlaces(stepSize), value-stepSize)
	m, _ := strconv.ParseFloat(toPrecision, 64)
	return m
}

func (smb symbols) ToPair(symbol string) TradingPair {
	for _, s := range smb.symbols {
		if s.Symbol == symbol {
			return TradingPair{
				Quote: s.QuoteAsset,
				Base:  s.BaseAsset,
			}
		}
	}
	return TradingPair{}
}

func (smb symbols) List() []string {
	var list []string
	for _, s := range smb.symbols {
		list = append(list, s.Symbol)
	}
	return list
}
