package bus

import (
	"sync"
)

type (
	Bus interface {
		Publish(topic string, data interface{})
		Subscribe(topic string, receiver chan Event)
	}

	// 事件
	Event struct {
		Topic string      // 事件主题
		Data  interface{} // 事件数据
	}

	bus struct {
		subscribers map[string][]eventChannel
		lock        sync.RWMutex
	}

	eventChannel chan Event
)

var _bus = New()

func Publish(topic string, data interface{}) {
	_bus.Publish(topic, data)
}

func Subscribe(topic string, receiver chan Event) {
	_bus.Subscribe(topic, receiver)
}

func New() Bus {
	return &bus{
		subscribers: make(map[string][]eventChannel),
	}
}

func (b *bus) Publish(topic string, data interface{}) {
	b.lock.RLock()
	defer b.lock.RUnlock()
	if chs, ok := b.subscribers[topic]; ok {
		channels := append([]eventChannel{}, chs...)
		go func(data Event, channels []eventChannel) {
			for _, ch := range channels {
				ch <- data
			}
		}(Event{Data: data, Topic: topic}, channels)
	}
}

func (b *bus) Subscribe(topic string, receiver chan Event) {
	b.lock.Lock()
	defer b.lock.Unlock()
	if prev, ok := b.subscribers[topic]; ok {
		b.subscribers[topic] = append(prev, receiver)
	} else {
		b.subscribers[topic] = append([]eventChannel{}, receiver)
	}
}
