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
	KategoriKegiatanNama string `form:"kategori_kegiatan_nama" binding:"required" example:"Bakti Sosial"`
}

type UpdateKategoriKegiatanRequest struct {
	KategoriKegiatanNama string `form:"kategori_kegiatan_nama" binding:"required" example:"Kerja Bakti"`
}

// CreateKategoriKegiatan godoc
// @Summary      Membuat kategori kegiatan baru
// @Description  Membuat kategori kegiatan baru dengan nama yang unik
// @Tags         kategori-kegiatan
// @Accept       multipart/form-data
// @Produce      json
// @Param        kategori_kegiatan_nama formData string true "Nama kategori kegiatan"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "Kategori kegiatan berhasil dibuat"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      500  {object}  map[string]interface{}  "Gagal membuat kategori kegiatan"
// @Router       /api/kategori-kegiatan [post]
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

	// Check if kategori dengan nama yang sama sudah ada
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

// GetAllKategoriKegiatan godoc
// @Summary      Mendapatkan semua kategori kegiatan
// @Description  Mendapatkan daftar semua kategori kegiatan tanpa pagination limit
// @Tags         kategori-kegiatan
// @Produce      json
// @Param        search query string false "Kata kunci pencarian"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar kategori kegiatan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kategori kegiatan"
// @Router       /api/kategori-kegiatan [get]
func (kc *KategoriKegiatanController) GetAllKategoriKegiatan(c *gin.Context) {
	var kategori []models.KategoriKegiatan

	// Search
	search := strings.TrimSpace(c.Query("search"))

	query := kc.db.Model(&models.KategoriKegiatan{})

	if search != "" {
		query = query.Where("kategori_kegiatan_nama LIKE ?", "%"+search+"%")
	}

	// Tidak ada pagination limit, hanya order by
	if err := query.
		Order("created_at DESC").
		Find(&kategori).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kategori kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kategori,
		"total": len(kategori),
	})
}

// GetKategoriKegiatanByID godoc
// @Summary      Mendapatkan kategori kegiatan berdasarkan ID
// @Description  Mendapatkan detail kategori kegiatan beserta daftar kegiatan terkait
// @Tags         kategori-kegiatan
// @Produce      json
// @Param        id   path      int  true  "ID Kategori Kegiatan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Detail kategori kegiatan"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Kategori kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kategori kegiatan"
// @Router       /api/kategori-kegiatan/{id} [get]
func (kc *KategoriKegiatanController) GetKategoriKegiatanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

// UpdateKategoriKegiatan godoc
// @Summary      Mengupdate kategori kegiatan
// @Description  Mengupdate nama kategori kegiatan berdasarkan ID
// @Tags         kategori-kegiatan
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "ID Kategori Kegiatan"
// @Param        kategori_kegiatan_nama formData string true "Nama kategori kegiatan baru"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Kategori kegiatan berhasil diupdate"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      404  {object}  map[string]interface{}  "Kategori kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengupdate kategori kegiatan"
// @Router       /api/kategori-kegiatan/{id} [put]
func (kc *KategoriKegiatanController) UpdateKategoriKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

	// Check if kategori dengan nama yang sama sudah ada (exclude current)
	var existingKategori models.KategoriKegiatan
	if err := kc.db.Where("kategori_kegiatan_nama = ? AND kategori_kegiatan_id != ?", req.KategoriKegiatanNama, kategoriID).First(&existingKategori).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori kegiatan dengan nama tersebut sudah ada",
		})
		return
	}

	// Update kategori
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

// DeleteKategoriKegiatan godoc
// @Summary      Menghapus kategori kegiatan
// @Description  Menghapus kategori kegiatan berdasarkan ID, hanya jika tidak memiliki kegiatan terkait
// @Tags         kategori-kegiatan
// @Produce      json
// @Param        id   path      int  true  "ID Kategori Kegiatan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Kategori kegiatan berhasil dihapus"
// @Failure      400  {object}  map[string]interface{}  "Tidak dapat menghapus kategori yang memiliki kegiatan"
// @Failure      404  {object}  map[string]interface{}  "Kategori kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal menghapus kategori kegiatan"
// @Router       /api/kategori-kegiatan/{id} [delete]
func (kc *KategoriKegiatanController) DeleteKategoriKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

	// Delete
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

// SearchKategoriKegiatan godoc
// @Summary      Mencari kategori kegiatan
// @Description  Mencari kategori kegiatan berdasarkan nama
// @Tags         kategori-kegiatan
// @Produce      json
// @Param        q   query     string  true  "Kata kunci pencarian"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Hasil pencarian kategori kegiatan"
// @Failure      400  {object}  map[string]interface{}  "Parameter pencarian harus diisi"
// @Failure      500  {object}  map[string]interface{}  "Gagal mencari kategori kegiatan"
// @Router       /api/kategori-kegiatan/search [get]
func (kc *KategoriKegiatanController) SearchKategoriKegiatan(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var kategori []models.KategoriKegiatan

	// Execute search query
	if err := kc.db.
		Where("kategori_kegiatan_nama LIKE ?", "%"+search+"%").
		Preload("Kegiatans").
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

// GetKategoriKegiatanWithStats godoc
// @Summary      Mendapatkan kategori kegiatan dengan statistik
// @Description  Mendapatkan kategori kegiatan beserta jumlah kegiatan di setiap kategori
// @Tags         kategori-kegiatan
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Kategori kegiatan dengan statistik"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil statistik kategori kegiatan"
// @Router       /api/kategori-kegiatan/stats [get]
func (kc *KategoriKegiatanController) GetKategoriKegiatanWithStats(c *gin.Context) {
	type KategoriWithStats struct {
		models.KategoriKegiatan
		TotalKegiatan int `json:"total_kegiatan"`
	}

	var results []KategoriWithStats

	// Query menggunakan GORM Joins
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

// GetKategoriKegiatanDropdown godoc
// @Summary      Mendapatkan dropdown kategori kegiatan
// @Description  Mendapatkan daftar kategori kegiatan sederhana untuk dropdown/select
// @Tags         kategori-kegiatan
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Dropdown kategori kegiatan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data dropdown"
// @Router       /api/kategori-kegiatan/dropdown [get]
func (kc *KategoriKegiatanController) GetKategoriKegiatanDropdown(c *gin.Context) {
	var kategori []struct {
		KategoriKegiatanID   uint   `json:"kategori_kegiatan_id"`
		KategoriKegiatanNama string `json:"kategori_kegiatan_nama"`
	}

	// Query - hanya mengambil field yang diperlukan
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