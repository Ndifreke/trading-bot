package binance

import (
	"encoding/json"

	"github.com/adshao/go-binance/v2"
)

func loadInfoString() binance.ExchangeInfo {
	var exchange binance.ExchangeInfo
	json.Unmarshal([]byte("exchange_info.json"), &exchange)
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

// func GetExchangeInfo2() binance.ExchangeInfo {
// 	data, err := GetClient().NewExchangeInfoService().Do(context.Background())
// 	if err != nil {
// 		utils.LogError(err, "ExchangeInfo()")
// 		return *data
// 	}
// 	return *data
// }

// func GetExchangeInfo() binance.ExchangeInfo {
// 	if len(exchangeInfo.Symbols) < 2000 {
// 		data, err := GetClient().NewExchangeInfoService().Do(context.Background())
// 		if err != nil {
// 			utils.LogError(err, "ExchangeInfo()")
// 			return *data
// 		}
// 		fmt.Println(data.Symbols)
// 		exchangeInfo = *data
// 	}
// 	return exchangeInfo
// }
