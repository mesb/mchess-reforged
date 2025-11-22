package server

import "sync"

// Broadcast represents a message destined for all clients in a game room.
type Broadcast struct {
	GameID  string
	Payload []byte
}

// Hub maintains active WebSocket clients grouped by game ID.
type Hub struct {
	register   chan *Client
	unregister chan *Client
	broadcast  chan Broadcast

	rooms map[string]map[*Client]bool
	mu    sync.RWMutex
}

// NewHub constructs a Hub.
func NewHub() *Hub {
	return &Hub{
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Broadcast, 16),
		rooms:      make(map[string]map[*Client]bool),
	}
}

// Run processes register/unregister/broadcast events.
func (h *Hub) Run() {
	for {
		select {
		case c := <-h.register:
			h.addClient(c)
		case c := <-h.unregister:
			h.removeClient(c)
		case msg := <-h.broadcast:
			h.sendToRoom(msg)
		}
	}
}

// BroadcastTo broadcasts a payload to all clients in a game room.
func (h *Hub) BroadcastTo(gameID string, payload []byte) {
	h.broadcast <- Broadcast{GameID: gameID, Payload: payload}
}

func (h *Hub) addClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.rooms[c.gameID] == nil {
		h.rooms[c.gameID] = make(map[*Client]bool)
	}
	h.rooms[c.gameID][c] = true
}

func (h *Hub) removeClient(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	clients := h.rooms[c.gameID]
	if clients == nil {
		return
	}
	if _, ok := clients[c]; ok {
		delete(clients, c)
		close(c.send)
		if len(clients) == 0 {
			delete(h.rooms, c.gameID)
		}
	}
}

func (h *Hub) sendToRoom(msg Broadcast) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	clients := h.rooms[msg.GameID]
	for c := range clients {
		select {
		case c.send <- msg.Payload:
		default:
			// Drop slow clients
			go h.unregisterClientAsync(c)
		}
	}
}

func (h *Hub) unregisterClientAsync(c *Client) {
	h.unregister <- c
}
