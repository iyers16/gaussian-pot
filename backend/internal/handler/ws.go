package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSEvent is the envelope for all WebSocket messages.
type WSEvent struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Hub manages all connected WebSocket clients and broadcasts events.
type Hub struct {
	mu      sync.RWMutex
	clients map[*wsClient]bool
}

type wsClient struct {
	conn *websocket.Conn
	send chan []byte
}

func NewHub() *Hub {
	return &Hub{clients: make(map[*wsClient]bool)}
}

// Broadcast sends an event to all connected clients.
func (h *Hub) Broadcast(eventType string, payload interface{}) {
	msg, err := json.Marshal(WSEvent{Type: eventType, Payload: payload})
	if err != nil {
		log.Printf("ws broadcast marshal error: %v", err)
		return
	}
	h.mu.RLock()
	for client := range h.clients {
		select {
		case client.send <- msg:
		default:
			// Slow client — drop message rather than block.
		}
	}
	h.mu.RUnlock()
}

// HandleWS upgrades an HTTP connection to WebSocket and registers the client.
func (h *Hub) HandleWS(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("ws upgrade error: %v", err)
		return
	}

	client := &wsClient{conn: conn, send: make(chan []byte, 64)}
	h.mu.Lock()
	h.clients[client] = true
	h.mu.Unlock()

	go h.writePump(client)
	h.readPump(client) // blocks until disconnect
}

func (h *Hub) writePump(c *wsClient) {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func (h *Hub) readPump(c *wsClient) {
	defer func() {
		h.mu.Lock()
		delete(h.clients, c)
		close(c.send)
		h.mu.Unlock()
		c.conn.Close()
	}()
	for {
		// We only receive pings / disconnects; game input comes via REST.
		if _, _, err := c.conn.ReadMessage(); err != nil {
			break
		}
	}
}
