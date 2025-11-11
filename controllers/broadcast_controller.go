package controllers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BroadcastController struct {
	db *gorm.DB
}

func NewBroadcastController(db *gorm.DB) *BroadcastController {
	return &BroadcastController{db: db}
}

// Request structs
type CreateBroadcastRequest struct {
	BroadcastNama      string `json:"broadcast_nama" binding:"required"`
	BroadcastDeskripsi string `json:"broadcast_deskripsi"`
	BroadcastFoto      string `json:"broadcast_foto"`
	BroadcastDokumen   string `json:"broadcast_dokumen"`
}

type UpdateBroadcastRequest struct {
	BroadcastNama      string `json:"broadcast_nama"`
	BroadcastDeskripsi string `json:"broadcast_deskripsi"`
	BroadcastFoto      string `json:"broadcast_foto"`
	BroadcastDokumen   string `json:"broadcast_dokumen"`
}

// ✅ CREATE - Membuat broadcast baru
func (bc *BroadcastController) CreateBroadcast(c *gin.Context) {
	var req CreateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.BroadcastNama = strings.TrimSpace(req.BroadcastNama)
	req.BroadcastDeskripsi = strings.TrimSpace(req.BroadcastDeskripsi)
	req.BroadcastFoto = strings.TrimSpace(req.BroadcastFoto)
	req.BroadcastDokumen = strings.TrimSpace(req.BroadcastDokumen)

	// Validasi required fields
	if req.BroadcastNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama broadcast harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.BroadcastNama) < 2 || len(req.BroadcastNama) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama broadcast harus 2-200 karakter",
		})
		return
	}

	// Check if broadcast dengan nama yang sama sudah ada
	var existingBroadcast models.Broadcast
	if err := bc.db.Where("broadcast_nama = ?", req.BroadcastNama).First(&existingBroadcast).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Broadcast dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat broadcast baru
	broadcast := models.Broadcast{
		BroadcastNama:      req.BroadcastNama,
		BroadcastDeskripsi: req.BroadcastDeskripsi,
		BroadcastFoto:      req.BroadcastFoto,
		BroadcastDokumen:   req.BroadcastDokumen,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := bc.db.Create(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat broadcast",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Broadcast berhasil dibuat",
		"data":    broadcast,
	})
}

// ✅ READ - Mendapatkan semua broadcast
func (bc *BroadcastController) GetAllBroadcast(c *gin.Context) {
	var broadcast []models.Broadcast

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := bc.db.Model(&models.Broadcast{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("broadcast_nama LIKE ? OR broadcast_deskripsi LIKE ?", 
			"%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data broadcast",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": broadcast,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan broadcast by ID
func (bc *BroadcastController) GetBroadcastByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data broadcast",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": broadcast,
	})
}

// ✅ UPDATE - Mengupdate broadcast
func (bc *BroadcastController) UpdateBroadcast(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan broadcast",
			})
		}
		return
	}

	var req UpdateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.BroadcastNama != "" {
		req.BroadcastNama = strings.TrimSpace(req.BroadcastNama)
	}
	if req.BroadcastDeskripsi != "" {
		req.BroadcastDeskripsi = strings.TrimSpace(req.BroadcastDeskripsi)
	}
	if req.BroadcastFoto != "" {
		req.BroadcastFoto = strings.TrimSpace(req.BroadcastFoto)
	}
	if req.BroadcastDokumen != "" {
		req.BroadcastDokumen = strings.TrimSpace(req.BroadcastDokumen)
	}

	// Validasi jika nama diupdate
	if req.BroadcastNama != "" {
		if len(req.BroadcastNama) < 2 || len(req.BroadcastNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama broadcast harus 2-200 karakter",
			})
			return
		}

		// Check duplicate name (exclude current)
		var existingBroadcast models.Broadcast
		if err := bc.db.Where("broadcast_nama = ? AND broadcast_id != ?", req.BroadcastNama, broadcastID).
			First(&existingBroadcast).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Broadcast dengan nama tersebut sudah ada",
			})
			return
		}
	}

	// Update fields menggunakan map (AMAN - GORM Updates dengan map)
	updates := make(map[string]interface{})
	
	if req.BroadcastNama != "" {
		updates["broadcast_nama"] = req.BroadcastNama
	}
	if req.BroadcastDeskripsi != "" {
		updates["broadcast_deskripsi"] = req.BroadcastDeskripsi
	}
	if req.BroadcastFoto != "" {
		updates["broadcast_foto"] = req.BroadcastFoto
	}
	if req.BroadcastDokumen != "" {
		updates["broadcast_dokumen"] = req.BroadcastDokumen
	}
	
	updates["updated_at"] = time.Now()

	if err := bc.db.Model(&broadcast).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate broadcast",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data broadcast yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast berhasil diupdate",
		"data":    broadcast,
	})
}

