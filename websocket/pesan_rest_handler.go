package websocket

import (
	"net/http"
	"rt-management/models"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PesanRESTHandler struct {
	DB *gorm.DB
}

func NewPesanRESTHandler(db *gorm.DB) *PesanRESTHandler {
	return &PesanRESTHandler{DB: db}
}

// GET /api/pesan?from=1&to=2
func (h *PesanRESTHandler) GetChatHistory(c *gin.Context) {
	fromStr := c.Query("from")
	toStr := c.Query("to")

	fromID, err := strconv.ParseUint(fromStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid from_user_id"})
		return
	}

	toID, err := strconv.ParseUint(toStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid to_user_id"})
		return
	}

	var messages []models.PesanWarga

	// Ambil semua pesan antara dua user
	if err := h.DB.
		Where("(from_user_id = ? AND to_user_id = ?) OR (from_user_id = ? AND to_user_id = ?)",
			fromID, toID, toID, fromID).
		Order("created_at ASC").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch messages"})
		return
	}

	// Format response sama seperti WebSocket
	var resp []gin.H
	for _, msg := range messages {
		resp = append(resp, gin.H{
			"pesan_id":     msg.PesanID,
			"pesan_teks":   msg.PesanTeks,
			"from_user_id": msg.FromUserID,
			"to_user_id":   msg.ToUserID,
			"created_at":   msg.CreatedAt,
		})
	}

	c.JSON(http.StatusOK, resp)
}
