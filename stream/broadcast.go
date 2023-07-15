package stream

import (
	"fmt"
	"strings"
	"sync"
	"trading/names"
)

type broadcast struct {
	subscribers map[string][]chan PriceStreamData
	lock        sync.RWMutex
}

func (ps *broadcast) readerReport(s map[string][]chan PriceStreamData) {
	var summary strings.Builder
	totalItems := 0
	// ps.lock.Lock()
	for key, channels := range s {
		numChannels := len(channels)
		totalItems += numChannels

		summary.WriteString(fmt.Sprintf("%s, Number of Channels: %d\n", key, numChannels))
	}
	// ps.lock.Unlock()
	summary.WriteString(fmt.Sprintf("Total Items: %d\n", totalItems))
	fmt.Println(summary.String())
}

var BROADCAST_ID = "BROADCAST_ID"

func NewBroadcast(source StreamManager) *broadcast {
	p := &broadcast{
		subscribers: make(map[string][]chan PriceStreamData),
		lock:        sync.RWMutex{},
	}
	source.StreamAll().RegisterReader(BROADCAST_ID, func(stream StreamInterface, data PriceStreamData) {
		// p.lock.Lock()
		// defer p.lock.Unlock()
		p.publish(data.Symbol, data)

	})
	return p
}

func (ps *broadcast) Subscribe(id string) chan PriceStreamData {
	subscriber := make(chan PriceStreamData)

	ps.lock.Lock()
	defer ps.lock.Unlock()
	subscribers, ok := ps.subscribers[id]
	if ok {
		ps.subscribers[id] = append(subscribers, subscriber)
	} else {
		ps.subscribers[id] = []chan PriceStreamData{subscriber}
	}

	ps.readerReport(ps.subscribers)

	return subscriber
}

func (ps *broadcast) Unsubscribe(id string, subscription chan PriceStreamData) {
	// ps.lock.Lock()
	// defer ps.lock.Unlock()
	if subscribers, ok := ps.subscribers[id]; ok {
		for i, s := range subscribers {
			if s == subscription {
				ps.lock.Lock()
				ps.subscribers[id] = append(subscribers[:i], subscribers[i+1:]...)
				ps.lock.Unlock()
				break
			}
		}
	}
	ps.readerReport(ps.subscribers)
}

type SubscriptionChan struct {
	Id           string
	Subscription chan PriceStreamData
}

func (ps *broadcast) UnsubscribeList(list []SubscriptionChan) {
	for _, sub := range list {
		go ps.Unsubscribe(sub.Id, sub.Subscription)
	}
}

func (ps *broadcast) publish(pubId string, message PriceStreamData) {

	ps.lock.Lock()
	subscribers := ps.subscribers[pubId]
	waitGroup := sync.WaitGroup{}

	waitGroup.Add(len(subscribers))
	ps.lock.Unlock()
	for _, subscribe := range subscribers {

		go func(ch chan PriceStreamData) {
			select {
			case ch <- message:
			default:
			}
			defer waitGroup.Done()
		}(subscribe)
	}

	waitGroup.Wait()
}

var Broadcaster = NewBroadcast(StreamManager{Symbols: names.GetSymbols().List()})