// ✅ DELETE - Menghapus broadcast
func (bc *BroadcastController) DeleteBroadcast(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan broadcast",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := bc.db.Delete(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus broadcast",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast berhasil dihapus",
	})
}

// ✅ GET - Mendapatkan broadcast terbaru
func (bc *BroadcastController) GetBroadcastTerbaru(c *gin.Context) {
	var broadcast []models.Broadcast

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// Query broadcast terbaru (AMAN)
	if err := bc.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data broadcast terbaru",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  broadcast,
		"total": len(broadcast),
	})
}

// ✅ GET - Pencarian broadcast
func (bc *BroadcastController) SearchBroadcast(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var broadcast []models.Broadcast

	// Execute search query dengan GORM (AMAN - parameterized)
	if err := bc.db.
		Where("broadcast_nama LIKE ? OR broadcast_deskripsi LIKE ?", 
			"%"+search+"%", "%"+search+"%").
		Order("created_at DESC").
		Limit(20).
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mencari broadcast",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         broadcast,
		"total":        len(broadcast),
		"search_query": search,
	})
}

// ✅ GET - Statistik broadcast
func (bc *BroadcastController) GetStatistikBroadcast(c *gin.Context) {
	type StatistikResult struct {
		TotalBroadcast int64 `json:"total_broadcast"`
		BulanIni       int64 `json:"bulan_ini"`
		MingguIni      int64 `json:"minggu_ini"`
	}

	var statistik StatistikResult

	// Hitung total broadcast (AMAN)
	bc.db.Model(&models.Broadcast{}).Count(&statistik.TotalBroadcast)

	// Hitung broadcast bulan ini (AMAN)
	awalBulan := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	bc.db.Model(&models.Broadcast{}).
		Where("created_at >= ?", awalBulan).
		Count(&statistik.BulanIni)

	// Hitung broadcast minggu ini (AMAN)
	awalMinggu := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1)
	bc.db.Model(&models.Broadcast{}).
		Where("created_at >= ?", awalMinggu).
		Count(&statistik.MingguIni)

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ PATCH - Update partial broadcast (hanya deskripsi/foto/dokumen)
func (bc *BroadcastController) UpdatePartialBroadcast(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan broadcast",
			})
		}
		return
	}

	var req UpdateBroadcastRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Update hanya field yang ada di request
	updates := make(map[string]interface{})
	updates["updated_at"] = time.Now()

	if req.BroadcastNama != "" {
		req.BroadcastNama = strings.TrimSpace(req.BroadcastNama)
		if len(req.BroadcastNama) < 2 || len(req.BroadcastNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama broadcast harus 2-200 karakter",
			})
			return
		}
		updates["broadcast_nama"] = req.BroadcastNama
	}

	if req.BroadcastDeskripsi != "" {
		updates["broadcast_deskripsi"] = strings.TrimSpace(req.BroadcastDeskripsi)
	}

	if req.BroadcastFoto != "" {
		updates["broadcast_foto"] = strings.TrimSpace(req.BroadcastFoto)
	}

	if req.BroadcastDokumen != "" {
		updates["broadcast_dokumen"] = strings.TrimSpace(req.BroadcastDokumen)
	}

	if len(updates) == 1 { // Hanya updated_at yang diupdate
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak ada data yang diupdate",
		})
		return
	}

	if err := bc.db.Model(&broadcast).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate broadcast",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data broadcast yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast berhasil diupdate",
		"data":    broadcast,
	})
}