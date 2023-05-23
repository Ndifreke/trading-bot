package binance

import (
	"context"
	"github.com/adshao/go-binance/v2"
)

func GetTradeHistories(symbol string) []*binance.Trade {
	s, _ := GetClient().NewHistoricalTradesService().Symbol(symbol).Do(context.Background())
	return s
}

