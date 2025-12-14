package websocket

import (
	"encoding/json"
	"log"
	"time"

	"github.com/gorilla/websocket"
)

func (c *Client) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(5120) // 5KB
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("❌ WebSocket error for user %d: %v", c.UserID, err)
			}
			break
		}

		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			log.Printf("❌ Error unmarshaling message: %v", err)
			continue
		}

		// Handle different message types
		switch wsMsg.Type {
		case "chat":
			if wsMsg.Content != "" && wsMsg.ToUser != 0 {
				// Set timestamp
				wsMsg.Timestamp = time.Now().Format(time.RFC3339)
				wsMsg.FromUser = c.UserID
				
				// Forward to hub
				hub.Broadcast <- &wsMsg
			}
			
		case "typing":
			wsMsg.FromUser = c.UserID
			wsMsg.Timestamp = time.Now().Format(time.RFC3339)
			hub.Broadcast <- &wsMsg
			
		case "ping":
			// Respond to ping
			resp := WSMessage{
				Type:      "pong",
				Timestamp: time.Now().Format(time.RFC3339),
			}
			respJSON, _ := json.Marshal(resp)
			c.Send <- respJSON
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second) // Ping interval
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
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

			// Close writer
			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			// Send ping
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}