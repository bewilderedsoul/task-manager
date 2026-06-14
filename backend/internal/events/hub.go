// Package events implements a minimal in-process pub/sub hub used to push
// real-time task changes to connected clients over Server-Sent Events (SSE).
//
// It is intentionally simple: events are fanned out per user id. In a
// multi-instance deployment this would be backed by Redis/NATS instead, but for
// a single instance an in-memory hub is enough and keeps the setup dependency-free.
package events

import (
	"sync"

	"github.com/google/uuid"
)

// Event is a JSON-serialisable message broadcast to a user's subscribers.
type Event struct {
	Type string `json:"type"` // task.created | task.updated | task.deleted
	Data any    `json:"data"`
}

type Hub struct {
	mu   sync.RWMutex
	subs map[uuid.UUID]map[chan Event]struct{}
}

func NewHub() *Hub {
	return &Hub{subs: make(map[uuid.UUID]map[chan Event]struct{})}
}

// Subscribe registers a new channel for a user and returns it along with an
// unsubscribe function the caller must defer.
func (h *Hub) Subscribe(userID uuid.UUID) (<-chan Event, func()) {
	ch := make(chan Event, 8)
	h.mu.Lock()
	if h.subs[userID] == nil {
		h.subs[userID] = make(map[chan Event]struct{})
	}
	h.subs[userID][ch] = struct{}{}
	h.mu.Unlock()

	return ch, func() {
		h.mu.Lock()
		delete(h.subs[userID], ch)
		if len(h.subs[userID]) == 0 {
			delete(h.subs, userID)
		}
		h.mu.Unlock()
		close(ch)
	}
}

// Publish sends an event to every subscriber of the given user. Slow consumers
// are skipped (non-blocking send) so one stuck client can't stall the others.
func (h *Hub) Publish(userID uuid.UUID, ev Event) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.subs[userID] {
		select {
		case ch <- ev:
		default:
		}
	}
}
