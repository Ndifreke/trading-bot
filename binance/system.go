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

func LoadStoredExchangeInfo() binance.ExchangeInfo {
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
	rwLock := sync.RWMutex{}

	wg.Add(len(symbols))
	for _, symbol := range symbols {
		go func(s string) {
			if utils.Env().IsTest() {

				rwLock.Lock()

				symbolFees[s] = tradeFeeDetails{
					Symbol:          s,
					MakerCommission: 0.001,
					TakerCommission: 0.001,
				}
				rwLock.Unlock()

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

			rwLock.Lock()

			symbolFees[s] = tradeFeeDetails{
				Symbol:          data[0].Symbol,
				MakerCommission: MakerCommission,
				TakerCommission: TakerCommission,
			}

			rwLock.Unlock()

			wg.Done()
		}(symbol)
	}
	wg.Wait()
	return symbolFees
}
