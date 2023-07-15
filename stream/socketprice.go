package stream

import (
	"fmt"
	"strconv"
	"trading/utils"
	"github.com/adshao/go-binance/v2"
)

type Socket struct {
	readers         map[string]ReaderFunc
	getDataReaderId func(PriceStreamData) string
	symbols         []string
	stopChannel     chan struct{}
	doneChannel     chan struct{}
	failHandler     func(StreamInterface)
	streamIsClosed  bool
}

func NewSocketStream(symbols []string) StreamInterface {
	return &Socket{
		readers: make(map[string]ReaderFunc),
		symbols: symbols,
		getDataReaderId: func(data PriceStreamData) string {
			return data.Symbol
		},
	}
}

func readSocketDataDispatch(s *Socket) {
	if len(s.readers) < 1 {
		s.CloseLog("No API data reader, will close connection")
	}

	errorHandler := func(err error) {
		if s.failHandler != nil {
			s.CloseLog("An Error happened will close and fail over")
			s.failHandler(s)
		}
		s.CloseLog(err.Error())
	}

	messageHandler := func(event *binance.WsMarketStatEvent) {
		price, _ := strconv.ParseFloat(event.LastPrice, 64)
		data := PriceStreamData{Price: price, Symbol: event.Symbol}
		readerId := s.getDataReaderId(data)
		reader, isReader := s.readers[readerId]
		publishReader, isPublisher := s.readers[BROADCAST_ID]
		if isReader {
			go reader(s, data)
		}

		if isPublisher {
		go 	publishReader(s, data)
		}
	}

	donChannel, stopChannel, err := binance.WsCombinedMarketStatServe(
		s.symbols,
		messageHandler,
		errorHandler,
	)

	s.stopChannel = stopChannel
	s.doneChannel = donChannel

	if err != nil {
		errorHandler(err)
	}

}

func (s *Socket) RegisterReader(symbol string, reader ReaderFunc) {
	if len(s.readers) == 0 {
		//This is the first reader lets start the subscription
		go readSocketDataDispatch(s)
	}
	s.readers[symbol] = reader
}

func (s *Socket) Close() bool {
	s.stopChannel <- struct{}{}
	// s.doneChannel <- struct{}{}
	s.streamIsClosed = true
	return s.streamIsClosed
}

func (s *Socket) CloseLog(message string) {
	if s.streamIsClosed {
		utils.LogInfo(fmt.Sprintf("%s: Connection Closed", message))
	} else {
		utils.LogWarn(fmt.Sprintf("Connection is not Closed: `%s`", message))
	}
}

func (s *Socket) IsClosed() bool {
	return s.streamIsClosed
}

func (s *Socket) State() streamState {
	return streamState{
		Readers: s.readers,
		Symbols: s.symbols,
		Type:    StreamTypeSocket,
	}
}

func (s *Socket) RegisterFailOver(failHandler func(failedStream StreamInterface)) {
	s.failHandler = failHandler
}
