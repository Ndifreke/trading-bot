package executor

import (
	"fmt"
	// "time"
	tradeBinance "trading/binance"
	"trading/helper"
	"trading/names"
	"trading/user"
	"trading/utils"

	binance "github.com/adshao/go-binance/v2"
)

type sellExecutor executorType

func SellExecutor(
	config names.TradeConfig,
	marketPrice float64,
	tradeStartPrice float64,
) ExecutorInterface {

	return &sellExecutor{
		marketPrice,
		tradeStartPrice,
		config,
		helper.TradeFee{},
	}
}

func (exec *sellExecutor) IsProfitable() bool {
	// We can look at the trade history and get the last trade
	// for symbol if action == Sell last trade == symbol.lastBuy
	// if does not exist then assume asset must have been transfered in
	if !exec.config.Sell.MustProfit || exec.tradeStartPrice == 0 {
		return true
	}
	return (exec.marketPrice - exec.tradeStartPrice) > exec.fees.Value*2
}

func sell(st *sellExecutor) bool {
	lastTradePrice := st.tradeStartPrice
	baseBalance := user.CreateUser().GetAccount().GetBalance(st.config.Symbol.ParseTradingPair().Base)

	quantity := st.config.Sell.Quantity
	if quantity <= 0 {
		quantity = st.config.Symbol.Quantity(baseBalance.Locked)
	}

	sellOrder := &binance.CreateOrderResponse{}

	if utils.Env().IsTest() {
		summary(
			st.config,
			st.config.Side,
			st.config.Symbol,
			lastTradePrice,
			st.tradeStartPrice,
			st.marketPrice,
			st.marketPrice-lastTradePrice,
			st.fees,
			st.config.Sell.Quantity,
			*sellOrder,
		)
		return true
	}

	var err error
	sellOrder, err = tradeBinance.CreateSellMarketOrder(
		st.config.Symbol.String(),
		quantity,
	)

	if err != nil {
		utils.LogError(err, fmt.Sprintf("Error Selling %s, Qty=%f Balance=%f", st.config.Symbol, quantity, baseBalance.Locked))
		return false
	}
	summary(
		st.config,
		st.config.Side,
		st.config.Symbol,
		lastTradePrice,
		st.tradeStartPrice,
		st.marketPrice,
		st.marketPrice-lastTradePrice,
		st.fees,
		st.config.Sell.Quantity,
		*sellOrder,
	)
	return true
}

func (exec *sellExecutor) Execute() bool {
	sold := sell(exec)
	if sold {
		// Pause for 1 second for the server executing server to complete the order
		// TODO it would be nice to recursive check if this is filled
		// time.Sleep(10 * time.Second)
	}
	return sold
}
