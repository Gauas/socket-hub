package realtime

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
)

type Hub struct {
	mu          sync.RWMutex
	channels    map[string]map[*Client]struct{}
	offset      atomic.Uint64
	epoch       string
	writeBuffer int
}

type Publication struct {
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type Result struct {
	Offset uint64 `json:"offset"`
	Epoch  string `json:"epoch"`
}

func New() *Hub {
	return &Hub{
		channels:    make(map[string]map[*Client]struct{}),
		epoch:       "default",
		writeBuffer: 32,
	}
}

func (h *Hub) Subscribe(channel string, conn Connection) *Client {
	client := &Client{
		channel: channel,
		conn:    conn,
		send:    make(chan []byte, h.writeBuffer),
		done:    make(chan struct{}),
	}

	h.mu.Lock()
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]struct{})
	}
	h.channels[channel][client] = struct{}{}
	h.mu.Unlock()

	return client
}

func (h *Hub) Unsubscribe(client *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()

	clients := h.channels[client.channel]
	if clients == nil {
		return
	}

	delete(clients, client)
	if len(clients) == 0 {
		delete(h.channels, client.channel)
	}

	client.Close()
}

func (h *Hub) Publish(ctx context.Context, publication Publication) (Result, error) {
	offset := h.offset.Add(1)
	message, err := json.Marshal(publication)
	if err != nil {
		return Result{}, err
	}

	h.mu.RLock()
	clients := make([]*Client, 0, len(h.channels[publication.Channel]))
	for client := range h.channels[publication.Channel] {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	for _, client := range clients {
		select {
		case client.send <- message:
		case <-ctx.Done():
			return Result{}, ctx.Err()
		default:
		}
	}

	return Result{Offset: offset, Epoch: h.epoch}, nil
}

func (h *Hub) Stats() map[string]any {
	h.mu.RLock()
	defer h.mu.RUnlock()

	clients := 0
	for _, channel := range h.channels {
		clients += len(channel)
	}

	return map[string]any{
		"channels": len(h.channels),
		"clients":  clients,
	}
}
