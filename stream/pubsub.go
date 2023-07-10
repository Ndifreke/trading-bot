package stream

import (
	"sync"
)

type PubSub struct {
	subscribers map[string][]chan any
}

func NewPubSub(source StreamManager) *PubSub {
	p := &PubSub{
		subscribers: make(map[string][]chan any, 0),
	}
	source.NewStream(nil).RegisterReader("", func(stream StreamInterface, data PriceStreamData) {
		p.publish(data.Symbol, data)
	})
	return p
}

func (ps *PubSub) Subscribe(id string) chan any {
	subscriber := make(chan any)
	subscribers, ok := ps.subscribers[id]
	if ok {
		ps.subscribers[id] = append(subscribers, subscriber)
	} else {
		ps.subscribers[id] = append(make([]chan any, 0), subscriber)
	}
	return subscriber
}

func (ps *PubSub) Unsubscribe(id string) {
	delete(ps.subscribers, id)
}

func (ps *PubSub) publish(pubId string, message any) {
	subscribers := ps.subscribers[pubId]
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(len(subscribers))

	for _, subscribe := range subscribers {
		go func(ch chan<- any) {
			ch <- message
			//remove from wg when received
			defer waitGroup.Done()
		}(subscribe)
	}

	waitGroup.Wait()
}
