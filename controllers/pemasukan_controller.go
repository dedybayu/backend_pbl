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

type PemasukanController struct {
	db *gorm.DB
}

func NewPemasukanController(db *gorm.DB) *PemasukanController {
	return &PemasukanController{db: db}
}

// Request structs
type CreatePemasukanRequest struct {
	KategoriPemasukanID uint      `json:"kategori_pemasukan_id" binding:"required"`
	PemasukanNama       string    `json:"pemasukan_nama" binding:"required"`
	PemasukanTanggal    time.Time `json:"pemasukan_tanggal" binding:"required"`
	PemasukanNominal    float64   `json:"pemasukan_nominal" binding:"required"`
	PemasukanBukti      string    `json:"pemasukan_bukti"`
}

type UpdatePemasukanRequest struct {
	KategoriPemasukanID uint      `json:"kategori_pemasukan_id"`
	PemasukanNama       string    `json:"pemasukan_nama"`
	PemasukanTanggal    time.Time `json:"pemasukan_tanggal"`
	PemasukanNominal    float64   `json:"pemasukan_nominal"`
	PemasukanBukti      string    `json:"pemasukan_bukti"`
}

// ✅ CREATE - Membuat pemasukan baru
func (pc *PemasukanController) CreatePemasukan(c *gin.Context) {
	var req CreatePemasukanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.PemasukanNama = strings.TrimSpace(req.PemasukanNama)
	req.PemasukanBukti = strings.TrimSpace(req.PemasukanBukti)

	// Validasi required fields
	if req.PemasukanNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama pemasukan harus diisi",
		})
		return
	}

	// Validasi nominal
	if req.PemasukanNominal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nominal pemasukan harus lebih dari 0",
		})
		return
	}

	// Validasi tanggal tidak boleh lebih besar dari hari ini
	if req.PemasukanTanggal.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tanggal pemasukan tidak boleh lebih besar dari hari ini",
		})
		return
	}

	// Check if kategori pemasukan exists
	var kategori models.KategoriPemasukan
	if err := pc.db.First(&kategori, req.KategoriPemasukanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi kategori pemasukan",
			})
		}
		return
	}

	// Buat pemasukan baru
	pemasukan := models.Pemasukan{
		KategoriPemasukanID: req.KategoriPemasukanID,
		PemasukanNama:       req.PemasukanNama,
		PemasukanTanggal:    req.PemasukanTanggal,
		PemasukanNominal:    req.PemasukanNominal,
		PemasukanBukti:      req.PemasukanBukti,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := pc.db.Create(&pemasukan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat pemasukan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori
	if err := pc.db.Preload("KategoriPemasukan").First(&pemasukan, pemasukan.PemasukanID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data pemasukan yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pemasukan berhasil dibuat",
		"data":    pemasukan,
	})
}

// ✅ READ - Mendapatkan semua pemasukan
func (pc *PemasukanController) GetAllPemasukan(c *gin.Context) {
	var pemasukan []models.Pemasukan

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
	query := pc.db.Model(&models.Pemasukan{}).Preload("KategoriPemasukan")

	// Apply filters
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("pemasukan_nama LIKE ?", "%"+searchSafe+"%")
	}

	if kategoriID != "" {
		kategoriIDSafe, err := strconv.ParseUint(kategoriID, 10, 32)
		if err == nil {
			query = query.Where("kategori_pemasukan_id = ?", kategoriIDSafe)
		}
	}

	if tanggalFrom != "" {
		if tanggalFromSafe, err := time.Parse("2006-01-02", tanggalFrom); err == nil {
			query = query.Where("DATE(pemasukan_tanggal) >= ?", tanggalFromSafe.Format("2006-01-02"))
		}
	}

	if tanggalTo != "" {
		if tanggalToSafe, err := time.Parse("2006-01-02", tanggalTo); err == nil {
			query = query.Where("DATE(pemasukan_tanggal) <= ?", tanggalToSafe.Format("2006-01-02"))
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("pemasukan_tanggal DESC, created_at DESC").
		Find(&pemasukan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data pemasukan",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pemasukan,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan pemasukan by ID
func (pc *PemasukanController) GetPemasukanByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pemasukanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pemasukan tidak valid",
		})
		return
	}

	var pemasukan models.Pemasukan
	if err := pc.db.Preload("KategoriPemasukan").First(&pemasukan, pemasukanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data pemasukan",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pemasukan,
	})
}

