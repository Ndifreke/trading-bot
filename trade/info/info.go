package info

import (
	"context"
	"fmt"
	"trading/binance"
	"trading/names"
	binLib "github.com/adshao/go-binance/v2"
)

type infoService struct {
	binLib.ExchangeInfo
}

func GetNewInfo() infoService {
	getInfo := func() binLib.ExchangeInfo {
		data, _ := binance.GetClient().NewExchangeInfoService().Do(context.Background())
		return *data
	}
	info := getInfo()
	return infoService{info}
}

func (info infoService) IsTrading(symbol names.Symbol) bool {
	s, err := info.findSymbol(symbol)
	if err != nil {
		return false
	}
	return s.Status == "TRADING"
}

func (info infoService) Spotable(symbol names.Symbol) bool {
	s, err := info.findSymbol(symbol)
	if err != nil {
		return false
	}
	isTrading := info.IsTrading(symbol)
	return isTrading && s.IsSpotTradingAllowed
}

func (info infoService) findSymbol(symbol names.Symbol) (binLib.Symbol, error) {
	for _, s := range info.Symbols {
		if s.Symbol == symbol.String() {
			return s, nil
		}
	}
	return binLib.Symbol{}, fmt.Errorf("no info for symbol '%s'", symbol.String())
}

func (info infoService) SpotableSymbol() []names.Symbol {
	spotable := []names.Symbol{}
	for _, s := range info.Symbols {
		symbol := names.Symbol(s.Symbol)
		if info.Spotable(symbol) {
			spotable = append(spotable, symbol)
		}
	}
	return spotable
}

func (info infoService) SpotableSymbolInfo() names.SymbolInfo {
	spotableSymbolInfo := []binLib.Symbol{}
	for _, s := range info.Symbols {
		symbol := names.Symbol(s.Symbol)
		if info.Spotable(symbol) {
			spotableSymbolInfo = append(spotableSymbolInfo, s)
		}
	}
	return names.NewSymbolInfo(spotableSymbolInfo)
}

// Filter symbols that can be spot traded
func (info infoService) FilterSpotable(symbol []names.Symbol) []names.Symbol {
	filter := []names.Symbol{}
	for _, s := range symbol {
		if info.Spotable(s) {
			filter = append(filter, s)
		}
	}
	return filter
}

// func GetStoredInfo() infoService  {
// 	return infoService{
// 		binance.LoadStoredExchangeInfo(),
// 	}
// }
