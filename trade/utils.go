package trade

import (
	"fmt"
	"trading/binance"
)

func GetSymbolPrices(symbols []string) map[string]float32 {
	var postRunPrices = make(map[string]float32, 2)
	for _, symbol := range symbols {
		price := binance.GetPriceLatest(symbol)
		fmt.Println(price.Body.Price, price.Body.Mins)
		if price.Ok {
			postRunPrices[symbol] = price.Body.Price
		}
	}
	return postRunPrices
}

func buyPriceFromPercent(percentage, preTradePrice float32) float32 {
	buyPrice := preTradePrice - ((percentage / 100) * preTradePrice)
	return buyPrice
}

func sellPriceBuyPercent(percentage, preTradePrice float32) float32 {
	buyPrice := preTradePrice + ((percentage / 100) * preTradePrice)
	return buyPrice
}

func CalculateTradeBuyPrice(price Price, preTradePrice float32) float32 {
	if price.Type.IsPercent() {
		return buyPriceFromPercent(price.Value, preTradePrice)
	}
	return price.Value
}

func CalculateTradeSellPrice(price Price, preTradePrice float32) float32 {

	if price.Type.IsPercent() {
		return sellPriceBuyPercent(price.Value, preTradePrice)
	}
	return price.Value
}

func CalculateTradePrice(trade TradeConfig, preTradePrice float32) float32 {
	switch trade.Action {
	case "BUY":
		return CalculateTradeBuyPrice(trade.Price.Buy, preTradePrice)
	case "SELL":
		return CalculateTradeSellPrice(trade.Price.Sell, preTradePrice)
	}
	panic(fmt.Sprintf("Unknown Trade Action %s", trade.Action))
}
