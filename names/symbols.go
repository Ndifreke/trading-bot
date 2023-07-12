package names

import (
	"fmt"
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

type Symbols struct {
	symbols []binanceLib.Symbol
}

var exchangeInfo *binanceLib.ExchangeInfo
func GetSymbols() Symbols {
	return Symbols{
		symbols: getExchangeInfo().Symbols,
	}
}

func getExchangeInfo() *binanceLib.ExchangeInfo {
	if exchangeInfo == nil {
		exchangeInfo = binance.ExchangeInfo()
		return exchangeInfo
	}
	return exchangeInfo
}

type SymbolPrecision struct {
	Quote int
	Base  int
}

func (smb Symbols) PreciseValue(symbol string, value float64) float64 {
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
	toPrecision := fmt.Sprintf("%.*f", countDecimalPlaces(stepSize), value - stepSize)
	m, _ := strconv.ParseFloat(toPrecision, 64)
	return m
} 

func (smb Symbols) ToPair(symbol string) TradingPair {
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
