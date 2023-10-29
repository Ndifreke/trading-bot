package user

import (
	"strconv"
	binanceLib "trading/binance"
	"trading/names"
	"trading/utils"

	binance "github.com/adshao/go-binance/v2"
)

type AccountInterface interface {
	GetBalance(asset string) Balance
	Account() *binance.Account
	Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide)
	UpdateLockBalance(asset string, quantity float64)
	UpdateFreeBalance(asset string, quantity float64)
}

type Account struct {
	balances map[string]Balance
	account  *binance.Account
}

func GetAccount() AccountInterface {
	if utils.Env().IsTest() {
		return MockAccount
	}

	bals := map[string]Balance{}
	account := binanceLib.GetBinanceAccount()
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

func (account *Account) Account() *binance.Account {
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

func (account *Account) Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide) {
	commission := 0.01

	// Check if the asset balance exists; create it if not
	baseAsset := symbol.Info().BaseAsset
	quoteAsset := symbol.Info().QuoteAsset

	// Calculate the total cost including commission
	totalCost := quantity * spot

	// Place the order
	// response, err := account.account.CreateOrder(binance.CreateOrderRequest{
	//     Symbol:      symbolStr,
	//     Side:        string(side),
	//     Type:        binance.OrderTypeMarket,
	//     Quantity:    quantity,
	//     Price:       price,
	//     TimeInForce: binance.TimeInForceGTC,
	// })

	// if err != nil {
	// 	// Handle the error
	// 	panic("Error placing order: " + err.Error())
	// }

	// Update balances based on the executed trade
	if side == names.TradeSideBuy {
		account.UpdateLockBalance(baseAsset, account.GetBalance(baseAsset).Free-quantity)
		account.UpdateLockBalance(quoteAsset, account.GetBalance(quoteAsset).Free-totalCost)
	} else if side == names.TradeSideSell {
		account.UpdateLockBalance(baseAsset, account.GetBalance(baseAsset).Free+quantity)
		account.UpdateLockBalance(quoteAsset, account.GetBalance(quoteAsset).Free+totalCost-commission)
	}
}
