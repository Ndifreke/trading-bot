package trade

import "fmt"

const (
	RatePercent     RateType    = "PERCENT"
	RateFixed       RateType    = "FIXED"
	TradeActionBuy  TradeAction = "BUY"
	TradeActionSell TradeAction = "SELL"
)

type RateType string

func (pt RateType) IsPercent() bool {
	return pt == RatePercent
}

func (pt RateType) IsFixed() bool {
	return pt == RateFixed
}

type TradeAction string

func (ta TradeAction) IsBuy() bool {
	// fmt.Println(">>>>>",ta)
	// if ta != TradeActionBuy{
	// 	panic(fmt.Sprintf("Invalid Buy Trading Action %s", ta))
	// }
	return ta == TradeActionBuy
}

func (ta TradeAction) IsSell() bool {
	// if ta != TradeActionSell {
	// 	panic(fmt.Sprintf("Invalid Sell Trading Action %s", ta))
	// }
	return ta == TradeActionSell
}
func (ta TradeAction) String() string {
	return string(ta)
}

type Price struct {
	PriceRate     float32
	PriceRateType RateType //PERCENT, FIXED_VALUE
	Quantity      int
	MustProfit    bool
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
	Action        TradeAction //BUY or SELL
	StopCondition bool        // a complex condition expression here
	Symbol        Symbol
	IsCyclick     bool // Will run both sell and buy after each other is completed
}

type TradeRunner interface {
	BuyRun(t TradeConfig)
	SellRun(t TradeConfig)
	RunAll()
}
