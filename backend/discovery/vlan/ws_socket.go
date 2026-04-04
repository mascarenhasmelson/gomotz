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
	pingPeriod     = (pongWait * 9) / 10
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
	Hub  *Hub
	Conn *websocket.Conn
	Send chan []byte
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
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client] = true
			h.mu.Unlock()
			log.Printf("📱 WebSocket client connected. Total: %d", len(h.Clients))

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("📱 WebSocket client disconnected. Total: %d", len(h.Clients))

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
		}
	}
}
func (h *Hub) BroadcastMessage(message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshaling message: %v", err)
		return
	}
	h.Broadcast <- data
}

func (h *Hub) HandleNotification(notification *utils.DeviceNotification) {
	log.Printf(" Broadcasting notification: %s - %s (VLAN %d)",
		notification.EventType, notification.IPAddress, notification.VLANId)

	h.BroadcastMessage(notification)
}

func (c *Client) ReadPump() {
	defer func() {
		// Safe cleanup with nil checks
		if c != nil && c.Hub != nil {
			c.Hub.Unregister <- c
		}
		if c != nil && c.Conn != nil {
			c.Conn.Close()
		}
	}()

	//  Add nil check
	if c == nil || c.Conn == nil {
		log.Println("ReadPump called with nil client or connection")
		return
	}

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf(" WebSocket error: %v", err)
			}
			break
		}
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

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to current websocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
