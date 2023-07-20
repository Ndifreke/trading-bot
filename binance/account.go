package binance

import (
	"context"
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

func GetBinanceAccount() *binance.Account {
	if utils.Env().IsTest() {
		return &binance.Account{}
	}
	s, err := GetClient().NewGetAccountService().Do(context.Background(), binance.WithRecvWindow(60000))
	if err != nil {
		utils.LogError(err, "GetBinanceAccount()")
		return s
	}
	return s
}
