package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"rt-management/models"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"gorm.io/gorm"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type PesanWSHandler struct {
	DB  *gorm.DB
	Hub *Hub
}

func NewPesanWSHandler(db *gorm.DB, hub *Hub) *PesanWSHandler {
	return &PesanWSHandler{DB: db, Hub: hub}
}

type PesanPayload struct {
	Type       string `json:"type"`
	FromUserID uint   `json:"from_user_id"`
	ToUserID   uint   `json:"to_user_id"`
	Pesan      string `json:"pesan"`
}

func (h *PesanWSHandler) Handle(c *gin.Context) {
	log.Println("🔥 WS HANDLE DIPANGGIL")

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("❌ Upgrade error:", err)
		return
	}

	conn.SetReadLimit(5120)
	conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	client := &Client{
		Conn: conn,
		Send: make(chan []byte),
		Hub:  h.Hub,
	}

	h.Hub.Register <- client

	go client.WritePump()

	for {
		var payload PesanPayload

		if err := conn.ReadJSON(&payload); err != nil {
			log.Println("❌ Read error:", err)
			break
		}

		// 1️⃣ SIMPAN KE DATABASE
		pesan := models.PesanWarga{
			PesanTeks:  payload.Pesan,
			FromUserID: payload.FromUserID,
			ToUserID:   payload.ToUserID,
		}

		if err := h.DB.Create(&pesan).Error; err != nil {
			log.Println("❌ DB insert error:", err)
			continue
		}

		// 2️⃣ FORMAT RESPONSE (SAMAKAN DENGAN API)
		response := map[string]interface{}{
			"pesan_id":     pesan.PesanID,
			"pesan_teks":   pesan.PesanTeks,
			"from_user_id": pesan.FromUserID,
			"to_user_id":   pesan.ToUserID,
			"created_at":   pesan.CreatedAt,
		}

		data, _ := json.Marshal(response)

		// 3️⃣ BROADCAST KE CLIENT
		h.Hub.Broadcast <- data
	}

	h.Hub.Unregister <- client

	conn.Close()
}