// ✅ UPDATE - Mengupdate pemasukan
func (pc *PemasukanController) UpdatePemasukan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pemasukanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pemasukan tidak valid",
		})
		return
	}

	var pemasukan models.Pemasukan
	if err := pc.db.First(&pemasukan, pemasukanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan pemasukan",
			})
		}
		return
	}

	var req UpdatePemasukanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.PemasukanNama != "" {
		req.PemasukanNama = strings.TrimSpace(req.PemasukanNama)
	}
	if req.PemasukanBukti != "" {
		req.PemasukanBukti = strings.TrimSpace(req.PemasukanBukti)
	}

	// Validasi nominal jika diupdate
	if req.PemasukanNominal != 0 && req.PemasukanNominal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nominal pemasukan harus lebih dari 0",
		})
		return
	}

	// Validasi tanggal jika diupdate
	if !req.PemasukanTanggal.IsZero() {
		if req.PemasukanTanggal.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tanggal pemasukan tidak boleh lebih besar dari hari ini",
			})
			return
		}
	}

	// Validasi kategori pemasukan jika diupdate
	if req.KategoriPemasukanID != 0 {
		var kategori models.KategoriPemasukan
		if err := pc.db.First(&kategori, req.KategoriPemasukanID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Kategori pemasukan tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi kategori pemasukan",
				})
			}
			return
		}
	}

	// Update fields menggunakan map (AMAN - GORM Updates dengan map)
	updates := make(map[string]interface{})
	
	if req.KategoriPemasukanID != 0 {
		updates["kategori_pemasukan_id"] = req.KategoriPemasukanID
	}
	if req.PemasukanNama != "" {
		updates["pemasukan_nama"] = req.PemasukanNama
	}
	if !req.PemasukanTanggal.IsZero() {
		updates["pemasukan_tanggal"] = req.PemasukanTanggal
	}
	if req.PemasukanNominal != 0 {
		updates["pemasukan_nominal"] = req.PemasukanNominal
	}
	if req.PemasukanBukti != "" {
		updates["pemasukan_bukti"] = req.PemasukanBukti
	}
	
	updates["updated_at"] = time.Now()

	if err := pc.db.Model(&pemasukan).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate pemasukan",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk kategori
	if err := pc.db.Preload("KategoriPemasukan").First(&pemasukan, pemasukanID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data pemasukan yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pemasukan berhasil diupdate",
		"data":    pemasukan,
	})
}

// ✅ DELETE - Menghapus pemasukan
func (pc *PemasukanController) DeletePemasukan(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pemasukanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pemasukan tidak valid",
		})
		return
	}

	var pemasukan models.Pemasukan
	if err := pc.db.First(&pemasukan, pemasukanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pemasukan tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan pemasukan",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := pc.db.Delete(&pemasukan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus pemasukan",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pemasukan berhasil dihapus",
	})
}

