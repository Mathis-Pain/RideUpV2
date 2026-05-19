package hub

import (
	"sync"

	"github.com/gorilla/websocket"
)

type Client struct {
	EventID string
	UserID  string
	Conn    *websocket.Conn
	Send    chan []byte
}

type Hub struct {
	mu      sync.RWMutex
	rooms   map[string]map[*Client]struct{}
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *Message
}

type Message struct {
	EventID string
	Data    []byte
}

func New() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]struct{}),
		Register:   make(chan *Client, 16),
		Unregister: make(chan *Client, 16),
		Broadcast:  make(chan *Message, 256),
	}
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			if h.rooms[client.EventID] == nil {
				h.rooms[client.EventID] = make(map[*Client]struct{})
			}
			h.rooms[client.EventID][client] = struct{}{}
			h.mu.Unlock()

		case client := <-h.Unregister:
			h.mu.Lock()
			if room, ok := h.rooms[client.EventID]; ok {
				delete(room, client)
				if len(room) == 0 {
					delete(h.rooms, client.EventID)
				}
			}
			h.mu.Unlock()
			close(client.Send)

		case msg := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.rooms[msg.EventID] {
				select {
				case client.Send <- msg.Data:
				default:
					close(client.Send)
					delete(h.rooms[msg.EventID], client)
				}
			}
			h.mu.RUnlock()
		}
	}
}
