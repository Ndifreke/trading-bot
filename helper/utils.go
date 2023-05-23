package helper

import (
	"fmt"
	"sync"
	"trading/binance"
	"trading/trade"
)

func GetSymbolPrices(symbols []string) map[string]float32 {
	var wg sync.WaitGroup

	wg.Add(len(symbols))
	var postRunPrices = make(map[string]float32)
	for _, symbol := range symbols {

		go func(symbol string) {
			price := binance.GetPriceLatest(symbol)
			fmt.Println(price.Body.Price, price.Body.Mins)
			if price.Ok {
				//Needless to use a thread Lock here as all symbols are unique
				postRunPrices[symbol] = price.Body.Price
			}
			wg.Done()
		}(symbol)

	}
	defer wg.Wait()
	return postRunPrices
}

func getBuyPriceFromPercent(percentage, preTradePrice float32) float32 {
	buyPrice := preTradePrice - ((percentage / 100) * preTradePrice)
	return buyPrice
}

func getSellPriceBuyPercent(percentage, preTradePrice float32) float32 {
	buyPrice := preTradePrice + ((percentage / 100) * preTradePrice)
	return buyPrice
}

func CalculateTradeBuyPrice(price trade.Price, preTradePrice float32) float32 {
	if price.PriceRateType.IsPercent() {
		return getBuyPriceFromPercent(price.PriceRate, preTradePrice)
	}
	return price.PriceRate
}

// calculates the sell price for the Price symbol based on the given price type and rate.
// If the price type is trade.RatePercent, the sell price will be the PriceRate percentage of the last trade price.
// If the price type is 'fixed', the sell price will be a fixed amount, the PriceRate.
// returns the calculated sell price for the given Price.
func CalculateTradeSellPrice(price trade.Price, lastTradePrice float32) float32 {
	if price.PriceRateType.IsPercent() {
		return getSellPriceBuyPercent(price.PriceRate, lastTradePrice)
	}
	return price.PriceRate
}

func CalculateTradePrice(trade trade.TradeConfig, preTradePrice float32) float32 {
	switch trade.Side {
	case "BUY":
		return CalculateTradeBuyPrice(trade.Price.Buy, preTradePrice)
	case "SELL":
		return CalculateTradeSellPrice(trade.Price.Sell, preTradePrice)
	}
	panic(fmt.Sprintf("Unknown Trade Action %s", trade.Side))
}

type TradeFee struct {
	Value  float32
	String string
}

func GetTradeFee(trade trade.TradeConfig, currentPrice float32) TradeFee {
	quanity := trade.Price.Sell.Quantity
	// if trade.Action.IsBuy() {
	// 	quanity = trade.Price.Buy.Quantity
	// }
	// price := binance.GetPriceLatest(symbol).Body.Price
	fee := (float32(quanity) * currentPrice) * 0.001
	return TradeFee{
		Value:  fee,
		String: trade.Symbol.FormatQuotePrice(fee),
	}
}
