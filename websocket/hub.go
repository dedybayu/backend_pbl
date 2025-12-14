package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Client represents a connected user
type Client struct {
	UserID   uint
	Conn     *websocket.Conn
	Send     chan []byte
	Username string
}

// Message structure for WebSocket
type WSMessage struct {
	Type      string      `json:"type"`      // chat, typing, online, offline
	FromUser  uint        `json:"from_user_id"`
	ToUser    uint        `json:"to_user_id,omitempty"`
	Content   string      `json:"content,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp string      `json:"timestamp"`
	MessageID uint        `json:"message_id,omitempty"`
}

// Hub manages all WebSocket connections
type Hub struct {
	Clients    map[uint]*Client // Map userID -> Client
	Register   chan *Client
	Unregister chan *Client
	Broadcast  chan *WSMessage
	mu         sync.RWMutex
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true // Untuk development, produksi harus lebih ketat
		},
	}
	GlobalHub *Hub
)

func NewHub() *Hub {
	if GlobalHub == nil {
		GlobalHub = &Hub{
			Clients:    make(map[uint]*Client),
			Register:   make(chan *Client),
			Unregister: make(chan *Client),
			Broadcast:  make(chan *WSMessage),
		}
	}
	return GlobalHub
}

func (h *Hub) Run() {
	log.Println("🚀 WebSocket Hub started")
	for {
		select {
		case client := <-h.Register:
			h.mu.Lock()
			h.Clients[client.UserID] = client
			h.mu.Unlock()
			log.Printf("👤 Client registered - UserID: %d, Total: %d", client.UserID, len(h.Clients))
			
			// Notify user is online
			h.broadcastPresence(client.UserID, true)

		case client := <-h.Unregister:
			h.mu.Lock()
			if _, ok := h.Clients[client.UserID]; ok {
				delete(h.Clients, client.UserID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("👋 Client unregistered - UserID: %d", client.UserID)
			
			// Notify user is offline
			h.broadcastPresence(client.UserID, false)

		case message := <-h.Broadcast:
			h.handleBroadcast(message)
		}
	}
}

func (h *Hub) handleBroadcast(msg *WSMessage) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	switch msg.Type {
	case "chat":
		// Send to specific recipient
		if client, ok := h.Clients[msg.ToUser]; ok {
			msgJSON, _ := json.Marshal(msg)
			select {
			case client.Send <- msgJSON:
				log.Printf("💬 Message sent from %d to %d", msg.FromUser, msg.ToUser)
			default:
				close(client.Send)
				delete(h.Clients, client.UserID)
			}
		}
		
	case "typing":
		// Notify typing status
		if client, ok := h.Clients[msg.ToUser]; ok {
			msgJSON, _ := json.Marshal(msg)
			client.Send <- msgJSON
		}
		
	case "online_status":
		// Broadcast online status to all interested users
		for userID, client := range h.Clients {
			// Skip sender
			if userID == msg.FromUser {
				continue
			}
			
			msgJSON, _ := json.Marshal(msg)
			select {
			case client.Send <- msgJSON:
			default:
				// Connection issue, cleanup
				close(client.Send)
				delete(h.Clients, userID)
			}
		}
	}
}

func (h *Hub) broadcastPresence(userID uint, isOnline bool) {
	status := "offline"
	if isOnline {
		status = "online"
	}

	msg := &WSMessage{
		Type:      "online_status",
		FromUser:  userID,
		Content:   status,
		Timestamp: time.Now().Format(time.RFC3339),
	}

	h.Broadcast <- msg
}

func (h *Hub) IsUserOnline(userID uint) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.Clients[userID]
	return exists
}

func (h *Hub) GetOnlineUsers() []uint {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	onlineUsers := make([]uint, 0, len(h.Clients))
	for userID := range h.Clients {
		onlineUsers = append(onlineUsers, userID)
	}
	return onlineUsers
}