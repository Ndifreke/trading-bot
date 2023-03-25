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
	conn *websocket.Conn
}

type Socket[T any] struct {
	conn            *websocket.Conn
	readers         map[string]func(conn Connection, message T)
	getDataReaderId func(data T) string
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
		conn:    s,
		readers: make(map[string]func(conn Connection, message T)),
	}
}

func (s *Socket[T]) Conn() *Connection {
	return &Connection{s.conn}
}

func (s *Socket[T]) ReadMessage(reader func(conn *Connection, message T)) {
	go readSocketData(s, reader)
	// defer s.Conn().Close()
	// for {
	// 	if s.Conn().IsClosed() {
	// 		return
	// 	}
	// 	var msg string
	// 	websocket.Message.Receive(s.Conn().Conn, &msg)
	// 	var data T
	// 	json.Unmarshal([]byte(msg), &data)
	// 	reader(s.Conn(), data)
	// }
}

func readSocketDataDispatch[D any](s *Socket[D]) {
	defer s.Conn().conn.Close()
	if len(s.readers) < 1 {
		s.Conn().CloseLog("No socket data reader, will close connection")
	}
	for {
		if s.Conn().IsClosed() {
			return
		}
		var msg string
		websocket.Message.Receive(s.Conn().conn, &msg)
		var data D
		json.Unmarshal([]byte(msg), &data)
		readerId := s.getDataReaderId(data)
		r, hasReader := s.readers[readerId]

		if hasReader {
			r(*s.Conn(), data)
		}
	}
}

func readSocketData[D any](s *Socket[D], reader func(conn *Connection, message D)) {
	defer s.Conn().conn.Close()
	for {
		if s.Conn().IsClosed() {
			return
		}
		var msg string
		websocket.Message.Receive(s.Conn().conn, &msg)
		var data D
		json.Unmarshal([]byte(msg), &data)

		reader(s.Conn(), data)
	}
}

func (s *Socket[T]) SetIdGetter(getter func(message T) string) *Socket[T] {
	s.getDataReaderId = getter
	return s
}

// Register all readers that are added to this Socket and start sending data to them
func (s *Socket[T]) SubscribeReaders() {
	// for i, reader := range s.readers {
	go readSocketDataDispatch(s)
	// }

	// defer s.Conn().Close()
	// for {
	// 	if s.Conn().IsClosed() {
	// 		return
	// 	}
	// 	var msg string
	// 	websocket.Message.Receive(s.Conn().Conn, &msg)
	// 	var data T
	// 	json.Unmarshal([]byte(msg), &data)
	// 	for _, reader := range s.readers {
	// 		reader(s.Conn(), data)
	// 	}
	// }
}

func (s *Socket[T]) RegisterReader(id string, reader func(conn Connection, message T)) {
	s.readers[id] = reader
}

func (s *Connection) Close() bool {
	closed := s.IsClosed()
	if !closed {
		s.conn.Close()
		closed = true
	}
	return closed
}

func (s *Connection) CloseLog(message string) {
	closed := s.Close()
	if closed {
		utils.LogInfo(fmt.Sprintf("%s: Connection Closed",message))
	} else {
		utils.LogWarn(fmt.Sprintf("Connection is not Closed: `%s`", message))
	}
}

func (s Connection) IsClosed() bool {
	var msg string
	err := websocket.Message.Receive(s.conn, &msg)
	if err != nil && strings.Contains(err.Error(), "closed") {
		return true
	}
	return err != nil && s.conn != nil
}
