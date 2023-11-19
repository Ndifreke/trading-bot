package user

import (
	"fmt"
	"time"
	"trading/names"
	"trading/utils"

	binLib "github.com/adshao/go-binance/v2"
	"github.com/joho/godotenv"
)

func init() {
	godotenv.Load()
}

type AccountMock struct {
	balances map[string]Balance
	account  *binLib.Account
}

func getEnvBalance() map[string]float64 {
	env :=  utils.Env()
	b :=  map[string]float64{
			"BTC":  env.BASE_BALANCE(),
			"USDT": env.QUOTE_BALANCE(),
			"BNB": env.QUOTE_BALANCE(),
	}
	return b
}

func CreateMockBalance(balance map[string]float64) AccountMock {
	bb := make( map[string]Balance)
	for symbol,free := range balance{
		bb[symbol] = Balance{Free: free}
	}
	utils.LoadMyEnvFile()
	b := AccountMock{
		balances: bb,
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

func (mock *AccountMock) Account() *binLib.Account {
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

func (mock *AccountMock) debit(asset string, cost float64) (error, bool) {
	balance := mock.GetBalance(asset).Free
	if cost > 0 && balance >= cost {
		mock.UpdateLockBalance(asset, balance-cost)
		return nil, true
	}
	return fmt.Errorf("%s cost %f, or balance %f, error ", asset, cost, balance), false
}

func (mock *AccountMock) credit(asset string, quantity float64) (error, bool) {
	balance := mock.GetBalance(asset).Free
	if quantity > 0 {
		mock.UpdateLockBalance(asset, balance+quantity)
		return nil, true
	}
	return fmt.Errorf("invalid %s quantity %f is less than zero", asset, quantity), false
}

func (mock *AccountMock) Trade(quantity, spot float64, symbol names.Symbol, side names.TradeSide) (error, bool) {
	commission := 0.01
	_ = commission
	// Check if the asset balance exists; create it if not
	baseAsset := symbol.Info().BaseAsset
	quoteAsset := symbol.Info().QuoteAsset
	err, debited := fmt.Errorf("invalid trade side, must be BUY or SELL"), false
	
	if side == names.TradeSideBuy {
		totalCost := quantity * spot
		if err, debited = mock.debit(quoteAsset, totalCost); debited {
			return mock.credit(baseAsset, quantity)
		}
	} else if side == names.TradeSideSell {
		earnings := quantity * spot
		if err, debited = mock.debit(baseAsset, quantity); debited {
			return mock.credit(quoteAsset, earnings)
		}
	}

	return err, debited
}

func (mock *AccountMock) TradeBuyConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error) {
	symbol := config.Symbol
	quoteBalance := mock.GetBalance(symbol.ParseTradingPair().Quote)

	quantity := config.Buy.Quantity
	if quantity <= 0 {
		quantity = symbol.Quantity(quoteBalance.Free / spot)
	}
	if err, _ := mock.Trade(quantity, spot, symbol, names.TradeSideBuy); err != nil {
		utils.TextToSpeach("Buy error")
		utils.LogError(err, fmt.Sprintf(
			"Error  Buying %s,\n Supplied Qty=%f\n Calculated Qty=%f\n Quote Balance=%f", config.Symbol, config.Buy.Quantity, quantity, quoteBalance.Free))
		return &binLib.CreateOrderResponse{}, err
	}

	buyOrder := &binLib.CreateOrderResponse{
		Price:            fmt.Sprintf("%f", spot),
		OrigQuantity:     fmt.Sprintf("%f", quantity),
		ExecutedQuantity: fmt.Sprintf("%f", quantity),
		Type:             binLib.OrderTypeMarket,
		Status:           "FILLED",
		TransactTime:     time.Now().Unix(),
		Symbol:           symbol.String(),
		Side:             binLib.SideTypeBuy,
		OrderID:          123,
	}
	return buyOrder, nil
}

func (mock *AccountMock) TradeSellConfig(config names.TradeConfig, spot float64) (*binLib.CreateOrderResponse, error) {
	symbol := config.Symbol
	baseBalance := mock.GetBalance(symbol.ParseTradingPair().Base)
	quantity := config.Sell.Quantity

	if quantity <= 0 {
		quantity = config.Symbol.Quantity(baseBalance.Free)
	}

	if err, _ := mock.Trade(quantity, spot, symbol, names.TradeSideSell); err != nil {
		utils.TextToSpeach("sell error")
		utils.LogError(err, fmt.Sprintf("Error Selling %s, Qty=%f Balance=%f", symbol, quantity, baseBalance.Free))
		return &binLib.CreateOrderResponse{}, err
	}

	sellOrder := &binLib.CreateOrderResponse{
		Price:            fmt.Sprintf("%f", spot),
		OrigQuantity:     fmt.Sprintf("%f", quantity),
		ExecutedQuantity: fmt.Sprintf("%f", quantity),
		Type:             binLib.OrderTypeMarket,
		Status:           "FILLED",
		TransactTime:     time.Now().Unix(),
		Symbol:           symbol.String(),
		Side:             binLib.SideTypeSell,
		OrderID:          123,
	}
	return sellOrder, nil
}

var MockAccount = CreateMockAccount(CreateMockBalance(getEnvBalance()))
