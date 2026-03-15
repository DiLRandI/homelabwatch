package events

import (
	"sync"

	"github.com/deleema/homelabwatch/internal/domain"
)

type Bus struct {
	mu          sync.RWMutex
	subscribers map[chan domain.EventEnvelope]struct{}
}

func NewBus() *Bus {
	return &Bus{subscribers: make(map[chan domain.EventEnvelope]struct{})}
}

func (b *Bus) Subscribe(buffer int) chan domain.EventEnvelope {
	ch := make(chan domain.EventEnvelope, buffer)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Bus) Unsubscribe(ch chan domain.EventEnvelope) {
	b.mu.Lock()
	if _, ok := b.subscribers[ch]; ok {
		delete(b.subscribers, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *Bus) Publish(event domain.EventEnvelope) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
		}
	}
}
