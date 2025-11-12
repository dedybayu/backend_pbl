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

type ProdukController struct {
	db *gorm.DB
}

func NewProdukController(db *gorm.DB) *ProdukController {
	return &ProdukController{db: db}
}

// Request structs - ubah binding dari json menjadi form
type CreateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama" binding:"required"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok" binding:"required"`
	ProdukHarga      float64 `form:"produk_harga" binding:"required"`
	ProdukFoto       string  `form:"produk_foto" binding:"required"`
	KategoriProdukID uint    `form:"kategori_produk_id" binding:"required"`
}

type UpdateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok"`
	ProdukHarga      float64 `form:"produk_harga"`
	ProdukFoto       string  `form:"produk_foto"`
	KategoriProdukID uint    `form:"kategori_produk_id"`
}

// ✅ CREATE - Membuat produk baru (FORM DATA)
func (pc *ProdukController) CreateProduk(c *gin.Context) {
	var req CreateProdukRequest
	
	// Gunakan ShouldBind instead of ShouldBindJSON untuk form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.ProdukNama = strings.TrimSpace(req.ProdukNama)
	req.ProdukDeskripsi = strings.TrimSpace(req.ProdukDeskripsi)
	req.ProdukFoto = strings.TrimSpace(req.ProdukFoto)

	// Validasi required fields
	if req.ProdukNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama produk harus diisi",
		})
		return
	}

	if req.ProdukFoto == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Foto produk harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.ProdukNama) < 2 || len(req.ProdukNama) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama produk harus 2-200 karakter",
		})
		return
	}

	// Validasi stok
	if req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Validasi harga
	if req.ProdukHarga <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harga produk harus lebih dari 0",
		})
		return
	}

	// Check if kategori produk exists
	var kategori models.KategoriProduk
	if err := pc.db.First(&kategori, req.KategoriProdukID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi kategori produk",
			})
		}
		return
	}

	// Check if produk dengan nama yang sama sudah ada
	var existingProduk models.Produk
	if err := pc.db.Where("produk_nama = ?", req.ProdukNama).First(&existingProduk).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Produk dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat produk baru
	produk := models.Produk{
		ProdukNama:       req.ProdukNama,
		ProdukDeskripsi:  req.ProdukDeskripsi,
		ProdukStok:       req.ProdukStok,
		ProdukHarga:      req.ProdukHarga,
		ProdukFoto:       req.ProdukFoto,
		KategoriProdukID: req.KategoriProdukID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := pc.db.Create(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori
	if err := pc.db.Preload("KategoriProduk").First(&produk, produk.ProdukID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil dibuat",
		"data":    produk,
	})
}

// ✅ READ - Mendapatkan semua produk (TETAP SAMA - GET request)
func (pc *ProdukController) GetAllProduk(c *gin.Context) {
	var produk []models.Produk

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	kategoriID := c.Query("kategori_id")
	search := c.Query("search")
	stokMin := c.Query("stok_min")
	stokMax := c.Query("stok_max")
	hargaMin := c.Query("harga_min")
	hargaMax := c.Query("harga_max")

	// Build query dengan GORM (AMAN - parameterized queries)
	query := pc.db.Model(&models.Produk{}).Preload("KategoriProduk")

	// Apply filters
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("produk_nama LIKE ? OR produk_deskripsi LIKE ?", 
			"%"+searchSafe+"%", "%"+searchSafe+"%")
	}

	if kategoriID != "" {
		kategoriIDSafe, err := strconv.ParseUint(kategoriID, 10, 32)
		if err == nil {
			query = query.Where("kategori_produk_id = ?", kategoriIDSafe)
		}
	}

	if stokMin != "" {
		if stokMinSafe, err := strconv.Atoi(stokMin); err == nil {
			query = query.Where("produk_stok >= ?", stokMinSafe)
		}
	}

	if stokMax != "" {
		if stokMaxSafe, err := strconv.Atoi(stokMax); err == nil {
			query = query.Where("produk_stok <= ?", stokMaxSafe)
		}
	}

	if hargaMin != "" {
		if hargaMinSafe, err := strconv.ParseFloat(hargaMin, 64); err == nil {
			query = query.Where("produk_harga >= ?", hargaMinSafe)
		}
	}

	if hargaMax != "" {
		if hargaMaxSafe, err := strconv.ParseFloat(hargaMax, 64); err == nil {
			query = query.Where("produk_harga <= ?", hargaMaxSafe)
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": produk,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan produk by ID (TETAP SAMA - GET request)
func (pc *ProdukController) GetProdukByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data produk",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": produk,
	})
}

