package hub

import (
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/coder/websocket"
)

type Event struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]struct{}
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
	done       chan struct{}
}

func New() *Hub {
	return &Hub{
		clients:    make(map[*Client]struct{}),
		broadcast:  make(chan []byte, 256),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		done:       make(chan struct{}),
	}
}

func (h *Hub) Start() {
	go h.run()
	go h.pingLoop()
}

func (h *Hub) Stop() {
	close(h.done)
}

func (h *Hub) run() {
	for {
		select {
		case <-h.done:
			h.mu.Lock()
			for client := range h.clients {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()
			return

		case client := <-h.register:
			h.mu.Lock()
			h.clients[client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client]; ok {
				close(client.send)
				delete(h.clients, client)
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			for client := range h.clients {
				select {
				case client.send <- msg:
				default:
					// Client too slow, drop it
					go func(c *Client) {
						h.unregister <- c
					}(client)
				}
			}
			h.mu.RUnlock()
		}
	}
}

func (h *Hub) pingLoop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-h.done:
			return
		case <-ticker.C:
			h.Broadcast("ping", json.RawMessage("{}"))
		}
	}
}

func (h *Hub) Broadcast(eventType string, payload json.RawMessage) {
	event := Event{Type: eventType, Payload: payload}
	data, err := json.Marshal(event)
	if err != nil {
		log.Printf("Hub: failed to marshal event: %v", err)
		return
	}

	select {
	case h.broadcast <- data:
	default:
		log.Println("Hub: broadcast channel full, dropping event")
	}
}

func (h *Hub) HandleConnection(conn *websocket.Conn) {
	client := &Client{
		conn: conn,
		send: make(chan []byte, 64),
	}

	h.register <- client

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Writer goroutine
	go func() {
		defer func() {
			h.unregister <- client
			conn.Close(websocket.StatusNormalClosure, "")
		}()
		for {
			select {
			case msg, ok := <-client.send:
				if !ok {
					return
				}
				if err := conn.Write(ctx, websocket.MessageText, msg); err != nil {
					return
				}
			case <-ctx.Done():
				return
			}
		}
	}()

	// Reader goroutine (just drains; we don't expect client messages)
	for {
		_, _, err := conn.Read(ctx)
		if err != nil {
			cancel()
			return
		}
	}
}

func (h *Hub) ClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}
