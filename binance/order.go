package binance

import (
	"context"
	"fmt"
	"strconv"
	"trading/api"
	"trading/trade"
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

type Order = *binance.Order
type OrderHistory []Order

func GetOpenOrders() api.RequestResponse[[]OrderJson] {
	var bn = New[[]OrderJson](apiArg{Api: Endpoints.OpenOrders})
	return bn.Request()
}

func GetOrderHistories(symbol string) OrderHistory {
	s, err := GetClient().NewListOrdersService().Symbol(symbol).Do(context.Background())
	if err != nil {
		utils.LogError(err, fmt.Sprintf("Get %s order histories error", symbol))
	}
	return s
}

func (orders OrderHistory) ListBuy() OrderHistory {
	var buy OrderHistory
	for _, o := range orders {
		if o.Side == binance.SideTypeBuy {
			buy = append(buy, o)
		}
	}
	return buy
}

func (orders OrderHistory) ListSell() OrderHistory {
	var buy OrderHistory
	for _, o := range orders {
		if o.Side == binance.SideTypeSell {
			buy = append(buy, o)
		}
	}
	return buy
}

func (o OrderHistory) Latest() Order {
	return o[len(o)-1]
}

func CreateOrder(symbol trade.Symbol, price float64, quantity float64, side trade.TradeSide, orderType binance.OrderType) (*binance.CreateOrderResponse, error) {
	data, err := GetClient().
		NewCreateOrderService().
		Price(strconv.FormatFloat(price, 'f', -1, 64)).
		Side(binance.SideType(side.String())).
		Symbol(symbol.String()).
		Quantity(strconv.FormatFloat(quantity, 'f', -1, 64)).
		Type(orderType).
		Do(context.Background())
	return data, err
}

func CreateBuyMarketOrder(symbol trade.Symbol, price float64, quantity float64) (*binance.CreateOrderResponse, error) {
	return CreateOrder(symbol, price, quantity, trade.TradeSideBuy, binance.OrderTypeMarket)
}

func CreateSellMarketOrder(symbol trade.Symbol, price float64, quantity float64) (*binance.CreateOrderResponse, error) {
	return CreateOrder(symbol, price, quantity, trade.TradeSideSell, binance.OrderTypeMarket)
}
