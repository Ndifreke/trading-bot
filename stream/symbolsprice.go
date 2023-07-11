package stream

import (
	"fmt"
	"time"
	"trading/binance"
	"trading/utils"
)

type Stream struct {
	symbols     []string
	readers     map[string]ReaderFunc
	closeStream bool
	failHandler func(StreamInterface)
}

func readApiDataDispatch(s *Stream) {

	if len(s.readers) < 1 {
		s.CloseLog("No socket data reader, will close connection")
	}
	for {
		if s.IsClosed() {
			return
		}

		symbols, err := binance.GetSymbolPrices(s.symbols)

		if err != nil && s.failHandler != nil {
			s.CloseLog("An Error happened will close and fail over")
			s.failHandler(s)
			break
		}
		for readerId, reader := range s.readers {
			Price := symbols[readerId]
			if readerId == BROADCAST_ID {
				// Broadcast should be handled seperatedl
				continue
			}
			func(reader func(StreamInterface, PriceStreamData), readerId string) {
				data := PriceStreamData{Price, readerId}
				reader(s, data)
			}(reader, readerId)

		}
		broadcaster, isBroadcast := s.readers[BROADCAST_ID]
		if isBroadcast {
			for readerId, Price := range symbols {
				data := PriceStreamData{Price, readerId}
				go broadcaster(s, data)
			}
		}
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

func (s *Stream) RegisterReader(id string, reader ReaderFunc) {
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
	}
}

func (s *Stream) CloseLog(message string) {
	closed := s.Close()
	if closed {
		utils.LogInfo(fmt.Sprintf("%s: Connection Closed", message))
	} else {
		utils.LogWarn(fmt.Sprintf("Connection is not Closed: `%s`", message))
	}
}

func (s *Stream) State() streamState {
	return streamState{
		Readers: s.readers,
		Symbols: s.symbols,
		Type:    StreamTypeAPI,
	}
}

func (s *Stream) RegisterFailOver(fh func(failedStream StreamInterface)) {
	s.failHandler = fh
}
