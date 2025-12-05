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

type KategoriPemasukanController struct {
	db *gorm.DB
}

func NewKategoriPemasukanController(db *gorm.DB) *KategoriPemasukanController {
	return &KategoriPemasukanController{db: db}
}

// Request structs
type CreateKategoriPemasukanRequest struct {
	KategoriPemasukanNama string `json:"kategori_pemasukan_nama" binding:"required"`
}

type UpdateKategoriPemasukanRequest struct {
	KategoriPemasukanNama string `json:"kategori_pemasukan_nama" binding:"required"`
}

// ✅ CREATE - Membuat kategori pemasukan baru
func (kpc *KategoriPemasukanController) CreateKategoriPemasukan(c *gin.Context) {
	var req CreateKategoriPemasukanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriPemasukanNama = strings.TrimSpace(req.KategoriPemasukanNama)

	// Validasi required fields
	if req.KategoriPemasukanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pemasukan harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriPemasukanNama) < 2 || len(req.KategoriPemasukanNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pemasukan harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada
	var existingKategori models.KategoriPemasukan
	if err := kpc.db.Where("kategori_pemasukan_nama = ?", req.KategoriPemasukanNama).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori pemasukan dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat kategori pemasukan baru
	kategori := models.KategoriPemasukan{
		KategoriPemasukanNama: req.KategoriPemasukanNama,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := kpc.db.Create(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat kategori pemasukan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kategori pemasukan berhasil dibuat",
		"data":    kategori,
	})
}

// ✅ READ - Mendapatkan semua kategori pemasukan
func (kpc *KategoriPemasukanController) GetAllKategoriPemasukan(c *gin.Context) {
	var kategori []models.KategoriPemasukan

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := kpc.db.Model(&models.KategoriPemasukan{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("kategori_pemasukan_nama LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Preload("Pemasukans").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori pemasukan",
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

// ✅ READ - Mendapatkan kategori pemasukan by ID
func (kpc *KategoriPemasukanController) GetKategoriPemasukanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pemasukan tidak valid",
		})
		return
	}

	var kategori models.KategoriPemasukan
	if err := kpc.db.Preload("Pemasukans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data kategori pemasukan",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}

// ✅ UPDATE - Mengupdate kategori pemasukan
func (kpc *KategoriPemasukanController) UpdateKategoriPemasukan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pemasukan tidak valid",
		})
		return
	}

	var kategori models.KategoriPemasukan
	if err := kpc.db.First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori pemasukan",
			})
		}
		return
	}

	var req UpdateKategoriPemasukanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriPemasukanNama = strings.TrimSpace(req.KategoriPemasukanNama)

	// Validasi required fields
	if req.KategoriPemasukanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pemasukan harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriPemasukanNama) < 2 || len(req.KategoriPemasukanNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori pemasukan harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada (exclude current)
	var existingKategori models.KategoriPemasukan
	if err := kpc.db.Where("kategori_pemasukan_nama = ? AND kategori_pemasukan_id != ?", req.KategoriPemasukanNama, kategoriID).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori pemasukan dengan nama tersebut sudah ada",
		})
		return
	}

	// Update kategori
	kategori.KategoriPemasukanNama = req.KategoriPemasukanNama
	kategori.UpdatedAt = time.Now()

	if err := kpc.db.Save(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate kategori pemasukan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk pemasukans
	if err := kpc.db.Preload("Pemasukans").First(&kategori, kategoriID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kategori pemasukan yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori pemasukan berhasil diupdate",
		"data":    kategori,
	})
}

// ✅ DELETE - Menghapus kategori pemasukan
func (kpc *KategoriPemasukanController) DeleteKategoriPemasukan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori pemasukan tidak valid",
		})
		return
	}

	var kategori models.KategoriPemasukan
	if err := kpc.db.Preload("Pemasukans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori pemasukan",
			})
		}
		return
	}

	// Check jika kategori memiliki pemasukan
	if len(kategori.Pemasukans) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak dapat menghapus kategori yang memiliki data pemasukan",
			"details": gin.H{
				"total_pemasukan": len(kategori.Pemasukans),
			},
		})
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := kpc.db.Delete(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus kategori pemasukan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori pemasukan berhasil dihapus",
	})
}

// ✅ GET - Dropdown kategori pemasukan
func (kpc *KategoriPemasukanController) GetKategoriPemasukanDropdown(c *gin.Context) {
	var kategori []struct {
		KategoriPemasukanID   uint   `json:"kategori_pemasukan_id"`
		KategoriPemasukanNama string `json:"kategori_pemasukan_nama"`
	}

	// Query aman - hanya mengambil field yang diperlukan
	if err := kpc.db.
		Model(&models.KategoriPemasukan{}).
		Select("kategori_pemasukan_id, kategori_pemasukan_nama").
		Order("kategori_pemasukan_nama ASC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dropdown kategori pemasukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}