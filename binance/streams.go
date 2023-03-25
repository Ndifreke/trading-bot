package binance

import (
	"fmt"
	"strings"
	"trading/request"
	"trading/trade"
)

const (
	streamAPI  = "wss://stream.binance.com:9443" //"wss://stream.binance.com:9443/stream?streams=btcusdt@ticker/ethusdt@ticker"
	streamProd = "wss://ws-api.binance.com/ws-api/v3"
	streamStg  = "wss://testnet.binance.vision/ws-api/v3"
	h          = "wss://testnet.binance.vision/ws"
)

type MiniTickerData struct {
	StreamName string `json:"stream"`
	Data       struct {
		EventType        string `json:"e"`
		EventTime        int64  `json:"E"`
		Symbol           string `json:"s"`
		ClosePrice       float32 `json:"c,string"`
		OpenPrice        float32 `json:"o,string"`
		HighPrice        string `json:"h"`
		LowPrice         string `json:"l"`
		Volume           string `json:"v"`
		QuoteAssetVolume string `json:"q"`
	} `json:"data"`
}

func PriceStream(symbols []trade.Symbol) *request.Socket[MiniTickerData] {
	stream := fmt.Sprintf("%s@miniTicker", strings.ToLower(symbols[0].String()))
	if len(symbols) > 1 {
		stream = ""
		for index, symbol := range symbols {
			if index == 1 {
				stream = fmt.Sprintf("%s/%s@miniTicker", stream,strings.ToLower( symbol.String()))
			} else {
				stream = fmt.Sprintf("%s%s@miniTicker", stream, strings.ToLower(symbol.String()))
			}
		}
	}

	var address = fmt.Sprintf("%s/stream?streams=%s", streamAPI, stream)
	return request.SocketConnection[MiniTickerData](address)
}



