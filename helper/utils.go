package helper

import (
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"math/rand"
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

func CalculateTradeBuyFixedPrice(price names.SideConfig, priceRate float64) float64 {
	if price.RateType.IsPercent() {
		return calculateBuyPriceFromPercent(price.RateLimit, priceRate)
	}
	return price.RateLimit
}

// calculates the sell price for the Price symbol based on the given price type and rate.
// If the price type is trade.RatePercent, the sell price will be the PriceRate percentage of the last trade price.
// If the price type is 'fixed', the sell price will be a fixed amount, the PriceRate.
// returns the calculated sell price for the given Price.
func CalculateTradeFixedSellPrice(price names.SideConfig, priceRate float64) float64 {
	if price.RateType.IsPercent() {
		return calculateSellPriceFromPercent(price.RateLimit, priceRate)
	}
	return price.RateLimit
}

func CalculateTradePrice(trade names.TradeConfig, priceRate float64) struct {
	Limit   float64
	Percent float64
} {
	// if trade.Side != names.TradeSideBuy || trade.Side != names.TradeSideSell {
	// 	panic(fmt.Sprintf("Unknown Trade Action %s", trade.Side))
	// }
	var fixedPrice, percent float64
	if trade.Side == names.TradeSideBuy {
		fixedPrice = CalculateTradeBuyFixedPrice(trade.Buy, priceRate)

		percent = trade.Buy.RateLimit
		if trade.Buy.RateType == names.RateFixed {
			//assumes buy position is always lower than current position
			percent = GetPercentGrowth(fixedPrice, priceRate)
		}
	} else {
		fixedPrice = CalculateTradeFixedSellPrice(trade.Sell, priceRate)

		percent = trade.Sell.RateLimit
		if trade.Sell.RateType == names.RateFixed {
			//assumes sell position is always higher than current position
			percent = GetPercentGrowth(fixedPrice, priceRate)
		}
	}
	return struct {
		Limit   float64
		Percent float64
	}{Limit: fixedPrice, Percent: math.Abs(percent)} //GetUnitPercentageOfPrice(fixedPrice-priceRate, priceDifference)

}

type TradeFee struct {
	Value  float64
	String string
}

func GetTradeFee(trade names.TradeConfig, currentPrice float64) TradeFee {
	//TODO IMPLEMENT FOR BUY
	quanity := trade.Sell.Quantity
	// if trade.Action.IsBuy() {
	// 	quanity = trade.Buy.Quantity
	// }
	// price := binance.GetPriceLatest(symbol).Body.Price
	fee := (float64(quanity) * currentPrice) * 0.001
	return TradeFee{
		Value:  fee,
		String: trade.Symbol.FormatQuotePrice(fee),
	}
}

// finds what percent of oldValue value has newValue value grown by
// E.g newValue = 10, oldValue = 5, newValue value has grown by 100% i.e 5 of oldValue value
func GetPercentGrowth(newValue, oldValue float64) float64 {
	// divident := oldValue
	// if oldValue < newValue {
	// 	divident = newValue
	// }
	// max, min := math.Max(newValue, oldValue), math.Min(newValue, oldValue)
	// percentageIncrease := ((max - min) / min) * 100
	percentageIncrease := ((newValue - oldValue) / oldValue) * 100
	return percentageIncrease
}

// Get what percentage of price is priceUnit
func GetUnitPercentageOfPrice(price, priceUnit float64) float64 {
	return (priceUnit / price) * 100.0
}

var unused = []string{
	"SOLUSDT",
	"MKRUSDT",
	"SNXUSDT",
	"CRVUSDT",
}

var TradeSymbolList = []string{
	"ACAUSDT",
	"FLMUSDT",
	"SOLUSDT",
	"MKRUSDT",
	"SNXUSDT",
	"CRVUSDT",
	"POWRUSDT",
	"BTCUSDT",
	"BTCBUSD",
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
	"CTSIUSDT",
	"YFIIUSDT",
	"PAXUSDT",
	"ZENUSDT",
	"REEFUSDT",
	"DODOUSDT",
	"MTLUSDT",
	"CELRUSDT",
	"BNBBUSD",
}





func GenerateTradeConfigs(symbols []string) []names.TradeConfig {


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
			Sell: names.SideConfig{
				RateLimit:  rateLimitSell,
				RateType:   names.RatePercent,
				Quantity:   quantitySell,
				MustProfit: mustProfitSell,
				LockDelta:  lockDeltaSell,
			},
			Buy: names.SideConfig{
				RateLimit:  rateLimitBuy,
				RateType:   names.RatePercent,
				Quantity:   quantityBuy,
				MustProfit: mustProfitBuy,
				LockDelta:  lockDeltaBuy,
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


func WriteStringToFile(filePath, content string) error {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("error opening file: %w", err)
	}
	defer file.Close()

	existingContent, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	contentToWrite := append([]byte(content), existingContent...)

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	if _, err := file.Write(contentToWrite); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}
