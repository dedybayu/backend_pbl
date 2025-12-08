package controllers

import (
	"net/http"
	"strconv"
	// "time"

	"rt-management/models"
	// "rt-management/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RumahController struct {
	db *gorm.DB
}

func NewRumahController(db *gorm.DB) *RumahController {
	return &RumahController{db: db}
}

// Request structs
type CreateRumahRequest struct {
	RumahAlamat string `form:"rumah_alamat" binding:"required"`
	RumahStatus string `form:"rumah_status"`
}

type UpdateRumahRequest struct {
	RumahAlamat string `form:"rumah_alamat"`
	RumahStatus string `form:"rumah_status"`
}

func (rc *RumahController) CreateRumah(c *gin.Context) {
	var req CreateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validasi alamat
	if req.RumahAlamat == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alamat rumah harus diisi",
		})
		return
	}

	// Set default status
	if req.RumahStatus == "" {
		req.RumahStatus = "tersedia"
	}

	// Validasi status
	if !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	rumah := models.Rumah{
		RumahAlamat: req.RumahAlamat,
		RumahStatus: req.RumahStatus,
	}

	if err := rc.db.Create(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rumah berhasil dibuat",
		"data":    rumah,
	})
}

func (rc *RumahController) GetAllRumah(c *gin.Context) {
	var rumah []models.Rumah

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	status := c.Query("status")
	wargaID := c.Query("warga_id")

	query := rc.db.Model(&models.Rumah{})

	if status != "" {
		query = query.Where("rumah_status = ?", status)
	}

	// Filter rumah berdasarkan warga penghuni
	if wargaID != "" {
		query = query.Joins("JOIN wargas ON wargas.rumah_id = rumahs.rumah_id").
			Where("wargas.warga_id = ?", wargaID)
	}

	var total int64
	query.Count(&total)

	if err := query.Offset(offset).Limit(limit).Find(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data rumah",
		})
		return
	}

	// Tambahkan data warga penghuni
	type RumahWithWarga struct {
		models.Rumah
		Warga models.Warga `json:"warga"`
	}

	var results []RumahWithWarga

	for _, r := range rumah {
		var warga models.Warga
		rc.db.Where("rumah_id = ?", r.RumahID).First(&warga)

		results = append(results, RumahWithWarga{
			Rumah: r,
			Warga: warga,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

func (rc *RumahController) GetRumahByID(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rumah tidak ditemukan",
		})
		return
	}

	var warga models.Warga
	rc.db.Where("rumah_id = ?", rumah.RumahID).First(&warga)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"rumah": rumah,
			"warga": warga,
		},
	})
}


func (rc *RumahController) UpdateRumah(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rumah tidak ditemukan",
		})
		return
	}

	var req UpdateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	if req.RumahStatus != "" && !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	updates := map[string]interface{}{}
	if req.RumahAlamat != "" {
		updates["rumah_alamat"] = req.RumahAlamat
	}
	if req.RumahStatus != "" {
		updates["rumah_status"] = req.RumahStatus
	}

	if err := rc.db.Model(&rumah).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal update rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rumah berhasil diupdate",
		"data":    rumah,
	})
}

// ✅ DELETE - Menghapus rumah
func (rc *RumahController) DeleteRumah(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Rumah tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan rumah",
			})
		}
		return
	}

	if err := rc.db.Delete(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rumah berhasil dihapus",
	})
}

// ✅ GET - Mendapatkan rumah by status
func (rc *RumahController) GetRumahByStatus(c *gin.Context) {
	status := c.Param("status")

	if !isValidRumahStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Gunakan 'tersedia' atau 'ditempati'",
		})
		return
	}

	var rumah []models.Rumah
	if err := rc.db.Preload("Warga").Where("rumah_status = ?", status).Find(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data rumah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rumah,
		"total": len(rumah),
	})
}

// ✅ Helper function untuk validasi status rumah
func isValidRumahStatus(status string) bool {
	return status == "tersedia" || status == "ditempati"
}
