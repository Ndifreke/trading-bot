package user

import (
	"strconv"
	"trading/binance"
)

type Account struct {
	balances map[string]Balance
}

func GetAccount() *Account {
	bals := map[string]Balance{}
	 for _, b := range binance.GetBinanceAccount().Balances {
		Locked, _ := strconv.ParseFloat(b.Locked, 64)
		Free, _ := strconv.ParseFloat(b.Free, 64)
		Asset := b.Asset
		bals[b.Asset] = Balance{
			Locked, Free, Asset,
		}
	}
	return &Account{
		balances: bals,
	}
}

func (account *Account) GetBalance(id string) Balance {
	return account.balances[id]
}
