package names

import (
	"context"
	"strconv"
	"sync"
	"trading/binance"
	"trading/utils"

	binLib "github.com/adshao/go-binance/v2"
)

type SymbolInfo struct {
	symbols []binLib.Symbol
}

func NewSymbolInfo(symbols []binLib.Symbol) SymbolInfo {
	return SymbolInfo{symbols}
}

func GetStoredInfo() SymbolInfo {
	return SymbolInfo{
		symbols: LoadStoredExchangeInfo().Symbols,
	}
}

func (smb SymbolInfo) ToPair(symbol string) TradingPair {
	for _, s := range smb.symbols {
		if s.Symbol == symbol {
			return TradingPair{
				Quote: s.QuoteAsset,
				Base:  s.BaseAsset,
			}
		}
	}
	return TradingPair{}
}

func (smb SymbolInfo) List() []string {
	var list []string
	for _, s := range smb.symbols {
		list = append(list, s.Symbol)
	}
	return list
}

type tradeFeeDetails struct {
	Symbol          string
	MakerCommission float64
	TakerCommission float64
}

func GetTradeFees(symbols []string) map[string]tradeFeeDetails {
	symbolFees := make(map[string]tradeFeeDetails)
	var wg sync.WaitGroup
	rwLock := sync.RWMutex{}

	wg.Add(len(symbols))
	for _, symbol := range symbols {
		go func(s string) {
			if utils.Env().IsMock() {

				rwLock.Lock()

				symbolFees[s] = tradeFeeDetails{
					Symbol:          s,
					MakerCommission: 0.001,
					TakerCommission: 0.001,
				}
				rwLock.Unlock()

				wg.Done()
				return
			}

			data, err := binance.GetClient().NewTradeFeeService().Symbol(s).Do(context.Background())
			if err != nil {
				utils.LogError(err, "Could not get trading fees")
				wg.Done()
				return
			}
			MakerCommission, _ := strconv.ParseFloat(data[0].MakerCommission, 64)
			TakerCommission, _ := strconv.ParseFloat(data[0].TakerCommission, 64)

			rwLock.Lock()

			symbolFees[s] = tradeFeeDetails{
				Symbol:          data[0].Symbol,
				MakerCommission: MakerCommission,
				TakerCommission: TakerCommission,
			}

			rwLock.Unlock()

			wg.Done()
		}(symbol)
	}
	wg.Wait()
	return symbolFees
}
