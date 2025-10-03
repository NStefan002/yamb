package broadcaster

import "sync"

type Broadcaster struct {
	subscribers map[chan Event]struct{}
	mu          sync.Mutex
}

func NewBroadcaster() *Broadcaster {
	return &Broadcaster{
		subscribers: make(map[chan Event]struct{}),
	}
}

func (b *Broadcaster) Subscribe() chan Event {
	ch := make(chan Event, 5)
	b.mu.Lock()
	b.subscribers[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *Broadcaster) Unsubscribe(ch chan Event) {
	b.mu.Lock()
	delete(b.subscribers, ch)
	close(ch)
	b.mu.Unlock()
}

func (b *Broadcaster) Broadcast(event Event) {
	b.mu.Lock()
	for ch := range b.subscribers {
		select {
		case ch <- event:
		default:
			// Drop the event if the channel is full
		}
	}
	b.mu.Unlock()
}
