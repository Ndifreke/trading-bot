package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
)

func GetBinanceAccount() *binance.Account {
	s, _ := GetClient().NewGetAccountService().Do(context.Background())
	return s
}

