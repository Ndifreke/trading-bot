package traders

import (
	"testing"
	"trading/names"
	"trading/user"
	"trading/utils"

	"github.com/stretchr/testify/assert"
)

func TestNewAssetGainsLock(t *testing.T) {
	// mockAccount := user.AccountMock{
	// 	Balances: map[string]user.Balance{"BTC": {Locked: 2}, "USDT": {Locked: 50}},
	// 	Account:  &binance.Account{},
	// }
	utils.Env().SetModeMock()

	BTCUSDT := names.TradeConfig{
		Side: names.TradeSideBuy,
		Buy: names.SideConfig{
			LimitType: names.RatePercent,
			StopLimit: 1,
			Quantity:  names.MAX_QUANTITY,
		},
		Sell: names.SideConfig{
			Quantity:  names.MAX_QUANTITY,
			LimitType: names.RatePercent,
			StopLimit: 1,
		},
		Symbol: "BTCUSDT",
	}
	configs := []names.TradeConfig{BTCUSDT}
	spot := map[string]float64{"BTCUSDT": 238.6}
	spotFees := map[string]float64{"BTCUSDT": 0.001}

	//Will return a mock account instance
	account := user.GetAccount()
	// updatedConfig := updateStopPriceWithStableLimit(config, 30000, 50, 0.1)
	updatedConfigs := calculateConfigPeggedLimit(configs, spot, spotFees, account)

	assert.EqualValues(t, updatedConfigs[0].Buy.StopLimit, 4.000166666666664, "should calculate the percent buy position to gain tradeLimit of 2USDT value provided in Buy using account balance of 50USDT as deposit")

	assert.EqualValues(t, updatedConfigs[0].Sell.StopLimit, 0.40166666666666667, "should calculate the percent sell position to realise tradeLimit 4 USDT value provided in sell using account balance of 4 BTC as deposit")

}
