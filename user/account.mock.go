package user

import (
	"trading/utils"

	"github.com/adshao/go-binance/v2"
)

type AccountMock struct {
	Balances map[string]Balance
	Account  *binance.Account
}

var mockAccount = AccountMock{

	Balances: map[string]Balance{
		"BTC":  {Locked: 2},
		"USDT": {Locked: utils.Env().QUOTE_BALANCE()},
	},
	Account: &binance.Account{},
}

func GetMockAccount(mock AccountMock) *Account {
	return &Account{
		balances: mock.Balances,
		account:  mock.Account,
	}
}

