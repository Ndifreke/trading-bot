package stream

import (
	"fmt"
	"sync"
	"time"
	"trading/binance"
	"trading/utils"
)

type Stream struct {
	symbols     []string
	readers     map[string]ReaderFunc
	closeStream bool
	failHandler func(StreamInterface)
	bulkReaders map[string]ReaderFunc
	lock        sync.RWMutex
}

func readApiDataDispatch(s *Stream) {

	if len(s.readers) < 1 {
		utils.LogInfo("<API Stream>: No socket data reader")
	}
	for {
		if s.IsClosed() {
			return
		}

		symbols, err := binance.GetSymbolPrices(s.symbols)

		if err != nil && s.failHandler != nil {
			s.CloseLog(fmt.Sprintf("<API Stream>: An Error happened will close and fail over: %s", err.Error()))
			s.failHandler(s)
			break
		}
		for readerId, reader := range s.readers {
			Price := symbols[readerId]
			if readerId == BROADCAST_ID {
				// Broadcast should be handled seperatedl
				continue
			}
			func(reader func(StreamInterface, SymbolPriceData), readerId string) {
				data := SymbolPriceData{Price, readerId}
				reader(s, data)
			}(reader, readerId)

		}
		go func(symbols map[string]float64) {
			for readerId, Price := range symbols {
				for _, bulkReader := range s.bulkReaders {
					data := SymbolPriceData{Price, readerId}
					go bulkReader(s, data)
				}
			}
		}(symbols)

		time.Sleep(1 * time.Second)
	}
}

func (s *Stream) Close() bool {
	s.closeStream = true
	return s.IsClosed()
}

func (s *Stream) IsClosed() bool {
	return s.closeStream
}

func (s *Stream) RegisterLegacyReader(id string, reader ReaderFunc) {
	if len(s.readers) == 0 {
		s.readers[id] = reader
		//start dispatching immediately a reader is added
		go readApiDataDispatch(s)
	} else {
		s.readers[id] = reader
	}
}

func NewAPIStream(symbols []string) StreamInterface {
	return &Stream{
		closeStream: false,
		symbols:     symbols,
		readers:     make(map[string]ReaderFunc),
		bulkReaders: map[string]ReaderFunc{},
		lock:        sync.RWMutex{},
	}
}

func (s *Stream) CloseLog(message string) {
	closed := s.Close()
	if closed {
		utils.LogInfo(fmt.Sprintf("<API Stream>: %s: Connection Closed", message))
	} else {
		utils.LogWarn(fmt.Sprintf("<API Stream>: Connection is not Closed: `%s`", message))
	}
}

func (s *Stream) RegisterBroadcast(readerId string, reader ReaderFunc) {
	if len(s.readers) == 0 && len(s.bulkReaders) == 0 {
		go readApiDataDispatch(s)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.bulkReaders[readerId] = reader
}

func (s *Stream) UnregisterBroadcast(readerId string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exist := s.bulkReaders[readerId]
	delete(s.bulkReaders, readerId)
	return exist
}

func (s *Stream) State() streamState {
	return streamState{
		Readers: s.readers,
		Symbols: s.symbols,
		Type:    StreamTypeAPI,
		BulkReader: s.bulkReaders,
	}
}

func (s *Stream) RegisterFailOver(fh func(failedStream StreamInterface)) {
	s.failHandler = fh
}
