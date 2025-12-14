package websocket

import (
	"net/http"
	"rt-management/models"
	"strconv"
	"time"

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

// GET /api/pesan/users_chat?user_id=1
func (h *PesanRESTHandler) GetUsersChat(c *gin.Context) {
	userStr := c.Query("user_id")
	userID, err := strconv.ParseUint(userStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user_id"})
		return
	}

	// Struct hasil
	var results []struct {
		UserID      uint      `json:"user_id"`
		Username    string    `json:"username"`
		UserNama    string    `json:"user_nama"`
		LastMessage string    `json:"pesan_terakhir"`
		CreatedAt   time.Time `json:"jam"`
	}

	// Subquery: cari user lain yang pernah chat dengan kita + waktu pesan terakhir
	subQuery := h.DB.Model(&models.PesanWarga{}).
		Select("CASE WHEN from_user_id = ? THEN to_user_id ELSE from_user_id END as user_id, MAX(created_at) as last_time", userID).
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Group("user_id")

	// Join dengan tabel users dan pesan terakhir
	h.DB.Table("(?) as sub", subQuery).
		Select("users.user_id, users.username, users.user_nama, pesan_wargas.pesan_teks as last_message, pesan_wargas.created_at").
		Joins("join users on users.user_id = sub.user_id").
		Joins("join pesan_wargas on ((pesan_wargas.from_user_id = ? AND pesan_wargas.to_user_id = sub.user_id) OR (pesan_wargas.from_user_id = sub.user_id AND pesan_wargas.to_user_id = ?)) AND pesan_wargas.created_at = sub.last_time", userID, userID).
		Scan(&results)

	c.JSON(http.StatusOK, results)
}


