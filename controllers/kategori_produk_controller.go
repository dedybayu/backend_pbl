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

type KategoriProdukController struct {
	db *gorm.DB
}

func NewKategoriProdukController(db *gorm.DB) *KategoriProdukController {
	return &KategoriProdukController{db: db}
}

// Request structs
type CreateKategoriProdukRequest struct {
	KategoriProdukNama string `json:"kategori_produk_nama" binding:"required"`
}

type UpdateKategoriProdukRequest struct {
	KategoriProdukNama string `json:"kategori_produk_nama" binding:"required"`
}

// ✅ CREATE - Membuat kategori produk baru
func (kpc *KategoriProdukController) CreateKategoriProduk(c *gin.Context) {
	var req CreateKategoriProdukRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriProdukNama = strings.TrimSpace(req.KategoriProdukNama)

	// Validasi required fields
	if req.KategoriProdukNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori produk harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriProdukNama) < 2 || len(req.KategoriProdukNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori produk harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada
	var existingKategori models.KategoriProduk
	if err := kpc.db.Where("kategori_produk_nama = ?", req.KategoriProdukNama).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori produk dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat kategori produk baru
	kategori := models.KategoriProduk{
		KategoriProdukNama: req.KategoriProdukNama,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := kpc.db.Create(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat kategori produk",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kategori produk berhasil dibuat",
		"data":    kategori,
	})
}

// ✅ READ - Mendapatkan semua kategori produk
func (kpc *KategoriProdukController) GetAllKategoriProduk(c *gin.Context) {
	var kategori []models.KategoriProduk

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := kpc.db.Model(&models.KategoriProduk{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("kategori_produk_nama LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Preload("Produks").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori produk",
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

// ✅ READ - Mendapatkan kategori produk by ID
func (kpc *KategoriProdukController) GetKategoriProdukByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori produk tidak valid",
		})
		return
	}

	var kategori models.KategoriProduk
	if err := kpc.db.Preload("Produks").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data kategori produk",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}

// ✅ UPDATE - Mengupdate kategori produk
func (kpc *KategoriProdukController) UpdateKategoriProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori produk tidak valid",
		})
		return
	}

	var kategori models.KategoriProduk
	if err := kpc.db.First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori produk",
			})
		}
		return
	}

	var req UpdateKategoriProdukRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KategoriProdukNama = strings.TrimSpace(req.KategoriProdukNama)

	// Validasi required fields
	if req.KategoriProdukNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori produk harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KategoriProdukNama) < 2 || len(req.KategoriProdukNama) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kategori produk harus 2-100 karakter",
		})
		return
	}

	// Check if kategori dengan nama yang sama sudah ada (exclude current)
	var existingKategori models.KategoriProduk
	if err := kpc.db.Where("kategori_produk_nama = ? AND kategori_produk_id != ?", req.KategoriProdukNama, kategoriID).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori produk dengan nama tersebut sudah ada",
		})
		return
	}

	// Update kategori
	kategori.KategoriProdukNama = req.KategoriProdukNama
	kategori.UpdatedAt = time.Now()

	if err := kpc.db.Save(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate kategori produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk produk
	if err := kpc.db.Preload("Produks").First(&kategori, kategoriID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kategori produk yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori produk berhasil diupdate",
		"data":    kategori,
	})
}

// ✅ DELETE - Menghapus kategori produk
func (kpc *KategoriProdukController) DeleteKategoriProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kategoriID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kategori produk tidak valid",
		})
		return
	}

	var kategori models.KategoriProduk
	if err := kpc.db.Preload("Produks").First(&kategori, kategoriID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kategori produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kategori produk",
			})
		}
		return
	}

	// Check jika kategori memiliki produk
	if len(kategori.Produks) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tidak dapat menghapus kategori yang memiliki produk",
			"details": gin.H{
				"total_produk": len(kategori.Produks),
			},
		})
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := kpc.db.Delete(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus kategori produk",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kategori produk berhasil dihapus",
	})
}

// ✅ GET - Dropdown kategori produk
func (kpc *KategoriProdukController) GetKategoriProdukDropdown(c *gin.Context) {
	var kategori []struct {
		KategoriProdukID   uint   `json:"kategori_produk_id"`
		KategoriProdukNama string `json:"kategori_produk_nama"`
	}

	// Query aman - hanya mengambil field yang diperlukan
	if err := kpc.db.
		Model(&models.KategoriProduk{}).
		Select("kategori_produk_id, kategori_produk_nama").
		Order("kategori_produk_nama ASC").
		Find(&kategori).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dropdown kategori produk",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
	})
}

// ✅ GET - Statistik kategori produk
func (kpc *KategoriProdukController) GetStatistikKategoriProduk(c *gin.Context) {
	type StatistikResult struct {
		TotalKategori int64 `json:"total_kategori"`
		TotalProduk   int64 `json:"total_produk"`
		KategoriDenganProdukTerbanyak string `json:"kategori_dengan_produk_terbanyak"`
		JumlahProdukTerbanyak int64 `json:"jumlah_produk_terbanyak"`
	}

	var statistik StatistikResult

	// Hitung total kategori (AMAN)
	kpc.db.Model(&models.KategoriProduk{}).Count(&statistik.TotalKategori)

	// Hitung total produk (AMAN)
	kpc.db.Model(&models.Produk{}).Count(&statistik.TotalProduk)

	// Cari kategori dengan produk terbanyak
	type KategoriProdukCount struct {
		KategoriProdukNama string
		JumlahProduk      int64
	}

	var kategoriTerbanyak KategoriProdukCount
	kpc.db.Model(&models.KategoriProduk{}).
		Select("kategori_produk_nama, COUNT(produks.produk_id) as jumlah_produk").
		Joins("LEFT JOIN produks ON produks.kategori_produk_id = kategori_produks.kategori_produk_id").
		Group("kategori_produks.kategori_produk_id, kategori_produks.kategori_produk_nama").
		Order("jumlah_produk DESC").
		Limit(1).
		Scan(&kategoriTerbanyak)

	statistik.KategoriDenganProdukTerbanyak = kategoriTerbanyak.KategoriProdukNama
	statistik.JumlahProdukTerbanyak = kategoriTerbanyak.JumlahProduk

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}