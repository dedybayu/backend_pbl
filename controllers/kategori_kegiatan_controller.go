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

type KategoriKegiatanController struct {
	db *gorm.DB
}

func NewKategoriKegiatanController(db *gorm.DB) *KategoriKegiatanController {
	return &KategoriKegiatanController{db: db}
}

// Request structs
type CreateKategoriKegiatanRequest struct {
	KategoriKegiatanNama string `form:"kategori_kegiatan_nama" binding:"required"`
}

type UpdateKategoriKegiatanRequest struct {
	KategoriKegiatanNama string `form:"kategori_kegiatan_nama" binding:"required"`
}

// ✅ CREATE - Membuat kategori kegiatan baru
func (kc *KategoriKegiatanController) CreateKategoriKegiatan(c *gin.Context) {
	var req CreateKategoriKegiatanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriKegiatanNama = strings.TrimSpace(req.KategoriKegiatanNama)

	// Validasi required fields
	if req.KategoriKegiatanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori kegiatan harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriKegiatanNama) < 2 || len(req.KategoriKegiatanNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori kegiatan harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada (AMAN - menggunakan GORM Where dengan parameter)
	var existingKategori models.KategoriKegiatan
	if err := kc.db.Where("kategori_kegiatan_nama = ?", req.KategoriKegiatanNama).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori kegiatan dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat kategori kegiatan baru
	kategori := models.KategoriKegiatan{
		KategoriKegiatanNama: req.KategoriKegiatanNama,
		CreatedAt:            time.Now(),
		UpdatedAt:            time.Now(),
	}

	if err := kc.db.Create(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat kategori kegiatan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kategori kegiatan berhasil dibuat",
		"data":    kategori,
	})
}

// ✅ READ - Mendapatkan semua kategori kegiatan
func (kc *KategoriKegiatanController) GetAllKategoriKegiatan(c *gin.Context) {
	var kategori []models.KategoriKegiatan

	// Pagination parameters (AMAN - dikonversi ke integer)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter (AMAN - digunakan sebagai parameter dalam Where)
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := kc.db.Model(&models.KategoriKegiatan{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("kategori_kegiatan_nama LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan preload kegiatan dan pagination
	if err := query.Preload("Kegiatans").
		Offset(offset).
		// Limit(limit).
		Order("created_at DESC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
		// "pagination": gin.H{
		// 	"page":  page,
		// 	"limit": limit,
		// 	"total": total,
		// },
	})
}

// ✅ READ - Mendapatkan kategori kegiatan by ID (AMAN - GORM First dengan ID)
func (kc *KategoriKegiatanController) GetKategoriKegiatanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori kegiatan tidak valid",
		})
		return
	}

	var kategori models.KategoriKegiatan
	if err := kc.db.Preload("Kegiatans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data kategori kegiatan",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}

// ✅ UPDATE - Mengupdate kategori kegiatan
func (kc *KategoriKegiatanController) UpdateKategoriKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori kegiatan tidak valid",
		})
		return
	}

	var kategori models.KategoriKegiatan
	if err := kc.db.First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori kegiatan",
			})
		}
		return
	}

	var req UpdateKategoriKegiatanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriKegiatanNama = strings.TrimSpace(req.KategoriKegiatanNama)

	// Validasi required fields
	if req.KategoriKegiatanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori kegiatan harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriKegiatanNama) < 2 || len(req.KategoriKegiatanNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori kegiatan harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada (exclude current) - AMAN
	var existingKategori models.KategoriKegiatan
	if err := kc.db.Where("kategori_kegiatan_nama = ? AND kategori_kegiatan_id != ?", req.KategoriKegiatanNama, kategoriID).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori kegiatan dengan nama tersebut sudah ada",
		})
		return
	}

	// Update kategori menggunakan GORM Save (AMAN)
	kategori.KategoriKegiatanNama = req.KategoriKegiatanNama
	kategori.UpdatedAt = time.Now()

	if err := kc.db.Save(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate kategori kegiatan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk kegiatans
	if err := kc.db.Preload("Kegiatans").First(&kategori, kategoriID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kategori kegiatan yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori kegiatan berhasil diupdate",
		"data":    kategori,
	})
}

// ✅ DELETE - Menghapus kategori kegiatan
func (kc *KategoriKegiatanController) DeleteKategoriKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori kegiatan tidak valid",
		})
		return
	}

	var kategori models.KategoriKegiatan
	if err := kc.db.Preload("Kegiatans").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori kegiatan",
			})
		}
		return
	}

	// Check jika kategori memiliki kegiatan
	if len(kategori.Kegiatans) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak dapat menghapus kategori yang memiliki kegiatan",
			"details": gin.H{
				"total_kegiatan": len(kategori.Kegiatans),
			},
		})
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := kc.db.Delete(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus kategori kegiatan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori kegiatan berhasil dihapus",
	})
}

// ✅ GET - Search kategori kegiatan by nama (AMAN - parameterized query)
func (kc *KategoriKegiatanController) SearchKategoriKegiatan(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var kategori []models.KategoriKegiatan

	// Execute search query dengan GORM (AMAN - parameterized)
	if err := kc.db.
		Where("kategori_kegiatan_nama LIKE ?", "%"+search+"%").
		Preload("Kegiatans").
		Limit(20).
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mencari kategori kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         kategori,
		"total":        len(kategori),
		"search_query": search,
	})
}

// ✅ GET - Mendapatkan kategori kegiatan dengan statistik (AMAN - menggunakan GORM Model dan Joins)
func (kc *KategoriKegiatanController) GetKategoriKegiatanWithStats(c *gin.Context) {
	type KategoriWithStats struct {
		models.KategoriKegiatan
		TotalKegiatan int `json:"total_kegiatan"`
	}

	var results []KategoriWithStats

	// Query aman menggunakan GORM Joins
	if err := kc.db.
		Model(&models.KategoriKegiatan{}).
		Select("kategori_kegiatan.*, COUNT(kegiatans.kegiatan_id) as total_kegiatan").
		Joins("LEFT JOIN kegiatans ON kegiatans.kategori_kegiatan_id = kategori_kegiatan.kategori_kegiatan_id").
		Group("kategori_kegiatan.kategori_kegiatan_id").
		Order("total_kegiatan DESC").
		Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil statistik kategori kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
	})
}

// ✅ GET - Mendapatkan kategori kegiatan dropdown (simple list)
func (kc *KategoriKegiatanController) GetKategoriKegiatanDropdown(c *gin.Context) {
	var kategori []struct {
		KategoriKegiatanID   uint   `json:"kategori_kegiatan_id"`
		KategoriKegiatanNama string `json:"kategori_kegiatan_nama"`
	}

	// Query aman - hanya mengambil field yang diperlukan
	if err := kc.db.
		Model(&models.KategoriKegiatan{}).
		Select("kategori_kegiatan_id, kategori_kegiatan_nama").
		Order("kategori_kegiatan_nama ASC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dropdown kategori kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}