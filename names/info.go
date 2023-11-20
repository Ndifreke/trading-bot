package names

import (
	"context"
	"fmt"
	binLib "github.com/adshao/go-binance/v2"
	"trading/binance"
	"encoding/json"
	"trading/constant"
)

func loadInfoString() binLib.ExchangeInfo {
	var exchange binLib.ExchangeInfo
	json.Unmarshal([]byte(constant.ExchangeInfo), &exchange)
	return exchange
}

var exchangeInfo binLib.ExchangeInfo = binLib.ExchangeInfo{}

func LoadStoredExchangeInfo() binLib.ExchangeInfo {
	if exchangeInfo.ServerTime == 0 {
		data := loadInfoString()
		exchangeInfo = data
	}
	return exchangeInfo
}

type infoService struct {
	binLib.ExchangeInfo
}

func GetNewInfo() infoService {
	data, _ := binance.GetClient().NewExchangeInfoService().Do(context.Background())
	getInfo := func() binLib.ExchangeInfo {
		return *data
	}
	info := getInfo()
	return infoService{info}
}

func (info infoService) IsTrading(symbol Symbol) bool {
	s, err := info.findSymbol(symbol)
	if err != nil {
		return false
	}
	return s.Status == "TRADING"
}

func (info infoService) Spotable(symbol Symbol) bool {
	s, err := info.findSymbol(symbol)
	if err != nil {
		return false
	}
	isTrading := info.IsTrading(symbol)
	return isTrading && s.IsSpotTradingAllowed
}

func (info infoService) findSymbol(symbol Symbol) (binLib.Symbol, error) {
	for _, s := range info.Symbols {
		if s.Symbol == symbol.String() {
			return s, nil
		}
	}
	return binLib.Symbol{}, fmt.Errorf("no info for symbol '%s'", symbol.String())
}

func (info infoService) SpotableSymbol() []Symbol {
	spotable := []Symbol{}
	for _, s := range info.Symbols {
		symbol := Symbol(s.Symbol)
		if info.Spotable(symbol) {
			spotable = append(spotable, symbol)
		}
	}
	return spotable
}

func (info infoService) SpotableSymbolInfo() SymbolInfo {
	spotableSymbolInfo := []binLib.Symbol{}
	for _, s := range info.Symbols {
		symbol := Symbol(s.Symbol)
		if info.Spotable(symbol) {
			spotableSymbolInfo = append(spotableSymbolInfo, s)
		}
	}
	return NewSymbolInfo(spotableSymbolInfo)
}

// Filter symbols that can be spot traded
func (info infoService) FilterSpotable(symbol []Symbol) []Symbol {
	filter := []Symbol{}
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
