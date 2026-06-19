package appliance

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type Event struct {
	Topic string      `json:"topic"`
	Time  string      `json:"time"`
	Data  interface{} `json:"data"`
}

type Hub struct {
	mu      sync.Mutex
	clients map[chan Event]bool
}

func NewHub() *Hub {
	return &Hub{clients: map[chan Event]bool{}}
}

func (h *Hub) Subscribe() chan Event {
	ch := make(chan Event, 16)
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()
	return ch
}

func (h *Hub) Unsubscribe(ch chan Event) {
	h.mu.Lock()
	if h.clients[ch] {
		delete(h.clients, ch)
		close(ch)
	}
	h.mu.Unlock()
}

func (h *Hub) Broadcast(topic string, data interface{}) {
	event := Event{Topic: topic, Time: nowString(), Data: data}
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- event:
		default:
		}
	}
}

func (a *App) eventsHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		unauthorized(w)
		return
	}
	if _, ok := a.sessionByToken(token); !ok {
		unauthorized(w)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := a.hub.Subscribe()
	defer a.hub.Unsubscribe(ch)

	fmt.Fprintf(w, ": connected\n\n")
	flusher.Flush()
	ticker := time.NewTicker(25 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ticker.C:
			fmt.Fprintf(w, ": keepalive\n\n")
			flusher.Flush()
		case event := <-ch:
			raw, _ := json.Marshal(event)
			fmt.Fprintf(w, "event: %s\n", event.Topic)
			fmt.Fprintf(w, "data: %s\n\n", raw)
			flusher.Flush()
		}
	}
}
