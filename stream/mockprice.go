package stream

import (
	"fmt"
	"sync"
	"time"
	"trading/utils"
)

type MockPrice struct {
	symbols     []string
	readers     map[string]ReaderFunc
	closeStream bool
	failHandler func(StreamInterface)
	bulkReaders map[string]ReaderFunc
	lock        sync.RWMutex
}

type Price struct {
	Iteration int
	Prices    []float64
}
type PriceMap map[string]Price

type mockGen struct {
	priceMap PriceMap
}

func (pm *mockGen) GetNextPrice(symbol string) float64 {
	value := pm.priceMap[symbol]
	iteration := value.Iteration
	priceList := value.Prices
	if len(priceList) == iteration {
		iteration = 0
	}
	price := priceList[iteration]
	value.Iteration = iteration + 1
	pm.priceMap[symbol] = value
	return price
}

func (pm *mockGen) GetNextPrice2() map[string]float64 {
	symbolPrices := make(map[string]float64)
	for symb, _ := range pm.priceMap {
		symbolPrices[symb] = pm.GetNextPrice(symb)
	}
	return symbolPrices
}

func (pm *mockGen) NextPrices() map[string]float64 {

	symbolPrices := make(map[string]float64)

	for symb, value := range pm.priceMap {
		iteration := value.Iteration
		priceList := value.Prices

		if len(priceList) == iteration {
			iteration = 0
		}

		symbolPrices[symb] = priceList[iteration]
		value.Iteration = iteration + 1
		pm.priceMap[symb] = value
	}
	return symbolPrices
}

func newMockPrice(symbols []string, priceSource map[string][]float64) *mockGen {
	priceMap := make(PriceMap)
	for _, symbol := range symbols {
		priceMap[symbol] = Price{Iteration: 0, Prices: priceSource[symbol]}
	}

	return &mockGen{
		priceMap: priceMap,
	}
}

func getMockPriceSource(symbol []string) map[string][]float64 {
	max := utils.Env().MaxInt()
	// min := utils.Env().MinInt()
	// sell := utils.Env().SellTrue()
	prices := []float64{max}
	for i := 1.0; i < 4; i++ {
		prices = append(prices, max+(i*2))
	}
	// prices = append(prices, prices[len(prices)-1]-1)
	prices = append(prices, max - 2)

	// For Buy Trigger deep below buy limit and rise again but just not enough
	prices = append(prices, max - 3)
	prices = append(prices, max - 3 - 4 + 2)
	prices = append(prices, max - 4)


	symbolPrices := make(map[string][]float64)
	for _, s := range symbol {
		symbolPrices[s] = prices
	}
	return symbolPrices
}

func readTesterDataDispatch(s *MockPrice) {
	mock := newMockPrice(s.symbols, getMockPriceSource(s.symbols))
	// type Price struct {
	// 	Iteration int
	// 	Prices    []float64
	// }
	// type PriceMap map[string]Price

	// symbolPriceMap := make(PriceMap)

	// for _, symbol := range s.symbols {
	// 	symbolPriceMap[symbol] = Price{Iteration: 0, Prices: []float64{}}
	// }

	// getSymbolPrices := func(priceMap *PriceMap) map[string]float64 {
	// 	symbolPrices := make(map[string]float64)

	// 	for symb, value := range *priceMap {
	// 		iteration := value.Iteration
	// 		priceList := value.Prices

	// 		if len(priceList) == iteration {
	// 			iteration = 0
	// 		}

	// 		symbolPrices[symb] = priceList[iteration]
	// 		value.Iteration = iteration + 1
	// 		(*priceMap)[symb] = value
	// 	}
	// 	return symbolPrices
	// }

	if len(s.readers) < 1 {
		utils.LogInfo("<Mock Tester>: No socket data reader")
	}
	for {
		if s.IsClosed() {
			return
		}

		symbols := mock.NextPrices()
		// symbols, err := binance.GetSymbolPrices(s.symbols)

		if false && s.failHandler != nil {
			s.CloseLog(fmt.Sprintf("<Mock Tester>: An Error happened will close and fail over"))
			s.failHandler(s)
			break
		}
		go func(symbols map[string]float64) {
			s.lock.RLock()

			for readerId, Price := range symbols {
				for _, bulkReader := range s.bulkReaders {
					data := SymbolPriceData{Price, readerId}
					go bulkReader(s, data)
				}
			}

			s.lock.RUnlock()
		}(symbols)

		time.Sleep(1 * time.Second)
	}
}

func (s *MockPrice) Close() bool {
	s.closeStream = true
	return s.IsClosed()
}

func (s *MockPrice) IsClosed() bool {
	return s.closeStream
}

func (s *MockPrice) RegisterLegacyReader(id string, reader ReaderFunc) {
	if len(s.readers) == 0 {
		s.readers[id] = reader

		//start dispatching immediately a reader is added
		go readTesterDataDispatch(s)
	} else {
		s.readers[id] = reader
	}
}

func NewMockStream(symbols []string) StreamInterface {
	return &MockPrice{
		closeStream: false,
		symbols:     symbols,
		readers:     make(map[string]ReaderFunc),
		bulkReaders: map[string]ReaderFunc{},
		lock:        sync.RWMutex{},
	}
}

func (s *MockPrice) CloseLog(message string) {
	closed := s.Close()
	if closed {
		utils.LogInfo(fmt.Sprintf("<API Tester>: %s: Connection Closed", message))
	} else {
		utils.LogWarn(fmt.Sprintf("<API Tester>: Connection is not Closed: `%s`", message))
	}
}

func (s *MockPrice) RegisterBroadcast(readerId string, reader ReaderFunc) {
	if len(s.readers) == 0 && len(s.bulkReaders) == 0 {
		go readTesterDataDispatch(s)
	}
	s.lock.Lock()
	defer s.lock.Unlock()
	s.bulkReaders[readerId] = reader
}

func (s *MockPrice) UnregisterBroadcast(readerId string) bool {
	s.lock.Lock()
	defer s.lock.Unlock()
	_, exist := s.bulkReaders[readerId]
	delete(s.bulkReaders, readerId)
	return exist
}

func (s *MockPrice) State() streamState {
	return streamState{
		Readers:    s.readers,
		Symbols:    s.symbols,
		Type:       StreamTypeAPI,
		BulkReader: s.bulkReaders,
	}
}

func (s *MockPrice) RegisterFailOver(fh func(failedStream StreamInterface)) {
	s.failHandler = fh
}
