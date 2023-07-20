package binance

import (
	"context"
	"strconv"

	// "fmt"
	"os"

	// "strconv"
	"strings"

	// "time"
	"trading/api"
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

type ServerTimeJson struct {
	ServerTime int64 `json:"serverTime"`
}

type PriceJson *binance.SymbolPrice

func GetPriceAverage(symbol string) api.RequestResponse[PriceJson] {
	var bn = New[PriceJson](apiArg{Api: Endpoints.PriceAverage})
	params := map[string]string{"symbol": strings.ToUpper(symbol)}
	return bn.RequestWithQuery(params)
}

func GetPriceLatest(symbol string) float64 {
	if utils.Env().IsTest() {
		return 10
	}
	price, error := GetClient().NewListPricesService().Symbol(symbol).Do(context.Background())
	if error != nil {
		utils.LogError(error, "Get Price Latest %s")
		return 0
	}
	f, _ := strconv.ParseFloat(price[0].Price, 64)
	return f
}

func GetClient() *binance.Client {
	var secret = os.Getenv("API_SECRET")
	var key = os.Getenv("API_KEY")
	env := os.Getenv("ENV")
	binance.UseTestnet = env != "production"
	return binance.NewClient(key, secret)
}

func RequestChannel(chan int) {
	//
}

func GetSymbolPrices(symbols []string) (map[string]float64, error) {
	var postRunPrices = make(map[string]float64)
	prices, err := GetClient().NewListPricesService().Symbols(symbols).Do(context.Background())
	if err != nil {
		utils.LogError(err, "GetSymbolPrices()")
		return nil, err
	}
	for _, price := range prices {
		f, _ := strconv.ParseFloat(price.Price, 64)
		postRunPrices[price.Symbol] = f
	}
	return postRunPrices, err
}
