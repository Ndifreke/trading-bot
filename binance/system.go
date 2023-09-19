package binance

import (
	"context"
	"encoding/json"
	"strconv"
	"sync"
	"trading/constant"
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

func loadInfoString() binance.ExchangeInfo {
	var exchange binance.ExchangeInfo
	json.Unmarshal([]byte(constant.ExchangeInfo), &exchange)
	return exchange
}

var exchangeInfo binance.ExchangeInfo = binance.ExchangeInfo{}

func LoadExchangeInfo() binance.ExchangeInfo {
	if exchangeInfo.ServerTime == 0 {
		data := loadInfoString()
		exchangeInfo = data
	}
	return exchangeInfo
}

type tradeFeeDetails struct {
	Symbol          string
	MakerCommission float64
	TakerCommission float64
}

func GetTradeFees(symbols []string) map[string]tradeFeeDetails {
	symbolFees := make(map[string]tradeFeeDetails)
	var wg sync.WaitGroup

	wg.Add(len(symbols))
	for _, symbol := range symbols {
		go func(s string) {
			if utils.Env().IsTest() {
				symbolFees[s] = tradeFeeDetails{
					Symbol:          s,
					MakerCommission: 0.001,
					TakerCommission: 0.001,
				}
				wg.Done()
				return
			}

			data, err := GetClient().NewTradeFeeService().Symbol(s).Do(context.Background())
			if err != nil {
				utils.LogError(err, "Could not get trading fees")
				wg.Done()
				return
			}
			MakerCommission, _ := strconv.ParseFloat(data[0].MakerCommission, 64)
			TakerCommission, _ := strconv.ParseFloat(data[0].TakerCommission, 64)
			symbolFees[s] = tradeFeeDetails{
				Symbol:          data[0].Symbol,
				MakerCommission: MakerCommission,
				TakerCommission: TakerCommission,
			}

			wg.Done()
		}(symbol)
	}
	wg.Wait()
	return symbolFees
}

// func GetSymbolPrices(symbols []string) (map[string]float64, error) {

// 	if utils.Env().IsTest() {
// 		var prices = make(map[string]float64)
// 		for _, sym := range symbols {
// 			prices[sym] = utils.Env().RandomNumber()
// 		}
// 		return prices, nil
// 	}

// 	var postRunPrices = make(map[string]float64)
// 	prices, err := GetClient().NewListPricesService().Symbols(symbols).Do(context.Background())
// 	if err != nil {
// 		utils.LogError(err, "GetSymbolPrices()")
// 		return nil, err
// 	}
// 	for _, price := range prices {
// 		f, _ := strconv.ParseFloat(price.Price, 64)
// 		postRunPrices[price.Symbol] = f
// 	}
// 	return postRunPrices, err
// }
