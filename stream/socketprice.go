package stream

import (
	"fmt"
	"strconv"
	"sync"
	"trading/utils"
	"github.com/adshao/go-binance/v2"
)

type Socket struct {
	readers         map[string]ReaderFunc
	bulkReaders     map[string]ReaderFunc //Maybe we should change to array to avoid cuncurrency
	getDataReaderId func(SymbolPriceData) string
	symbols         []string
	stopChannel     chan struct{}
	doneChannel     chan struct{}
	failHandler     func(StreamInterface)
	streamIsClosed  bool
	lock            sync.RWMutex
}

func NewSocketStream(symbols []string) StreamInterface {

	return &Socket{
		lock:        sync.RWMutex{},
		readers:     make(map[string]ReaderFunc),
		symbols:     symbols,
		bulkReaders: map[string]ReaderFunc{},
		getDataReaderId: func(data SymbolPriceData) string {
			return data.Symbol
		},
	}
}

func readSocketDataDispatch(s *Socket) {

	if len(s.bulkReaders) == 0 {
		utils.LogInfo("<Socket Stream>: No API data reader found")
	}

	errorHandler := func(err error) {
		if s.failHandler != nil {
			s.CloseLog(fmt.Sprintf("<Socket Stream>: An Error happened will close and fail over %s", err.Error()))
			s.failHandler(s)
		}
		s.CloseLog(err.Error())
	}

	messageHandler := func(event *binance.WsMarketStatEvent) {
		price, _ := strconv.ParseFloat(event.LastPrice, 64)
		data := SymbolPriceData{Price: price, Symbol: event.Symbol}

		go func(data SymbolPriceData) {
			for _, bulkReader := range s.bulkReaders {
				go bulkReader(s, data)
			}
		}(data)

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

func (s *Socket) RegisterBroadcast(key string, reader ReaderFunc) {
	if len(s.readers) == 0 && len(s.bulkReaders) == 0 {
		go readSocketDataDispatch(s)
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.bulkReaders[key] = reader
}

func (s *Socket) UnregisterBroadcast(id string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exist := s.bulkReaders[id]
	delete(s.bulkReaders, id)
	return exist
}

func (s *Socket) RegisterLegacyReader(symbol string, reader ReaderFunc) {
	if len(s.readers) == 0 && len(s.bulkReaders) == 0 {
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
		utils.LogInfo(fmt.Sprintf("<Socket Stream>: %s: Connection Closed", message))
	} else {
		utils.LogWarn(fmt.Sprintf("<Socket Stream>: Connection is not Closed: `%s`", message))
	}
}

func (s *Socket) IsClosed() bool {
	return s.streamIsClosed
}

func (s *Socket) State() streamState {
	return streamState{
		Readers:    s.readers,
		BulkReader: s.bulkReaders,
		Symbols:    s.symbols,
		Type:       StreamTypeSocket,
	}
}

func (s *Socket) RegisterFailOver(failHandler func(failedStream StreamInterface)) {
	s.failHandler = failHandler
}
