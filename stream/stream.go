package stream

import (
	"fmt"
	// "trading/constant"

	"trading/names"
	"trading/utils"
)

type ReaderFunc func(stream StreamInterface, data SymbolPriceData)

type SymbolPriceData struct {
	Price  float64
	Symbol string
}

type StreamInterface interface {
	Close() bool
	CloseLog(message string)
	IsClosed() bool
	RegisterLegacyReader(symbol string, broadcasterFunc ReaderFunc)
	RegisterBroadcast(readerId string, broadcasterFunc ReaderFunc)
	UnregisterBroadcast(readerId string) bool
	State() streamState
	RegisterFailOver(func(failedStream StreamInterface))
}

type StreamType string

var StreamTypeAPI StreamType = "STREAM_API"
var StreamTypeSocket StreamType = "STREAM_SOCKET"

type streamState struct {
	Readers    map[string]ReaderFunc
	Symbols    []string
	Type       StreamType
	BulkReader map[string]ReaderFunc
}

func GetPriceStreamer(symbols []string, useAPI bool) StreamInterface {
	if utils.Env().IsTest() {
		return NewMockStream(symbols)
	}
	if useAPI {
		return NewAPIStream(symbols)
	}
	return NewSocketStream(symbols)
}

type StreamManager struct {
	streamer StreamInterface
	Symbols  []string
}

func (sm *StreamManager) copystream(factory func(symbols []string) StreamInterface, s StreamInterface) StreamInterface {
	state := s.State()
	stream := factory(state.Symbols)
	stream.RegisterFailOver(sm.SwitchStream)
	for id, reader := range state.Readers {
		stream.RegisterLegacyReader(id, reader)
	}
	for id, reader := range state.BulkReader {
		stream.RegisterBroadcast(id, reader)
	}
	return stream
}

func (sm *StreamManager) SwitchStream(failedStream StreamInterface) {

	if failedStream == nil && sm.streamer != nil {
		failedStream = sm.streamer
	}
	state := failedStream.State()
	switch state.Type {
	case StreamTypeAPI:
		sm.streamer = sm.copystream(NewSocketStream, sm.streamer)
	case StreamTypeSocket:
		sm.streamer = sm.copystream(NewAPIStream, sm.streamer)
	default:
		utils.LogWarn(fmt.Sprintf("There was problem switching the stream %s", state.Type))
	}
}

func (sm *StreamManager) GetStream() StreamInterface {
	utils.LogInfo(fmt.Sprintf("Using Stream %s", " sm.streamer.State().Type"))
	return sm.streamer
}

func (sm *StreamManager) NewStream(symbols []string) StreamInterface {
	sm.streamer = GetPriceStreamer(symbols, !true)
	sm.streamer.RegisterFailOver(sm.SwitchStream)
	return sm.streamer
}

func (sm *StreamManager) StreamAll() StreamInterface {
	// return sm.NewStream(constant.SymbolList)
	v := names.GetNewInfo().SpotableSymbolInfo().List()
	utils.LogWarn(fmt.Sprintf("Loaded %d, only 1090 will be streamed", len(v)))
	return sm.NewStream(v[0:1090])
}

var Streamer = func() StreamInterface {
	s := StreamManager{}
	// s := StreamManager{Symbols: names.GetSymbols().List()}
	return s.StreamAll()
}()
