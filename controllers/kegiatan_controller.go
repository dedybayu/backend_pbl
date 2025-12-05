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

type KegiatanController struct {
	db *gorm.DB
}

func NewKegiatanController(db *gorm.DB) *KegiatanController {
	return &KegiatanController{db: db}
}

// Request structs
type CreateKegiatanRequest struct {
	KegiatanNama       string    `form:"kegiatan_nama" binding:"required"`
	KategoriKegiatanID uint      `form:"kategori_kegiatan_id" binding:"required"`
	KegiatanTanggal    time.Time `form:"kegiatan_tanggal" binding:"required"`
	KegiatanLokasi     string    `form:"kegiatan_lokasi"`
	KegiatanPJ         string    `form:"kegiatan_pj"`
	KegiatanDeskripsi  string    `form:"kegiatan_deskripsi"`
}

type UpdateKegiatanRequest struct {
	KegiatanNama       string    `form:"kegiatan_nama"`
	KategoriKegiatanID uint      `form:"kategori_kegiatan_id"`
	KegiatanTanggal    time.Time `form:"kegiatan_tanggal"`
	KegiatanLokasi     string    `form:"kegiatan_lokasi"`
	KegiatanPJ         string    `form:"kegiatan_pj"`
	KegiatanDeskripsi  string    `form:"kegiatan_deskripsi"`
}

// ✅ CREATE - Membuat kegiatan baru
func (kc *KegiatanController) CreateKegiatan(c *gin.Context) {
	var req CreateKegiatanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.KegiatanNama = strings.TrimSpace(req.KegiatanNama)
	req.KegiatanLokasi = strings.TrimSpace(req.KegiatanLokasi)
	req.KegiatanPJ = strings.TrimSpace(req.KegiatanPJ)
	req.KegiatanDeskripsi = strings.TrimSpace(req.KegiatanDeskripsi)

	// Validasi required fields
	if req.KegiatanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kegiatan harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.KegiatanNama) < 2 || len(req.KegiatanNama) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama kegiatan harus 2-200 karakter",
		})
		return
	}

	// Validasi tanggal tidak boleh lebih kecil dari hari ini
	if req.KegiatanTanggal.Before(time.Now().AddDate(0, 0, -1)) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tanggal kegiatan tidak boleh lebih kecil dari hari ini",
		})
		return
	}

	// Check if kategori kegiatan exists
	var kategori models.KategoriKegiatan
	if err := kc.db.First(&kategori, req.KategoriKegiatanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi kategori kegiatan",
			})
		}
		return
	}

	// Check if kegiatan dengan nama yang sama sudah ada dalam rentang waktu yang sama
	var existingKegiatan models.Kegiatan
	if err := kc.db.Where("kegiatan_nama = ? AND DATE(kegiatan_tanggal) = ?", 
		req.KegiatanNama, 
		req.KegiatanTanggal.Format("2006-01-02")).
		First(&existingKegiatan).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kegiatan dengan nama tersebut sudah ada pada tanggal yang sama",
		})
		return
	}

	// Buat kegiatan baru
	kegiatan := models.Kegiatan{
		KegiatanNama:       req.KegiatanNama,
		KategoriKegiatanID: req.KategoriKegiatanID,
		KegiatanTanggal:    req.KegiatanTanggal,
		KegiatanLokasi:     req.KegiatanLokasi,
		KegiatanPJ:         req.KegiatanPJ,
		KegiatanDeskripsi:  req.KegiatanDeskripsi,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := kc.db.Create(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat kegiatan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori
	if err := kc.db.Preload("KategoriKegiatan").First(&kegiatan, kegiatan.KegiatanID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kegiatan yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Kegiatan berhasil dibuat",
		"data":    kegiatan,
	})
}

