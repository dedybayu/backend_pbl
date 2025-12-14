package controllers

import (
	"net/http"
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AgamaController struct {
	db *gorm.DB
}

func NewAgamaController(db *gorm.DB) *AgamaController {
	return &AgamaController{db: db}
}

// GetAll godoc
// @Summary Mendapatkan semua data agama
// @Description Mengambil seluruh data agama dari database
// @Tags Agama
// @Accept json
// @Produce json
// @Success 200 {array} models.Agama "Berhasil mendapatkan data agama"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data agama"
// @Router /api/agama [get]
func (c *AgamaController) GetAll(ctx *gin.Context) {
	var agamas []models.Agama

	// Gunakan WithContext untuk menghindari race condition
	if err := c.db.WithContext(ctx.Request.Context()).Find(&agamas).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch agama data",
		})
		return
	}

	ctx.JSON(http.StatusOK, agamas)
}