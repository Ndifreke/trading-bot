package binance

import (
	"context"
	"trading/utils"
	"github.com/adshao/go-binance/v2"
)

func ExchangeInfo() *binance.ExchangeInfo {
	data, err := GetClient().NewExchangeInfoService().Do(context.Background())
	if err != nil {
		utils.LogError(err, "ExchangeInfo()")
		return nil
	}
	return data
}