// ✅ READ - Mendapatkan semua kegiatan dengan filter
func (kc *KegiatanController) GetAllKegiatan(c *gin.Context) {
	var kegiatan []models.Kegiatan

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	kategoriID := c.Query("kategori_id")
	tanggalFrom := c.Query("tanggal_from")
	tanggalTo := c.Query("tanggal_to")
	search := c.Query("search")

	// Build query dengan GORM (AMAN - parameterized queries)
	query := kc.db.Model(&models.Kegiatan{}).Preload("KategoriKegiatan")

	// Apply search filter
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("kegiatan_nama LIKE ? OR kegiatan_lokasi LIKE ? OR kegiatan_pj LIKE ?", 
			"%"+searchSafe+"%", "%"+searchSafe+"%", "%"+searchSafe+"%")
	}

	// Apply kategori filter
	if kategoriID != "" {
		kategoriIDSafe, err := strconv.ParseUint(kategoriID, 10, 32)
		if err == nil {
			query = query.Where("kategori_kegiatan_id = ?", kategoriIDSafe)
		}
	}

	// Apply tanggal filter
	if tanggalFrom != "" {
		if tanggalFromSafe, err := time.Parse("2006-01-02", tanggalFrom); err == nil {
			query = query.Where("DATE(kegiatan_tanggal) >= ?", tanggalFromSafe.Format("2006-01-02"))
		}
	}

	if tanggalTo != "" {
		if tanggalToSafe, err := time.Parse("2006-01-02", tanggalTo); err == nil {
			query = query.Where("DATE(kegiatan_tanggal) <= ?", tanggalToSafe.Format("2006-01-02"))
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("kegiatan_tanggal DESC, created_at DESC").
		Find(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kegiatan,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan kegiatan by ID
func (kc *KegiatanController) GetKegiatanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kegiatanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kegiatan tidak valid",
		})
		return
	}

	var kegiatan models.Kegiatan
	if err := kc.db.Preload("KategoriKegiatan").First(&kegiatan, kegiatanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data kegiatan",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": kegiatan,
	})
}

// ✅ UPDATE - Mengupdate kegiatan
func (kc *KegiatanController) UpdateKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kegiatanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kegiatan tidak valid",
		})
		return
	}

	var kegiatan models.Kegiatan
	if err := kc.db.First(&kegiatan, kegiatanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kegiatan",
			})
		}
		return
	}

	var req UpdateKegiatanRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.KegiatanNama != "" {
		req.KegiatanNama = strings.TrimSpace(req.KegiatanNama)
	}
	if req.KegiatanLokasi != "" {
		req.KegiatanLokasi = strings.TrimSpace(req.KegiatanLokasi)
	}
	if req.KegiatanPJ != "" {
		req.KegiatanPJ = strings.TrimSpace(req.KegiatanPJ)
	}
	if req.KegiatanDeskripsi != "" {
		req.KegiatanDeskripsi = strings.TrimSpace(req.KegiatanDeskripsi)
	}

	// Validasi jika nama diupdate
	if req.KegiatanNama != "" {
		if len(req.KegiatanNama) < 2 || len(req.KegiatanNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama kegiatan harus 2-200 karakter",
			})
			return
		}

		// Check duplicate name dengan tanggal yang sama
		tanggalUntukValidasi := kegiatan.KegiatanTanggal
		if !req.KegiatanTanggal.IsZero() {
			tanggalUntukValidasi = req.KegiatanTanggal
		}

		var existingKegiatan models.Kegiatan
		if err := kc.db.Where("kegiatan_nama = ? AND DATE(kegiatan_tanggal) = ? AND kegiatan_id != ?", 
			req.KegiatanNama, 
			tanggalUntukValidasi.Format("2006-01-02"),
			kegiatanID).
			First(&existingKegiatan).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kegiatan dengan nama tersebut sudah ada pada tanggal yang sama",
			})
			return
		}
	}

	// Validasi tanggal jika diupdate
	if !req.KegiatanTanggal.IsZero() {
		if req.KegiatanTanggal.Before(time.Now().AddDate(0, 0, -1)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tanggal kegiatan tidak boleh lebih kecil dari hari ini",
			})
			return
		}
	}

	// Validasi kategori kegiatan jika diupdate
	if req.KategoriKegiatanID != 0 {
		var kategori models.KategoriKegiatan
		if err := kc.db.First(&kategori, req.KategoriKegiatanID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Kategori kegiatan tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi kategori kegiatan",
				})
			}
			return
		}
	}

	// Update fields menggunakan map (AMAN - GORM Updates dengan map)
	updates := make(map[string]interface{})
	
	if req.KegiatanNama != "" {
		updates["kegiatan_nama"] = req.KegiatanNama
	}
	if req.KategoriKegiatanID != 0 {
		updates["kategori_kegiatan_id"] = req.KategoriKegiatanID
	}
	if !req.KegiatanTanggal.IsZero() {
		updates["kegiatan_tanggal"] = req.KegiatanTanggal
	}
	if req.KegiatanLokasi != "" {
		updates["kegiatan_lokasi"] = req.KegiatanLokasi
	}
	if req.KegiatanPJ != "" {
		updates["kegiatan_pj"] = req.KegiatanPJ
	}
	if req.KegiatanDeskripsi != "" {
		updates["kegiatan_deskripsi"] = req.KegiatanDeskripsi
	}
	
	updates["updated_at"] = time.Now()

	if err := kc.db.Model(&kegiatan).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate kegiatan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk kategori
	if err := kc.db.Preload("KategoriKegiatan").First(&kegiatan, kegiatanID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data kegiatan yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kegiatan berhasil diupdate",
		"data":    kegiatan,
	})
}

