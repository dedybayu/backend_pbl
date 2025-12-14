package controllers

import (
	"net/http"
	"strconv"

	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PesanWargaController struct {
	db *gorm.DB
}

func NewPesanWargaController(db *gorm.DB) *PesanWargaController {
	return &PesanWargaController{db: db}
}

/* =========================
   CREATE PESAN
========================= */
func (c *PesanWargaController) Create(ctx *gin.Context) {
	var payload models.PesanWarga

	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.db.Create(&payload).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create message"})
		return
	}

	ctx.JSON(http.StatusCreated, payload)
}

/* =========================
   UPDATE PESAN
========================= */
func (c *PesanWargaController) Update(ctx *gin.Context) {
	id := ctx.Param("id")

	var pesan models.PesanWarga
	if err := c.db.First(&pesan, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	if err := ctx.ShouldBindJSON(&pesan); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := c.db.Save(&pesan).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}

	ctx.JSON(http.StatusOK, pesan)
}

/* =========================
   GET ALL PESAN
========================= */
func (c *PesanWargaController) GetAll(ctx *gin.Context) {
	var pesan []models.PesanWarga

	if err := c.db.
		Preload("FromUser").
		Preload("ToUser").
		Find(&pesan).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var result []gin.H
	for _, p := range pesan {
		result = append(result, gin.H{
			"pesan_id":         p.PesanID,
			"pesan_teks":       p.PesanTeks,
			"from_user":        p.FromUser.Username,
			"from_user_nama":   p.FromUser.UserNama,
			"to_user":          p.ToUser.Username,
			"to_user_nama":     p.ToUser.UserNama,
			"created_at":       p.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, result)
}



/* =========================
   GET BY ID
========================= */
func (c *PesanWargaController) GetByID(ctx *gin.Context) {
	id := ctx.Param("id")

	var p models.PesanWarga
	if err := c.db.
		Preload("FromUser").
		Preload("ToUser").
		First(&p, id).Error; err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"pesan_id":         p.PesanID,
		"pesan_teks":       p.PesanTeks,
		"from_user":        p.FromUser.Username,
		"from_user_nama":   p.FromUser.UserNama,
		"to_user":          p.ToUser.Username,
		"to_user_nama":     p.ToUser.UserNama,
		"created_at":       p.CreatedAt,
	})
}



/* =========================
   GET BY USER ID
   + status masuk / keluar
========================= */
func (c *PesanWargaController) GetByUserID(ctx *gin.Context) {
	userIDParam := ctx.Param("user_id")
	userID, _ := strconv.Atoi(userIDParam)

	authUserID := ctx.GetUint("userID")

	var pesan []models.PesanWarga
	if err := c.db.
		Where("from_user_id = ? OR to_user_id = ?", userID, userID).
		Preload("FromUser").
		Preload("ToUser").
		Order("created_at ASC").
		Find(&pesan).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	var result []gin.H
	for _, p := range pesan {
		status := "masuk"
		if p.FromUserID == authUserID {
			status = "keluar"
		}

		result = append(result, gin.H{
			"pesan_id":         p.PesanID,
			"pesan_teks":       p.PesanTeks,
			"from_user":        p.FromUser.Username,
			"from_user_nama":   p.FromUser.UserNama,
			"to_user":          p.ToUser.Username,
			"to_user_nama":     p.ToUser.UserNama,
			"status":           status,
			"created_at":       p.CreatedAt,
		})
	}

	ctx.JSON(http.StatusOK, result)
}
