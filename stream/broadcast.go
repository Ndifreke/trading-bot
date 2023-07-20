package stream

import (
	"fmt"
	"sync"
	"trading/helper"
	"trading/names"
	"trading/utils"
)

// Broadcaster handles the task of receiving price stream from
// One of the streamers and then publish those streams to the trader
// Every trader has an instance of broadcaster which provides methods
// to allow the trader manage that instance of broadcast
// When a trader is done with a broadcast it should destroy that broadcaster
type Broadcaster struct {
	subscribers map[names.TradeConfig]Subscription
	lock        sync.RWMutex
	streamer    StreamInterface
	broadcastId string
}

type Subscription struct {
	channel     chan SymbolPriceData
	tradeConfig names.TradeConfig
	broadcast   *Broadcaster
}

func (c *Subscription) Unsubscribe() bool {
	return c.broadcast.Unsubscribe(c.tradeConfig)
}

func (c *Subscription) State() struct{ TradingConfig names.TradeConfig} {
	return struct{TradingConfig names.TradeConfig}{
		TradingConfig: c.tradeConfig,
	}
}

func (c *Subscription) GetBroadCaster() *Broadcaster {
	return c.broadcast
}

func (c *Subscription) GetChannel() chan SymbolPriceData {
	return c.channel
}

func (ps *Broadcaster) readerReport(s map[names.TradeConfig]Subscription) {
	summary := fmt.Sprintf("Total Subscribers: %d\n", len(s))
	fmt.Println(summary)
}

var BROADCAST_ID = "BROADCAST_ID"

func NewBroadcast(broadcastId string) *Broadcaster {
	p := &Broadcaster{
		subscribers: map[names.TradeConfig]Subscription{},
		lock:        sync.RWMutex{},
		broadcastId: broadcastId,
		streamer:    Streamer,
	}

	// start this stream manager
	Streamer.RegisterBroadcast(broadcastId, func(stream StreamInterface, streamData SymbolPriceData) {
		p.publish(streamData.Symbol, streamData)
	})

	// lets keep track of the stream manager so we can use it to manager
	// our subscribers and possible all operations of stream manager
	// p.streaManager = Streamer
	// p.isRegistered = true
	return p
}

func (ps *Broadcaster) TerminateBroadCast() bool {
	cancelled := ps.streamer.UnregisterBroadcast(ps.broadcastId)
	if cancelled {
		utils.LogInfo(fmt.Sprintf("Cancelled Manager with ID %s", ps.broadcastId))
	}
	return cancelled
}

func (ps *Broadcaster) Subscribe(config names.TradeConfig) Subscription {

	for _, sub := range ps.subscribers {
		if sub.tradeConfig == config {
			return sub
		}
	}

	subscriber := Subscription{
		channel:     make(chan SymbolPriceData),
		tradeConfig: config,
		broadcast:   ps,
	}
	ps.lock.Lock()
	ps.subscribers[config] = subscriber
	ps.readerReport(ps.subscribers)
	ps.lock.Unlock()

	fmt.Printf("Subscribed %s %s Completed\n", config.Symbol, config.Side)
	return subscriber
}

// remove this trading config from the list of subscription
func (ps *Broadcaster) Unsubscribe(config names.TradeConfig) bool {
	var removed bool
	var totalBefore = len(ps.subscribers)

	ps.lock.Lock()
	if _, ok := ps.subscribers[config]; ok {
		delete(ps.subscribers, config)
		removed = true
	}
	ps.lock.Unlock()
	// subscription must not allow duplicate subscribers
	// or breaking will leave duplicates. If we must do so
	// then remove the break to  allow searching all the subscribers

	if removed {
		fmt.Printf("Removed %s:%s vs Before %d After %d %s",
			config.Symbol, config.Side, totalBefore, len(ps.subscribers), config.Side)
	} else {
		fmt.Printf("Could not remove %s:%s vs  Before %d After %s %s", config.Symbol, config.Side, totalBefore, helper.Stringify(ps.subscribers), config.Side)
	}
	ps.readerReport(ps.subscribers)
	return removed
}

func (ps *Broadcaster) UnsubscribeList(list []Subscription) {
	for _, sub := range list {
		go sub.Unsubscribe()
	}
}

func (ps *Broadcaster) publish(symbol string, symbolData SymbolPriceData) {
		for _, sub := range ps.subscribers {
			if sub.tradeConfig.Symbol.String() != symbol {
				continue
			}
			select {
			case sub.channel <- symbolData:
			default:
			}
		}
}
