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

type KategoriPengeluaranController struct {
	db *gorm.DB
}

func NewKategoriPengeluaranController(db *gorm.DB) *KategoriPengeluaranController {
	return &KategoriPengeluaranController{db: db}
}

// Request structs
type CreateKategoriPengeluaranRequest struct {
	KategoriPengeluaranNama string `json:"kategori_pengeluaran_nama" binding:"required"`
}

type UpdateKategoriPengeluaranRequest struct {
	KategoriPengeluaranNama string `json:"kategori_pengeluaran_nama" binding:"required"`
}

// ✅ CREATE - Membuat kategori pengeluaran baru
func (kpc *KategoriPengeluaranController) CreateKategoriPengeluaran(c *gin.Context) {
	var req CreateKategoriPengeluaranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriPengeluaranNama = strings.TrimSpace(req.KategoriPengeluaranNama)

	// Validasi required fields
	if req.KategoriPengeluaranNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pengeluaran harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriPengeluaranNama) < 2 || len(req.KategoriPengeluaranNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pengeluaran harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada
	var existingKategori models.KategoriPengeluaran
	if err := kpc.db.Where("kategori_pengeluaran_nama = ?", req.KategoriPengeluaranNama).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori pengeluaran dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat kategori pengeluaran baru
	kategori := models.KategoriPengeluaran{
		KategoriPengeluaranNama: req.KategoriPengeluaranNama,
		CreatedAt:               time.Now(),
		UpdatedAt:               time.Now(),
	}

	if err := kpc.db.Create(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat kategori pengeluaran",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kategori pengeluaran berhasil dibuat",
		"data":    kategori,
	})
}

// ✅ READ - Mendapatkan semua kategori pengeluaran
func (kpc *KategoriPengeluaranController) GetAllKategoriPengeluaran(c *gin.Context) {
	var kategori []models.KategoriPengeluaran

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := kpc.db.Model(&models.KategoriPengeluaran{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("kategori_pengeluaran_nama LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Preload("Pengeluarans").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori pengeluaran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan kategori pengeluaran by ID
func (kpc *KategoriPengeluaranController) GetKategoriPengeluaranByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pengeluaran tidak valid",
		})
		return
	}

	var kategori models.KategoriPengeluaran
	if err := kpc.db.Preload("Pengeluarans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data kategori pengeluaran",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}

// ✅ UPDATE - Mengupdate kategori pengeluaran
func (kpc *KategoriPengeluaranController) UpdateKategoriPengeluaran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pengeluaran tidak valid",
		})
		return
	}

	var kategori models.KategoriPengeluaran
	if err := kpc.db.First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori pengeluaran",
			})
		}
		return
	}

	var req UpdateKategoriPengeluaranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriPengeluaranNama = strings.TrimSpace(req.KategoriPengeluaranNama)

	// Validasi required fields
	if req.KategoriPengeluaranNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pengeluaran harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriPengeluaranNama) < 2 || len(req.KategoriPengeluaranNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pengeluaran harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada (exclude current)
	var existingKategori models.KategoriPengeluaran
	if err := kpc.db.Where("kategori_pengeluaran_nama = ? AND kategori_pengeluaran_id != ?", req.KategoriPengeluaranNama, kategoriID).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori pengeluaran dengan nama tersebut sudah ada",
		})
		return
	}

	// Update kategori
	kategori.KategoriPengeluaranNama = req.KategoriPengeluaranNama
	kategori.UpdatedAt = time.Now()

	if err := kpc.db.Save(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate kategori pengeluaran",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk pengeluarans
	if err := kpc.db.Preload("Pengeluarans").First(&kategori, kategoriID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kategori pengeluaran yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori pengeluaran berhasil diupdate",
		"data":    kategori,
	})
}

// ✅ DELETE - Menghapus kategori pengeluaran
func (kpc *KategoriPengeluaranController) DeleteKategoriPengeluaran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pengeluaran tidak valid",
		})
		return
	}

	var kategori models.KategoriPengeluaran
	if err := kpc.db.Preload("Pengeluarans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori pengeluaran",
			})
		}
		return
	}

	// Check jika kategori memiliki pengeluaran
	if len(kategori.Pengeluarans) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak dapat menghapus kategori yang memiliki data pengeluaran",
			"details": gin.H{
				"total_pengeluaran": len(kategori.Pengeluarans),
			},
		})
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := kpc.db.Delete(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus kategori pengeluaran",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori pengeluaran berhasil dihapus",
	})
}

// ✅ GET - Dropdown kategori pengeluaran
func (kpc *KategoriPengeluaranController) GetKategoriPengeluaranDropdown(c *gin.Context) {
	var kategori []struct {
		KategoriPengeluaranID   uint   `json:"kategori_pengeluaran_id"`
		KategoriPengeluaranNama string `json:"kategori_pengeluaran_nama"`
	}

	// Query aman - hanya mengambil field yang diperlukan
	if err := kpc.db.
		Model(&models.KategoriPengeluaran{}).
		Select("kategori_pengeluaran_id, kategori_pengeluaran_nama").
		Order("kategori_pengeluaran_nama ASC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dropdown kategori pengeluaran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}