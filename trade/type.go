package trade

import (
	"fmt"
	"trading/request"
)

const (
	RatePercent   RateType  = "PERCENT"
	RateFixed     RateType  = "FIXED"
	TradeSideBuy  TradeSide = "BUY"
	TradeSideSell TradeSide = "SELL"
)

type RateType string

func (pt RateType) IsPercent() bool {
	return pt == RatePercent
}

func (pt RateType) IsFixed() bool {
	return pt == RateFixed
}

type TradeSide string

func (ta TradeSide) IsBuy() bool {
	return ta == TradeSideBuy
}

func (ta TradeSide) IsSell() bool {
	return ta == TradeSideSell
}

func (ta TradeSide) String() string {
	return string(ta)
}

// func (ta TradeAction) GetLockDelta(cfg TradeConfig) string {
// 	if ta.IsBuy(){

// 	}
// }

type Price struct {
	PriceRate     float32
	PriceRateType RateType //PERCENT, FIXED_VALUE
	Quantity      int
	MustProfit    bool
	// determines what percentage change in price to lock positive price movement
	//
	LockDelta float32
}

type Symbol string

type TradingPair struct {
	Base  string
	Quote string
}

// ParseTradingPair parses a trading pair string into a TradingPair struct.
func (s Symbol) ParseTradingPair() TradingPair {
	return TradingPair{
		Base:  string(s)[:3],
		Quote: string(s)[3:],
	}
}

func (s Symbol) String() string {
	return string(s)
}

// FormatBasePrice formats a price as a string with the base currency symbol.
func (s Symbol) FormatBasePrice(price float32) string {
	baseSymbol := s.ParseTradingPair().Base
	return fmt.Sprintf("%f %s", price, baseSymbol)
}

// FormatQuotePrice formats a price as a string with the quote currency symbol.
func (s Symbol) FormatQuotePrice(price float32) string {
	quoteSymbol := s.ParseTradingPair().Quote
	return fmt.Sprintf("%f %s", price, quoteSymbol)
}

type TradeConfig struct {
	Price struct {
		Sell Price
		Buy  Price
	}
	Side          TradeSide //BUY or SELL
	StopCondition bool      // a complex condition expression here
	Symbol        Symbol
	IsCyclick     bool // Will run both sell and buy after each other is completed
}

type Trader interface {
	Run()
}

type TradeManager func(t ...TradeConfig) Trader

type TradeRunner[Data any] struct {
	Config TradeConfig
	Socket *request.Socket[Data]
}
