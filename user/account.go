package user

import (
	// "os"
	"strconv"
	binanceLib "trading/binance"
	"trading/utils"

	binance "github.com/adshao/go-binance/v2"
)

type Accounter interface {
	GetBalance(asset string) Balance
	Account() *binance.Account
}

type Account struct {
	balances map[string]Balance
	account  *binance.Account
}

func GetAccount() *Account {
	if utils.Env().IsTest() {
		return GetMockAccount(mockAccount)
	}

	bals := map[string]Balance{}
	account := binanceLib.GetBinanceAccount()
	for _, b := range account.Balances {
		Locked, _ := strconv.ParseFloat(b.Locked, 64)
		Free, _ := strconv.ParseFloat(b.Free, 64)
		Asset := b.Asset
		bals[b.Asset] = Balance{
			Locked, Free, Asset,
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
