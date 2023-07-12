package binance

import (
	"context"
	"trading/utils"
	"github.com/adshao/go-binance/v2"
)

func GetBinanceAccount() *binance.Account {
	s, err := GetClient().NewGetAccountService().Do(context.Background())
	if err != nil {
		utils.LogError(err, "ExchangeInfo()")
		return s 
	}
	return s
}

