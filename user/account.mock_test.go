package user

import (
	"testing"
	"trading/names"
	"trading/utils"

	"github.com/stretchr/testify/assert"
)

func TestMockAccountInstance(t *testing.T) {
	utils.Env().SetModeMock()

	account := GetAccount()
	assert.Equal(t, account, MockAccount, "Mock account points to one instance")
	assert.Equal(t, account, GetAccount(), "GetAccount will always return the same instance of account")

}

func TestMockAccount(t *testing.T) {
	utils.Env().SetModeMock()

	var buy = names.TradeSideBuy
	var sell = names.TradeSideSell

	var mock = AccountMock{
		balances: map[string]Balance{
			"BTC":  {Free: 100},
			"USDT": {Free: 100},
		},
	}

	var account = CreateMockAccount(mock)

	account.Trade(10, 2, names.Symbol("BTCUSDT"), buy)
	assert.EqualValues(t, account.GetBalance("USDT").Free, 80, "Debit quote balance on buy")
	assert.EqualValues(t, account.GetBalance("BTC").Free, 110, "Credit base balance on buy")

	account.Trade(2, 10, names.Symbol("BTCUSDT"), sell)
	assert.EqualValues(t, account.GetBalance("USDT").Free, 100, "Debit quote balance on sell")
	assert.EqualValues(t, account.GetBalance("BTC").Free, 108, "Credit base balance on sell")
}
