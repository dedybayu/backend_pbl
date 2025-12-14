package controllers

import (
	"net/http"
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PekerjaanController struct {
	db *gorm.DB
}

func NewPekerjaanController(db *gorm.DB) *PekerjaanController {
	return &PekerjaanController{db: db}
}

// GetAll godoc
// @Summary Mendapatkan semua data pekerjaan
// @Description Mengambil seluruh data pekerjaan dari database
// @Tags Pekerjaan
// @Accept json
// @Produce json
// @Success 200 {array} models.Pekerjaan "Berhasil mendapatkan data pekerjaan"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data pekerjaan"
// @Router /api/pekerjaan [get]
func (c *PekerjaanController) GetAll(ctx *gin.Context) {
	var pekerjaans []models.Pekerjaan

	// Gunakan WithContext
	if err := c.db.WithContext(ctx.Request.Context()).Find(&pekerjaans).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch pekerjaan data",
		})
		return
	}

	ctx.JSON(http.StatusOK, pekerjaans)
}