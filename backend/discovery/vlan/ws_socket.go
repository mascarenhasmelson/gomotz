package vlan

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mascarenhasmelson/gomotz/utils"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10 // 54 seconds
	maxMessageSize = 512
)

type Hub struct {
	Clients    map[*Client]bool
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	mu         sync.RWMutex
}

type Client struct {
	Hub      *Hub
	Conn     *websocket.Conn
	Send     chan []byte
	mu       sync.Mutex
	LastPing time.Time
}

func NewHub() *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte, 256),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
	}
}

func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("  WebSocket client connected. Total: %d", len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("   WebSocket client disconnected. Total: %d", len(h.Clients))

		case message := <-h.Broadcast:
			h.mu.RLock()
			for client := range h.Clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.Clients, client)
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			h.cleanupStaleClients()
		}
	}
}

func (h *Hub) cleanupStaleClients() {
	h.mu.Lock()
	defer h.mu.Unlock()
	now := time.Now()
	for client := range h.Clients {
		client.mu.Lock()
		if now.Sub(client.LastPing) > pongWait+10*time.Second {
			client.mu.Unlock()
			log.Printf("Removing stale client, last ping: %v ago", now.Sub(client.LastPing))
			delete(h.Clients, client)
			close(client.Send)
			client.Conn.Close()
		} else {
			client.mu.Unlock()
		}
	}
}

func (h *Hub) BroadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	select {
	case h.Broadcast <- data:
	default:
		log.Printf("Broadcast channel full, dropping message")
	}
}

func (h *Hub) HandleNotification(notification *utils.DeviceNotification) {
	log.Printf("    Broadcasting notification: %s - %s (networkid%d)",
		notification.EventType, notification.IPAddress, notification.NetworkId)

	h.BroadcastMessage(notification)
}

func (c *Client) ReadPump() {
	defer func() {
		if c != nil && c.Hub != nil {
			c.Hub.Unregister <- c
		}
		if c != nil && c.Conn != nil {
			c.Conn.Close()
		}
	}()

	if c == nil || c.Conn == nil {
		log.Println("ReadPump called with nil client or connection")
		return
	}
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	c.Conn.SetPingHandler(func(data string) error {
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
		c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
		err := c.Conn.WriteMessage(websocket.PongMessage, []byte(data))
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return err
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			}
			break
		}

		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err == nil {
			if msg["type"] == "ping" {
				c.mu.Lock()
				c.LastPing = time.Now()
				c.mu.Unlock()
				response := map[string]string{"type": "pong"}
				if respData, err := json.Marshal(response); err == nil {
					select {
					case c.Send <- respData:
					default:
						log.Printf("Client send buffer full, dropping pong response")
					}
				}
				continue
			}
		}
		c.mu.Lock()
		c.LastPing = time.Now()
		c.mu.Unlock()
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		if c != nil && c.Conn != nil {
			c.Conn.Close()
		}
	}()

	if c == nil || c.Conn == nil || c.Send == nil {
		log.Println("WritePump called with nil client, connection, or send channel")
		return
	}
	c.mu.Lock()
	c.LastPing = time.Now()
	c.mu.Unlock()

	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				log.Printf("Failed to get writer: %v", err)
				return
			}

			if _, err := w.Write(message); err != nil {
				log.Printf("Failed to write message: %v", err)
				w.Close()
				return
			}
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				log.Printf("Failed to close writer: %v", err)
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("Failed to send ping: %v", err)
				return
			}
			c.mu.Lock()
			c.LastPing = time.Now()
			c.mu.Unlock()
		}
	}
}
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.Clients)
}

func (h *Hub) BroadcastToAll(message []byte) {
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.Clients))
	for client := range h.Clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()
	for _, client := range clients {
		select {
		case client.Send <- message:
		default:
			log.Printf("Client send buffer full, skipping")
		}
	}
}
