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
	KegiatanNama       string    `form:"kegiatan_nama" binding:"required" example:"Kerja Bakti Mingguan"`
	KategoriKegiatanID uint      `form:"kategori_kegiatan_id" binding:"required" example:"1"`
	KegiatanTanggal    time.Time `form:"kegiatan_tanggal" binding:"required" example:"2024-01-15T08:00:00Z"`
	KegiatanLokasi     string    `form:"kegiatan_lokasi" example:"Lapangan RT 01"`
	KegiatanPJ         string    `form:"kegiatan_pj" example:"Budi Santoso"`
	KegiatanDeskripsi  string    `form:"kegiatan_deskripsi" example:"Kerja bakti membersihkan lingkungan RT"`
}

type UpdateKegiatanRequest struct {
	KegiatanNama       string    `form:"kegiatan_nama" example:"Kerja Bakti Rutin"`
	KategoriKegiatanID uint      `form:"kategori_kegiatan_id" example:"2"`
	KegiatanTanggal    time.Time `form:"kegiatan_tanggal" example:"2024-01-20T08:00:00Z"`
	KegiatanLokasi     string    `form:"kegiatan_lokasi" example:"Taman RT 01"`
	KegiatanPJ         string    `form:"kegiatan_pj" example:"Siti Aminah"`
	KegiatanDeskripsi  string    `form:"kegiatan_deskripsi" example:"Kerja bakti menanam pohon di taman"`
}

// CreateKegiatan godoc
// @Summary      Membuat kegiatan baru
// @Description  Membuat kegiatan baru dengan data yang diberikan
// @Tags         kegiatan
// @Accept       multipart/form-data
// @Produce      json
// @Param        kegiatan_nama formData string true "Nama kegiatan"
// @Param        kategori_kegiatan_id formData integer true "ID kategori kegiatan"
// @Param        kegiatan_tanggal formData string true "Tanggal kegiatan (format: RFC3339)"
// @Param        kegiatan_lokasi formData string false "Lokasi kegiatan"
// @Param        kegiatan_pj formData string false "Penanggung jawab kegiatan"
// @Param        kegiatan_deskripsi formData string false "Deskripsi kegiatan"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "Kegiatan berhasil dibuat"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      500  {object}  map[string]interface{}  "Gagal membuat kegiatan"
// @Router       /api/kegiatan [post]
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

// GetAllKegiatan godoc
// @Summary      Mendapatkan semua kegiatan
// @Description  Mendapatkan daftar semua kegiatan dengan filter dan pagination
// @Tags         kegiatan
// @Produce      json
// @Param        page query int false "Halaman (default: 1)"
// @Param        limit query int false "Limit per halaman (default: 10)"
// @Param        search query string false "Kata kunci pencarian"
// @Param        kategori_id query int false "Filter berdasarkan ID kategori"
// @Param        tanggal_from query string false "Filter tanggal mulai (format: YYYY-MM-DD)"
// @Param        tanggal_to query string false "Filter tanggal sampai (format: YYYY-MM-DD)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar kegiatan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kegiatan"
// @Router       /api/kegiatan [get]
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

	// Build query dengan GORM
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

// GetKegiatanByID godoc
// @Summary      Mendapatkan kegiatan berdasarkan ID
// @Description  Mendapatkan detail kegiatan berdasarkan ID
// @Tags         kegiatan
// @Produce      json
// @Param        id   path      int  true  "ID Kegiatan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Detail kegiatan"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kegiatan"
// @Router       /api/kegiatan/{id} [get]
func (kc *KegiatanController) GetKegiatanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

// UpdateKegiatan godoc
// @Summary      Mengupdate kegiatan
// @Description  Mengupdate data kegiatan berdasarkan ID
// @Tags         kegiatan
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "ID Kegiatan"
// @Param        kegiatan_nama formData string false "Nama kegiatan"
// @Param        kategori_kegiatan_id formData integer false "ID kategori kegiatan"
// @Param        kegiatan_tanggal formData string false "Tanggal kegiatan (format: RFC3339)"
// @Param        kegiatan_lokasi formData string false "Lokasi kegiatan"
// @Param        kegiatan_pj formData string false "Penanggung jawab kegiatan"
// @Param        kegiatan_deskripsi formData string false "Deskripsi kegiatan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Kegiatan berhasil diupdate"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      404  {object}  map[string]interface{}  "Kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengupdate kegiatan"
// @Router       /api/kegiatan/{id} [put]
func (kc *KegiatanController) UpdateKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

	// Update fields menggunakan map
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

// DeleteKegiatan godoc
// @Summary      Menghapus kegiatan
// @Description  Menghapus kegiatan berdasarkan ID
// @Tags         kegiatan
// @Produce      json
// @Param        id   path      int  true  "ID Kegiatan"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Kegiatan berhasil dihapus"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Kegiatan tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal menghapus kegiatan"
// @Router       /api/kegiatan/{id} [delete]
func (kc *KegiatanController) DeleteKegiatan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
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

	// Delete
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

// GetKegiatanMendatang godoc
// @Summary      Mendapatkan kegiatan mendatang
// @Description  Mendapatkan daftar kegiatan yang akan datang (tanggal >= hari ini)
// @Tags         kegiatan
// @Produce      json
// @Param        limit query int false "Jumlah data (default: 5)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar kegiatan mendatang"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kegiatan"
// @Router       /api/kegiatan/mendatang [get]
func (kc *KegiatanController) GetKegiatanMendatang(c *gin.Context) {
	var kegiatan []models.Kegiatan

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// Query kegiatan dengan tanggal >= hari ini
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

// GetKegiatanByBulanTahun godoc
// @Summary      Mendapatkan kegiatan per bulan dan tahun
// @Description  Mendapatkan daftar kegiatan berdasarkan bulan dan tahun tertentu
// @Tags         kegiatan
// @Produce      json
// @Param        bulan   path      int  true  "Bulan (1-12)"
// @Param        tahun   path      int  true  "Tahun (2000-2100)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar kegiatan bulanan"
// @Failure      400  {object}  map[string]interface{}  "Parameter tidak valid"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data kegiatan"
// @Router       /api/kegiatan/bulan/{bulan}/tahun/{tahun} [get]
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

	// Query kegiatan by bulan dan tahun
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

// GetStatistikKegiatan godoc
// @Summary      Mendapatkan statistik kegiatan
// @Description  Mendapatkan statistik jumlah kegiatan per bulan (6 bulan terakhir)
// @Tags         kegiatan
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Statistik kegiatan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil statistik"
// @Router       /api/kegiatan/statistik [get]
func (kc *KegiatanController) GetStatistikKegiatan(c *gin.Context) {
	type StatistikBulanan struct {
		Bulan         string `json:"bulan" example:"January 2024"`
		Tahun         int    `json:"tahun" example:"2024"`
		BulanAngka    int    `json:"bulan_angka" example:"1"`
		TotalKegiatan int    `json:"total_kegiatan" example:"5"`
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

		// Hitung total kegiatan per bulan
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

// SearchKegiatan godoc
// @Summary      Mencari kegiatan
// @Description  Mencari kegiatan berdasarkan kata kunci
// @Tags         kegiatan
// @Produce      json
// @Param        q   query     string  true  "Kata kunci pencarian"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Hasil pencarian kegiatan"
// @Failure      400  {object}  map[string]interface{}  "Parameter pencarian harus diisi"
// @Failure      500  {object}  map[string]interface{}  "Gagal mencari kegiatan"
// @Router       /api/kegiatan/search [get]
func (kc *KegiatanController) SearchKegiatan(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var kegiatan []models.Kegiatan

	// Execute search query
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