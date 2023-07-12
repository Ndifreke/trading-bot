package binance

import (
	"context"
	"fmt"
	"strconv"
	"github.com/adshao/go-binance/v2"
	"trading/utils"
)

type Order = *binance.Order
type OrderHistory []Order

func GetOpenOrders() []Order { //api.RequestResponse[[]OrderJson]
	orders, err := GetClient().NewListOpenOrdersService().Do(context.Background())
	// var bn = New[[]OrderJson](apiArg{Api: Endpoints.OpenOrders})
	// return bn.Request()
	if err != nil {
		utils.LogError(err, "GetOpenOrders")
	}
	return orders
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

func CreateOrder(symbol string, quantity float64, side string, orderType binance.OrderType) (*binance.CreateOrderResponse, error) {
	data, err := GetClient().
		NewCreateOrderService().
		Side(binance.SideType(side)).
		Symbol(symbol).
		Quantity(strconv.FormatFloat(quantity, 'f', -1, 64)).
		Type(orderType).
		Do(context.Background())
	return data, err
}

func CreateBuyMarketOrder(symbol string, quantity float64) (*binance.CreateOrderResponse, error) {
	return CreateOrder(symbol, quantity, "BUY", binance.OrderTypeMarket)
}

func CreateSellMarketOrder(symbol string, quantity float64) (*binance.CreateOrderResponse, error) {
	return CreateOrder(symbol, quantity, "SELL", binance.OrderTypeMarket)
}