// ✅ GET - Statistik pemasukan
func (pc *PemasukanController) GetStatistikPemasukan(c *gin.Context) {
	type StatistikResult struct {
		TotalPemasukan   int64   `json:"total_pemasukan"`
		RataRataBulanan  float64 `json:"rata_rata_bulanan"`
		BulanIni         int64   `json:"bulan_ini"`
		MingguIni        int64   `json:"minggu_ini"`
		TotalNominal     float64 `json:"total_nominal"`
	}

	var statistik StatistikResult

	// Hitung total pemasukan (AMAN)
	pc.db.Model(&models.Pemasukan{}).Count(&statistik.TotalPemasukan)

	// Hitung pemasukan bulan ini (AMAN)
	awalBulan := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	pc.db.Model(&models.Pemasukan{}).
		Where("pemasukan_tanggal >= ?", awalBulan).
		Count(&statistik.BulanIni)

	// Hitung pemasukan minggu ini (AMAN)
	awalMinggu := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1)
	pc.db.Model(&models.Pemasukan{}).
		Where("pemasukan_tanggal >= ?", awalMinggu).
		Count(&statistik.MingguIni)

	// Hitung total nominal (AMAN)
	pc.db.Model(&models.Pemasukan{}).
		Select("COALESCE(SUM(pemasukan_nominal), 0)").
		Row().
		Scan(&statistik.TotalNominal)

	// Hitung rata-rata bulanan
	if statistik.TotalPemasukan > 0 {
		// Asumsi data selama 12 bulan
		statistik.RataRataBulanan = statistik.TotalNominal / 12
	}

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ GET - Total nominal pemasukan per kategori
func (pc *PemasukanController) GetTotalNominalPerKategori(c *gin.Context) {
	type TotalPerKategori struct {
		KategoriPemasukanNama string  `json:"kategori_pemasukan_nama"`
		TotalNominal         float64 `json:"total_nominal"`
		Persentase           float64 `json:"persentase"`
	}

	var results []TotalPerKategori

	// Query total nominal per kategori (AMAN)
	if err := pc.db.
		Model(&models.Pemasukan{}).
		Select("kategori_pemasukan.kategori_pemasukan_nama, SUM(pemasukan_nominal) as total_nominal").
		Joins("LEFT JOIN kategori_pemasukans ON kategori_pemasukans.kategori_pemasukan_id = pemasukans.kategori_pemasukan_id").
		Group("kategori_pemasukan.kategori_pemasukan_id, kategori_pemasukan.kategori_pemasukan_nama").
		Order("total_nominal DESC").
		Find(&results).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data total nominal per kategori",
		})
		return
	}

	// Hitung total keseluruhan untuk persentase
	var totalKeseluruhan float64
	for _, result := range results {
		totalKeseluruhan += result.TotalNominal
	}

	// Hitung persentase
	for i := range results {
		if totalKeseluruhan > 0 {
			results[i].Persentase = (results[i].TotalNominal / totalKeseluruhan) * 100
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"data": results,
		"total_keseluruhan": totalKeseluruhan,
	})
}

// ✅ GET - Laporan pemasukan bulanan
func (pc *PemasukanController) GetLaporanPemasukanBulanan(c *gin.Context) {
	type LaporanBulanan struct {
		Bulan          string  `json:"bulan"`
		Tahun          int     `json:"tahun"`
		BulanAngka     int     `json:"bulan_angka"`
		TotalPemasukan float64 `json:"total_pemasukan"`
		JumlahTransaksi int64  `json:"jumlah_transaksi"`
	}

	var laporan []LaporanBulanan

	// Hitung 6 bulan terakhir
	sekarang := time.Now()
	for i := 0; i < 6; i++ {
		tanggal := sekarang.AddDate(0, -i, 0)
		bulan := int(tanggal.Month())
		tahun := tanggal.Year()

		var total float64
		var jumlah int64

		awalBulan := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.UTC)
		akhirBulan := awalBulan.AddDate(0, 1, -1)

		// Hitung total pemasukan per bulan (AMAN)
		pc.db.Model(&models.Pemasukan{}).
			Where("pemasukan_tanggal BETWEEN ? AND ?", awalBulan, akhirBulan).
			Count(&jumlah)

		// Hitung jumlah transaksi per bulan (AMAN)
		pc.db.Model(&models.Pemasukan{}).
			Select("COALESCE(SUM(pemasukan_nominal), 0)").
			Where("pemasukan_tanggal BETWEEN ? AND ?", awalBulan, akhirBulan).
			Row().
			Scan(&total)

		laporan = append(laporan, LaporanBulanan{
			Bulan:           tanggal.Format("January 2006"),
			Tahun:           tahun,
			BulanAngka:      bulan,
			TotalPemasukan:  total,
			JumlahTransaksi: jumlah,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": laporan,
	})
}