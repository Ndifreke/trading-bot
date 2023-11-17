package user

import (
	"fmt"
	"strconv"
	"trading/binance"
	"trading/names"
	"trading/utils"

	binLib "github.com/adshao/go-binance/v2"
)

type AccountInterface interface {
	GetBalance(asset string) Balance
	Account() *binLib.Account
	Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide) (error, bool)
	UpdateLockBalance(asset string, quantity float64)
	UpdateFreeBalance(asset string, quantity float64)
	TradeBuyConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error)
	TradeSellConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error)
}

type Account struct {
	balances map[string]Balance
	account  *binLib.Account
}

func GetAccount() AccountInterface {
	if utils.Env().UseMockAccount()  {
		return MockAccount
	}

	bals := map[string]Balance{}
	account := binance.GetBinanceAccount()
	for _, b := range account.Balances {
		Locked, _ := strconv.ParseFloat(b.Locked, 64)
		Free, _ := strconv.ParseFloat(b.Free, 64)
		Asset := b.Asset
		bals[b.Asset] = Balance{
			Locked: Locked,
			Free:   Free,
			Asset:  Asset,
		}
	}
	return &Account{
		balances: bals,
		account:  account,
	}
}

func (account *Account) GetBalance(asset string) Balance {
	return account.balances[asset]
}

func (account *Account) Account() *binLib.Account {
	return account.account
}

func (account *Account) UpdateLockBalance(asset string, quantity float64) {
	if b, exists := account.balances[asset]; !exists {
		account.balances[asset] = Balance{
			Free:   quantity,
			Locked: 0,
			Asset:  asset,
		}
	} else {
		account.balances[asset] = Balance{
			Free:   quantity,
			Locked: b.Locked,
		}
	}
}

func (account *Account) UpdateFreeBalance(asset string, quantity float64) {
	if b, exists := account.balances[asset]; !exists {
		account.balances[asset] = Balance{
			Free:   0,
			Locked: quantity,
			Asset:  asset,
		}
	} else {
		account.balances[asset] = Balance{
			Free:   b.Free,
			Locked: quantity,
		}
	}
}

func (account *Account) Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide) (error, bool) {
	panic("Not implemented for production account")
}

func (account *Account) TradeBuyConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error) {
	symbol := config.Symbol
	quoteBalance := account.GetBalance(symbol.ParseTradingPair().Quote)
	quantity := config.Buy.Quantity

	if quantity <= 0 {
		quantity = symbol.Quantity(quoteBalance.Free / spot)
	}
	buyOrder, err := binance.CreateBuyMarketOrder(symbol.String(), quantity)

	if err != nil {
		utils.TextToSpeach("Buy error")
		utils.LogError(err, fmt.Sprintf(
			"Error  Buying %s,\n Supplied Qty=%f\n Calculated Qty=%f\n Quote Balance=%f", config.Symbol, config.Buy.Quantity, quantity, quoteBalance.Free))
	}

	return buyOrder, err
}

func (account *Account) TradeSellConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error) {
	symbol := config.Symbol
	baseBalance := account.GetBalance(symbol.ParseTradingPair().Base)
	quantity := config.Sell.Quantity

	if quantity <= 0 {
		quantity = config.Symbol.Quantity(baseBalance.Free)
	}

	sellOrder, err := binance.CreateSellMarketOrder(symbol.String(), quantity)

	if err != nil {
		utils.TextToSpeach("sell error")
		utils.LogError(err, fmt.Sprintf("Error Selling %s, Qty=%f Balance=%f", symbol, quantity, baseBalance.Free))
	}

	return sellOrder, err
}
