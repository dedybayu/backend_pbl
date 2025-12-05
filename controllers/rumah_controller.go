package controllers

import (
	"net/http"
	"strconv"
	"time"

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
	WargaID     uint   `form:"warga_id"`
}

type UpdateRumahRequest struct {
	RumahAlamat string `form:"rumah_alamat"`
	RumahStatus string `form:"rumah_status"`
	WargaID     uint   `form:"warga_id"`
}

// ✅ CREATE - Membuat rumah baru
func (rc *RumahController) CreateRumah(c *gin.Context) {
	var req CreateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validasi required fields
	if req.RumahAlamat == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alamat rumah harus diisi",
		})
		return
	}

	// Set default status jika tidak diisi
	if req.RumahStatus == "" {
		req.RumahStatus = "tersedia"
	}

	// Validasi status
	if !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status rumah harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	// Jika WargaID diisi, validasi apakah warga exists
	if req.WargaID != 0 {
		var warga models.Warga
		if err := rc.db.First(&warga, req.WargaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Warga tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi warga",
				})
			}
			return
		}
	}

	// Buat rumah baru
	rumah := models.Rumah{
		RumahAlamat: req.RumahAlamat,
		RumahStatus: req.RumahStatus,
		WargaID:     req.WargaID,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := rc.db.Create(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat rumah",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data warga
	if err := rc.db.Preload("Warga").First(&rumah, rumah.RumahID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data rumah yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rumah berhasil dibuat",
		"data":    rumah,
	})
}

// ✅ READ - Mendapatkan semua rumah
func (rc *RumahController) GetAllRumah(c *gin.Context) {
	var rumah []models.Rumah

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	status := c.Query("status")
	wargaID := c.Query("warga_id")

	query := rc.db.Preload("Warga")

	// Apply filters
	if status != "" {
		query = query.Where("rumah_status = ?", status)
	}
	if wargaID != "" {
		query = query.Where("warga_id = ?", wargaID)
	}

	// Get total count for pagination
	var total int64
	query.Model(&models.Rumah{}).Count(&total)

	// Execute query with pagination
	if err := query.
		Offset(offset).
		// Limit(limit).
		Find(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data rumah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rumah,
		// "pagination": gin.H{
		// 	"page":  page,
		// 	"limit": limit,
		// 	"total": total,
		// },
	})
}

// ✅ READ - Mendapatkan rumah by ID
func (rc *RumahController) GetRumahByID(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.Preload("Warga").First(&rumah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Rumah tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data rumah",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": rumah,
	})
}

// ✅ UPDATE - Mengupdate rumah
func (rc *RumahController) UpdateRumah(c *gin.Context) {
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

	var req UpdateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Validasi status jika diupdate
	if req.RumahStatus != "" && !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status rumah harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	// Jika WargaID diisi, validasi apakah warga exists
	if req.WargaID != 0 {
		var warga models.Warga
		if err := rc.db.First(&warga, req.WargaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Warga tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi warga",
				})
			}
			return
		}
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.RumahAlamat != "" {
		updates["rumah_alamat"] = req.RumahAlamat
	}
	if req.RumahStatus != "" {
		updates["rumah_status"] = req.RumahStatus
	}
	if req.WargaID != 0 {
		updates["warga_id"] = req.WargaID
	}
	updates["updated_at"] = time.Now()

	if err := rc.db.Model(&rumah).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate rumah",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru
	if err := rc.db.Preload("Warga").First(&rumah, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data rumah yang diupdate",
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