// ✅ DELETE - Menghapus kegiatan
func (kc *KegiatanController) DeleteKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	kegiatanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID kegiatan tidak valid",
		})
		return
	}

	var kegiatan models.Kegiatan
	if err := kc.db.First(&kegiatan, kegiatanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Kegiatan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan kegiatan",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := kc.db.Delete(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus kegiatan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Kegiatan berhasil dihapus",
	})
}

// ✅ GET - Mendapatkan kegiatan mendatang
func (kc *KegiatanController) GetKegiatanMendatang(c *gin.Context) {
	var kegiatan []models.Kegiatan

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// Query kegiatan dengan tanggal >= hari ini (AMAN)
	if err := kc.db.
		Preload("KategoriKegiatan").
		Where("kegiatan_tanggal >= ?", time.Now().Format("2006-01-02")).
		Order("kegiatan_tanggal ASC").
		Limit(limit).
		Find(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kegiatan mendatang",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  kegiatan,
		"total": len(kegiatan),
	})
}

// ✅ GET - Mendapatkan kegiatan by bulan dan tahun
func (kc *KegiatanController) GetKegiatanByBulanTahun(c *gin.Context) {
	bulan := c.Param("bulan")
	tahun := c.Param("tahun")

	// Validasi bulan dan tahun
	bulanInt, err := strconv.Atoi(bulan)
	if err != nil || bulanInt < 1 || bulanInt > 12 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Bulan harus antara 1-12",
		})
		return
	}

	tahunInt, err := strconv.Atoi(tahun)
	if err != nil || tahunInt < 2000 || tahunInt > 2100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tahun tidak valid",
		})
		return
	}

	// Buat tanggal awal dan akhir bulan
	awalBulan := time.Date(tahunInt, time.Month(bulanInt), 1, 0, 0, 0, 0, time.UTC)
	akhirBulan := awalBulan.AddDate(0, 1, -1)

	var kegiatan []models.Kegiatan

	// Query kegiatan by bulan dan tahun (AMAN - parameterized)
	if err := kc.db.
		Preload("KategoriKegiatan").
		Where("kegiatan_tanggal BETWEEN ? AND ?", awalBulan.Format("2006-01-02"), akhirBulan.Format("2006-01-02")).
		Order("kegiatan_tanggal ASC").
		Find(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  kegiatan,
		"total": len(kegiatan),
		"bulan": bulanInt,
		"tahun": tahunInt,
	})
}

// ✅ GET - Statistik kegiatan per bulan
func (kc *KegiatanController) GetStatistikKegiatan(c *gin.Context) {
	type StatistikBulanan struct {
		Bulan         string `form:"bulan"`
		Tahun         int    `form:"tahun"`
		BulanAngka    int    `form:"bulan_angka"`
		TotalKegiatan int    `form:"total_kegiatan"`
	}

	var statistik []StatistikBulanan

	// Hitung 6 bulan terakhir
	sekarang := time.Now()
	for i := 0; i < 6; i++ {
		tanggal := sekarang.AddDate(0, -i, 0)
		bulan := int(tanggal.Month())
		tahun := tanggal.Year()

		var total int64
		awalBulan := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.UTC)
		akhirBulan := awalBulan.AddDate(0, 1, -1)

		// Hitung total kegiatan per bulan (AMAN - parameterized query)
		kc.db.Model(&models.Kegiatan{}).
			Where("kegiatan_tanggal BETWEEN ? AND ?", awalBulan, akhirBulan).
			Count(&total)

		statistik = append(statistik, StatistikBulanan{
			Bulan:         tanggal.Format("January 2006"),
			Tahun:         tahun,
			BulanAngka:    bulan,
			TotalKegiatan: int(total),
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ GET - Pencarian kegiatan
func (kc *KegiatanController) SearchKegiatan(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var kegiatan []models.Kegiatan

	// Execute search query dengan GORM (AMAN - parameterized)
	if err := kc.db.
		Preload("KategoriKegiatan").
		Where("kegiatan_nama LIKE ? OR kegiatan_lokasi LIKE ? OR kegiatan_pj LIKE ? OR kegiatan_deskripsi LIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
		Order("kegiatan_tanggal DESC").
		Limit(20).
		Find(&kegiatan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mencari kegiatan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         kegiatan,
		"total":        len(kegiatan),
		"search_query": search,
	})
}