package main

import (
	"gantt-backend-go/data"
	"sync"
)

type Event struct {
	Type   string `json:"type"`
	Action string `json:"action"`
	ID     int    `json:"id"`
	Data   any    `json:"data,omitempty"`
	From   string `json:"from"`
}

type TaskAddEvent struct {
	data.Task
	Target int    `json:"target,omitempty"`
	Mode   string `json:"mode,omitempty"`
	Index  *int   `json:"index,omitempty"`
}

type TaskUpdateEvent struct {
	data.Task
	Operation string `json:"operation"`
	Target    int    `json:"target,omitempty"`
	Mode      string `json:"mode,omitempty"`
}

const eventBufferSize = 256

type Subscriber struct {
	ClientID string
	Events   chan Event
	done     chan struct{}
	stopOnce sync.Once
}

func (s *Subscriber) Done() <-chan struct{} {
	return s.done
}

func (s *Subscriber) stop() {
	s.stopOnce.Do(func() { close(s.done) })
}

type Hub struct {
	subs map[string]*Subscriber
	mu   sync.Mutex
}

func (h *Hub) Connect(clientID string) *Subscriber {
	sub := &Subscriber{
		ClientID: clientID,
		Events:   make(chan Event, eventBufferSize),
		done:     make(chan struct{}),
	}
	h.mu.Lock()
	if old, ok := h.subs[clientID]; ok {
		old.stop()
	}
	h.subs[clientID] = sub
	h.mu.Unlock()
	return sub
}

func (h *Hub) Disconnect(sub *Subscriber) {
	h.mu.Lock()
	// Don't remove if the client has already reconnected
	if h.subs[sub.ClientID] == sub {
		delete(h.subs, sub.ClientID)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(clientID string, events []Event) {
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, ev := range events {
		ev.From = clientID
		for id, sub := range h.subs {
			if id == clientID {
				continue
			}
			select {
			case sub.Events <- ev:
			default:
				sub.stop()
			}
		}
	}
}

var hub = &Hub{subs: make(map[string]*Subscriber)}
