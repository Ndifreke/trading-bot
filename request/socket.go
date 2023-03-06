package request

import (
	"encoding/json"
	"fmt"
	"strings"
	"trading/utils"

	"golang.org/x/net/websocket"
)

const (
	streamAPI  = "wss://stream.binance.com:9443" //"wss://stream.binance.com:9443/stream?streams=btcusdt@ticker/ethusdt@ticker"
	streamProd = "wss://ws-api.binance.com/ws-api/v3"
	streamStg  = "wss://testnet.binance.vision/ws-api/v3"
	h          = "wss://testnet.binance.vision/ws"
)

var symbol = "btcbusd"
var interval = "1m"
var ur = fmt.Sprintf("%s/%s@kline_%s", h, symbol, interval)


type Connection struct {
	*websocket.Conn
}

type Socket[T any] struct {
	*websocket.Conn
}

// func NewSocket(address string) *Socket {
// 	s, err := websocket.Dial(ur, "", ur)
// 	if err != nil {
// 		panic(err)
// 	}
// 	fmt.Printf("Connected to %s", s.LocalAddr())
// 	// b := make([]byte, 1024)
// 	// for {
// 	// i, _ := s.Read(b)
// 	// fmt.Println(string(b[:i]))
// 	var msg string
// 	websocket.Message.Receive(s, &msg)
// 	fmt.Print(msg)
// 	// }
// 	return &Socket{
// 		connection: s,
// 	}
// }

func SocketConnection[T any](address string) *Socket[T] {
	s, err := websocket.Dial(address, "", address)
	if err != nil {
		fmt.Print(err)
	}
	utils.LogInfo(fmt.Sprintf("Connected to %s", s.LocalAddr()))
	var msg string
	websocket.Message.Receive(s, &msg)
	return &Socket[T]{
		s,
	}
}

func (s Socket[T]) ReadMessage(reader func(conn Connection, message T)) {

	defer s.Close()
	for {
		var msg string
		err := websocket.Message.Receive(s.Conn, &msg)

		if err != nil && strings.Contains(err.Error(), "closed") {
			return
		}
		var data T
		json.Unmarshal([]byte(msg), &data)
		reader(Connection(s), data)
	}
}

func (s Connection) Close() {
	if !s.IsClosed() {
		s.Close()
	}
}

func (s Connection) IsClosed() bool {
	var msg string
	err := websocket.Message.Receive(s.Conn, &msg)
	if err != nil && strings.Contains(err.Error(), "closed") {
		return true
	}
	return err != nil && s.Conn != nil
}
