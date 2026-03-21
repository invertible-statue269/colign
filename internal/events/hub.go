package events

import "sync"

// Event represents a real-time event to broadcast to subscribers.
type Event struct {
	Type     string // "task_created", "task_updated", "document_updated", "change_updated"
	ChangeID int64
	Payload  string // JSON-encoded data
}

// Hub manages event subscribers and broadcasting.
type Hub struct {
	mu   sync.RWMutex
	subs map[*Subscriber]struct{}
}

// Subscriber receives events on its channel.
type Subscriber struct {
	changeID int64 // 0 = all changes
	ch       chan Event
}

func NewHub() *Hub {
	return &Hub{subs: make(map[*Subscriber]struct{})}
}

// Subscribe creates a new subscriber. changeID=0 means subscribe to all.
func (h *Hub) Subscribe(changeID int64) *Subscriber {
	s := &Subscriber{
		changeID: changeID,
		ch:       make(chan Event, 32),
	}
	h.mu.Lock()
	h.subs[s] = struct{}{}
	h.mu.Unlock()
	return s
}

// Unsubscribe removes a subscriber and closes its channel.
func (h *Hub) Unsubscribe(s *Subscriber) {
	h.mu.Lock()
	delete(h.subs, s)
	h.mu.Unlock()
	close(s.ch)
}

// Publish sends an event to all matching subscribers.
func (h *Hub) Publish(e Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for s := range h.subs {
		if s.changeID == 0 || s.changeID == e.ChangeID {
			select {
			case s.ch <- e:
			default:
				// drop if subscriber is slow
			}
		}
	}
}

// Events returns the subscriber's event channel.
func (s *Subscriber) Events() <-chan Event {
	return s.ch
}
