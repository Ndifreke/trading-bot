package helper

import (
	"fmt"
	"math/rand"
	"time"
	"trading/names"
)

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

func CalculateTradeBuyFixedPrice(price names.Price, pretradePrice float64) float64 {
	if price.RateType.IsPercent() {
		return calculateBuyPriceFromPercent(price.RateLimit, pretradePrice)
	}
	return price.RateLimit
}

// calculates the sell price for the Price symbol based on the given price type and rate.
// If the price type is trade.RatePercent, the sell price will be the PriceRate percentage of the last trade price.
// If the price type is 'fixed', the sell price will be a fixed amount, the PriceRate.
// returns the calculated sell price for the given Price.
func CalculateTradeFixedSellPrice(price names.Price, lastTradePrice float64) float64 {
	if price.RateType.IsPercent() {
		return calculateSellPriceFromPercent(price.RateLimit, lastTradePrice)
	}
	return price.RateLimit
}

func CalculateTradePrice(trade names.TradeConfig, preTradePrice float64) float64 {
	switch trade.Side {
	case "BUY":
		return CalculateTradeBuyFixedPrice(trade.Price.Buy, preTradePrice)
	case "SELL":
		return CalculateTradeFixedSellPrice(trade.Price.Sell, preTradePrice)
	}
	panic(fmt.Sprintf("Unknown Trade Action %s", trade.Side))
}

type TradeFee struct {
	Value  float64
	String string
}

func GetTradeFee(trade names.TradeConfig, currentPrice float64) TradeFee {
	quanity := trade.Price.Sell.Quantity
	// if trade.Action.IsBuy() {
	// 	quanity = trade.Price.Buy.Quantity
	// }
	// price := binance.GetPriceLatest(symbol).Body.Price
	fee := (float64(quanity) * currentPrice) * 0.001
	return TradeFee{
		Value:  fee,
		String: trade.Symbol.FormatQuotePrice(fee),
	}
}

// finds what percent of initial value has current value grown by
// E.g current = 10, initial = 5, current value has grown by 100% i.e 5 of initial value
func GetPercentChange(currentValue, initialValue float64) float64 {
	percentageIncrease := ((currentValue - initialValue) / initialValue) * 100
	return percentageIncrease
}

// Get what percentage of price is priceUnit
func GetUnitPercentageOfPrice(price, priceUnit float64) float64 {
	return (priceUnit / price) * 100.0
}


var TradeSymbolList = []string{
	"BTCUSDT",
	"ETHUSDT",
	"BNBUSDT",
	"XRPUSDT",
	"LTCUSDT",
	"ADAUSDT",
	"DOGEUSDT",
	"DOTUSDT",
	"LINKUSDT",
	"BCHUSDT",
	"XLMUSDT",
	"ETCUSDT",
	"VETUSDT",
	"TRXUSDT",
	"FILUSDT",
	"UNIUSDT",
	"EOSUSDT",
	"AAVEUSDT",
	"XTZUSDT",
	"SOLUSDT",
	"XMRUSDT",
	"DASHUSDT",
	"ATOMUSDT",
	"MKRUSDT",
	"THETAUSDT",
	"AVAXUSDT",
	"COMPUSDT",
	"SNXUSDT",
	"YFIUSDT",
	"FTTUSDT",
	"NEOUSDT",
	"ALGOUSDT",
	"WAVESUSDT",
	"ZECUSDT",
	"HBARUSDT",
	"EGLDUSDT",
	"ENJUSDT",
	"CRVUSDT",
	"RSRUSDT",
	"CHZUSDT",
	"IOSTUSDT",
	"GRTUSDT",
	"ONEUSDT",
	"KSMUSDT",
	"RVNUSDT",
	"RUNEUSDT",
	"ZILUSDT",
	"BTTUSDT",
	"HNTUSDT",
	"ONTUSDT",
	"DGBUSDT",
	"LRCUSDT",
	"ZRXUSDT",
	"MANAUSDT",
	"BATUSDT",
	"SUSHIUSDT",
	"QTUMUSDT",
	"UMAUSDT",
	"NANOUSDT",
	"IOTAUSDT",
	"SXPUSDT",
	// "CRVUSDT",
	"CTSIUSDT",
	"YFIIUSDT",
	// "SNXUSDT",
	"PAXUSDT",
	"ZENUSDT",
	"REEFUSDT",
	"DODOUSDT",
	// "MKRUSDT",
	"MTLUSDT",
	"CELRUSDT",
	"BNBBUSD",
}

func GenerateTradeConfigs(symbols []string) []names.TradeConfig {
	rand.Seed(time.Now().UnixNano())

	tradeConfigs := make([]names.TradeConfig, 0)
	for i := 0; i < len(TradeSymbolList); i++ {
	
		rateLimitSell := rand.Float64() * 5.0
		rateLimitBuy := rand.Float64() * 5.0
		quantitySell := rand.Float64() * 5.0
		quantityBuy := rand.Float64() * 5.0
		mustProfitSell := rand.Intn(2) == 0
		mustProfitBuy := rand.Intn(2) == 0
		lockDeltaSell := rand.Float64() * 10.0
		lockDeltaBuy := rand.Float64() * 10.0

		tradeConfig := names.TradeConfig{
			Price: struct {
				Sell names.Price
				Buy  names.Price
			}{
				Sell: names.Price{
					RateLimit:  rateLimitSell,
					RateType:   names.RatePercent,
					Quantity:   quantitySell,
					MustProfit: mustProfitSell,
					LockDelta:  lockDeltaSell,
				},
				Buy: names.Price{
					RateLimit:  rateLimitBuy,
					RateType:   names.RatePercent,
					Quantity:   quantityBuy,
					MustProfit: mustProfitBuy,
					LockDelta:  lockDeltaBuy,
				},
			},
			Side:          names.TradeSideBuy,
			StopCondition: false,
			Symbol:        names.Symbol(symbols[i]),
			IsCyclick:     false,
		}

		if rand.Intn(2) == 0 {
			tradeConfig.Side = names.TradeSideSell
		}

		tradeConfigs = append(tradeConfigs, tradeConfig)
	}

	return tradeConfigs
}
