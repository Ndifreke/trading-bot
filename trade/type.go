package trade

const (
	PriceTypePercent PriceType = "PERCENT"
	PriceTypeFixed PriceType = "FIXED"
	TradeActionBuy TradeAction = "BUY"
	TradeActionSell TradeAction = "SELL"
)

type PriceType string
func (pt PriceType) IsPercent() bool {
	return pt == PriceTypePercent
}

func (pt PriceType) IsFixed() bool {
	return pt == PriceTypeFixed
}


type TradeAction string
func (ta TradeAction) IsBuy() bool {
	return ta == TradeActionBuy
}
func (ta TradeAction) IsSell() bool {
	return ta == TradeActionSell
}

type Price struct {
	Value      float32
	Type       PriceType //PERCENT, FIXED_VALUE
	Quantity   int
	MustProfit bool
}

type TradeConfig struct {
	Price struct {
		Sell Price
		Buy  Price
	}
	Action        TradeAction //BUY or SELL
	StopCondition bool        // a complex condition expression here
	Symbol        []string
	IsCyclick     bool // Will run both sell and buy after each other is completed
}

type TradeRunner interface {
	BuyRun()
	SellRun()
	Run()
}