// ✅ UPDATE - Mengupdate produk (FORM DATA)
func (pc *ProdukController) UpdateProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	var req UpdateProdukRequest
	// Gunakan ShouldBind untuk form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.ProdukNama != "" {
		req.ProdukNama = strings.TrimSpace(req.ProdukNama)
	}
	if req.ProdukDeskripsi != "" {
		req.ProdukDeskripsi = strings.TrimSpace(req.ProdukDeskripsi)
	}
	if req.ProdukFoto != "" {
		req.ProdukFoto = strings.TrimSpace(req.ProdukFoto)
	}

	// Validasi jika nama diupdate
	if req.ProdukNama != "" {
		if len(req.ProdukNama) < 2 || len(req.ProdukNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama produk harus 2-200 karakter",
			})
			return
		}

		// Check duplicate name (exclude current)
		var existingProduk models.Produk
		if err := pc.db.Where("produk_nama = ? AND produk_id != ?", req.ProdukNama, produkID).
			First(&existingProduk).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Produk dengan nama tersebut sudah ada",
			})
			return
		}
	}

	// Validasi stok jika diupdate
	if req.ProdukStok != 0 && req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Validasi harga jika diupdate
	if req.ProdukHarga != 0 && req.ProdukHarga <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harga produk harus lebih dari 0",
		})
		return
	}

	// Validasi kategori produk jika diupdate
	if req.KategoriProdukID != 0 {
		var kategori models.KategoriProduk
		if err := pc.db.First(&kategori, req.KategoriProdukID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Kategori produk tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi kategori produk",
				})
			}
			return
		}
	}

	// Update fields menggunakan map (AMAN - GORM Updates dengan map)
	updates := make(map[string]interface{})
	
	if req.ProdukNama != "" {
		updates["produk_nama"] = req.ProdukNama
	}
	if req.ProdukDeskripsi != "" {
		updates["produk_deskripsi"] = req.ProdukDeskripsi
	}
	if req.ProdukStok != 0 {
		updates["produk_stok"] = req.ProdukStok
	}
	if req.ProdukHarga != 0 {
		updates["produk_harga"] = req.ProdukHarga
	}
	if req.ProdukFoto != "" {
		updates["produk_foto"] = req.ProdukFoto
	}
	if req.KategoriProdukID != 0 {
		updates["kategori_produk_id"] = req.KategoriProdukID
	}
	
	updates["updated_at"] = time.Now()

	if err := pc.db.Model(&produk).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk kategori
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil diupdate",
		"data":    produk,
	})
}

// ✅ DELETE - Menghapus produk (TETAP SAMA - DELETE request)
func (pc *ProdukController) DeleteProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := pc.db.Delete(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus produk",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil dihapus",
	})
}

// ✅ GET - Produk terbaru (TETAP SAMA - GET request)
func (pc *ProdukController) GetProdukTerbaru(c *gin.Context) {
	var produk []models.Produk

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))

	// Query produk terbaru (AMAN)
	if err := pc.db.
		Preload("KategoriProduk").
		Order("created_at DESC").
		Limit(limit).
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk terbaru",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  produk,
		"total": len(produk),
	})
}

// ✅ GET - Produk dengan stok menipis (TETAP SAMA - GET request)
func (pc *ProdukController) GetProdukStokMenipis(c *gin.Context) {
	var produk []models.Produk

	// Query produk dengan stok menipis (AMAN)
	if err := pc.db.
		Preload("KategoriProduk").
		Where("produk_stok < 10").
		Order("produk_stok ASC").
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk stok menipis",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  produk,
		"total": len(produk),
	})
}

// ✅ GET - Statistik produk (TETAP SAMA - GET request)
func (pc *ProdukController) GetStatistikProduk(c *gin.Context) {
	type StatistikResult struct {
		TotalProduk      int64   `json:"total_produk"`
		TotalStok        int64   `json:"total_stok"`
		NilaiInventori   float64 `json:"nilai_inventori"`
		ProdukStokHabis  int64   `json:"produk_stok_habis"`
		ProdukStokMenipis int64  `json:"produk_stok_menipis"`
	}

	var statistik StatistikResult

	// Hitung total produk (AMAN)
	pc.db.Model(&models.Produk{}).Count(&statistik.TotalProduk)

	// Hitung total stok (AMAN)
	pc.db.Model(&models.Produk{}).
		Select("COALESCE(SUM(produk_stok), 0)").
		Row().
		Scan(&statistik.TotalStok)

	// Hitung nilai inventori (AMAN)
	pc.db.Model(&models.Produk{}).
		Select("COALESCE(SUM(produk_stok * produk_harga), 0)").
		Row().
		Scan(&statistik.NilaiInventori)

	// Hitung produk stok habis (AMAN)
	pc.db.Model(&models.Produk{}).
		Where("produk_stok = 0").
		Count(&statistik.ProdukStokHabis)

	// Hitung produk stok menipis (AMAN)
	pc.db.Model(&models.Produk{}).
		Where("produk_stok > 0 AND produk_stok < 10").
		Count(&statistik.ProdukStokMenipis)

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ PATCH - Update stok produk (FORM DATA)
func (pc *ProdukController) UpdateStokProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	type UpdateStokRequest struct {
		ProdukStok int `form:"produk_stok" binding:"required"`
	}

	var req UpdateStokRequest
	// Gunakan ShouldBind untuk form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Validasi stok
	if req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Update stok
	produk.ProdukStok = req.ProdukStok
	produk.UpdatedAt = time.Now()

	if err := pc.db.Save(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate stok produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stok produk berhasil diupdate",
		"data":    produk,
	})
}