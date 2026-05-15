// Package events provides an in-memory pub/sub bus used to fan out
// CRM mutation events (new lead, status change, interaction, telegram message)
// to all connected SSE subscribers in real time.
package events

import (
	"sync"
	"time"
)

// Event is the JSON envelope sent to every subscriber.
type Event struct {
	Type      string      `json:"type"`
	Payload   interface{} `json:"payload,omitempty"`
	Timestamp time.Time   `json:"ts"`
}

// Bus is a non-blocking fan-out. A slow subscriber is skipped (its buffer drops
// the event) rather than stalling other clients.
type Bus struct {
	mu   sync.RWMutex
	subs map[chan Event]struct{}
}

func New() *Bus {
	return &Bus{subs: make(map[chan Event]struct{})}
}

// Subscribe returns a buffered channel for the new subscriber and a cancel func
// that the caller MUST invoke when done (typically on HTTP request context done).
func (b *Bus) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 32)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()
	cancel := func() {
		b.mu.Lock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
		b.mu.Unlock()
	}
	return ch, cancel
}

// Publish broadcasts an event to every current subscriber. Never blocks —
// if a subscriber's buffer is full, that delivery is dropped.
func (b *Bus) Publish(evt Event) {
	if evt.Timestamp.IsZero() {
		evt.Timestamp = time.Now()
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.subs {
		select {
		case ch <- evt:
		default:
			// Subscriber too slow — drop this event for them.
		}
	}
}

// Subscribers count for diagnostics.
func (b *Bus) Subscribers() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.subs)
}
