package stream

import (
	"fmt"
	"trading/utils"
)

type ReaderFunc func(stream StreamInterface, data PriceStreamData)

type PriceStreamData struct {
	Price  float64
	Symbol string
}

type StreamInterface interface {
	Close() bool
	CloseLog(message string)
	IsClosed() bool
	RegisterReader(symbol string, reader ReaderFunc)
	State() streamState
	RegisterFailOver(func(failedStream StreamInterface))
}

type StreamType string

var StreamTypeAPI StreamType = "STREAM_API"
var StreamTypeSocket StreamType = "STREAM_SOCKET"

type streamState struct {
	Readers map[string]ReaderFunc
	Symbols []string
	Type    StreamType
}

func GetPriceStreamer(symbols []string, useAPI bool) StreamInterface {
	if useAPI {
		return NewAPIStream(symbols)
	}
	return NewSocketStream(symbols)
}

type StreamManager struct {
	streamer StreamInterface
}

func (sm *StreamManager) copystream(factory func(symbols []string) StreamInterface, s StreamInterface) StreamInterface {
	state := s.State()
	stream := factory(state.Symbols)
	stream.RegisterFailOver(sm.SwitchStream)
	for id, reader := range state.Readers {
		stream.RegisterReader(id, reader)
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
	utils.LogInfo(fmt.Sprintf("Using Stream %s", sm.streamer.State().Type))
	return sm.streamer
}

func (sm *StreamManager) NewStream(symbols []string) StreamInterface {
	sm.streamer = GetPriceStreamer(symbols, true)
	sm.streamer.RegisterFailOver(sm.SwitchStream)
	return sm.streamer
}
