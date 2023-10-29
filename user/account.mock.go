package user

import (
	"github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
	"trading/names"
	"trading/utils"
)

func init() {
	godotenv.Load()
}

type AccountMock struct {
	balances map[string]Balance
	account  *binance.Account
}

func getMock() AccountMock {
	utils.LoadMyEnvFile()
	b := AccountMock{
		balances: map[string]Balance{
			"BTC":  {Free: utils.Env().BASE_BALANCE()},
			"USDT": {Free: utils.Env().QUOTE_BALANCE()},
			"BNB":  {Free: utils.Env().QUOTE_BALANCE()},
		},
	}
	return b
}

func CreateMockAccount(mock AccountMock) AccountInterface {
	return &AccountMock{
		balances: mock.balances,
		account:  mock.account,
	}
}

func (mock *AccountMock) UpdateLockBalance(asset string, locked float64) {
	if b, exists := mock.balances[asset]; !exists {
		mock.balances[asset] = Balance{
			Free:   locked,
			Locked: 0,
			Asset:  asset,
		}
	} else {
		mock.balances[asset] = Balance{
			Free:   locked,
			Locked: b.Locked,
		}
	}
}

func (mock *AccountMock) Account() *binance.Account {
	return mock.account
}

func (mock *AccountMock) GetBalance(asset string) Balance {
	return mock.balances[asset]
}

func (mock *AccountMock) UpdateFreeBalance(asset string, free float64) {
	if b, exists := mock.balances[asset]; !exists {
		mock.balances[asset] = Balance{
			Free:   0,
			Locked: free,
			Asset:  asset,
		}
	} else {
		mock.balances[asset] = Balance{
			Free:   b.Free,
			Locked: free,
		}
	}
}

func (mock *AccountMock) debit(asset string, cost float64) bool {
	balance := mock.GetBalance(asset).Free
	if cost > 0 && balance >= cost {
		mock.UpdateLockBalance(asset, balance-cost)
		return true
	}
	return false
}

func (mock *AccountMock) credit(asset string, quantity float64) bool {
	balance := mock.GetBalance(asset).Free
	if quantity > 0 {
		mock.UpdateLockBalance(asset, balance+quantity)
		return true
	}
	return false
}

func (mock *AccountMock) Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide) {
	commission := 0.01
	_ = commission
	// Check if the asset balance exists; create it if not
	baseAsset := symbol.Info().BaseAsset
	quoteAsset := symbol.Info().QuoteAsset

	// Calculate the total cost including commission

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
		totalCost := quantity * spot
		if mock.debit(quoteAsset, totalCost) {
			mock.credit(baseAsset, quantity)
		}
	} else if side == names.TradeSideSell {
		earnings := quantity * spot
		if mock.debit(baseAsset, quantity) {
			mock.credit(quoteAsset, earnings)
		}
	}
}

var MockAccount = CreateMockAccount(getMock())
