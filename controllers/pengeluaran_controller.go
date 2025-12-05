package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rt-management/helper"
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type PengeluaranController struct {
	db *gorm.DB
}

func NewPengeluaranController(db *gorm.DB) *PengeluaranController {
	return &PengeluaranController{db: db}
}

// Request structs
type CreatePengeluaranRequest struct {
	KategoriPengeluaranID uint      `form:"kategori_pengeluaran_id" binding:"required"`
	PengeluaranNama       string    `form:"pengeluaran_nama" binding:"required"`
	PengeluaranTanggal    string `form:"pengeluaran_tanggal" binding:"required"`
	PengeluaranNominal    float64   `form:"pengeluaran_nominal" binding:"required"`
	PengeluaranBukti      string    `form:"pengeluaran_bukti"`
}

type UpdatePengeluaranRequest struct {
	KategoriPengeluaranID uint      `form:"kategori_pengeluaran_id"`
	PengeluaranNama       string    `form:"pengeluaran_nama"`
	PengeluaranTanggal    string `form:"pengeluaran_tanggal"`
	PengeluaranNominal    float64   `form:"pengeluaran_nominal"`
	PengeluaranBukti      string    `form:"pengeluaran_bukti"`
}

// ✅ CREATE - Membuat pengeluaran baru
func (pc *PengeluaranController) CreatePengeluaran(c *gin.Context) {
	// Binding manual untuk form data
	kategoriPengeluaranIDStr := c.PostForm("kategori_pengeluaran_id")
	pengeluaranNama := strings.TrimSpace(c.PostForm("pengeluaran_nama"))
	pengeluaranTanggalStr := strings.TrimSpace(c.PostForm("pengeluaran_tanggal"))
	pengeluaranNominalStr := c.PostForm("pengeluaran_nominal")

	// Validasi required fields
	if kategoriPengeluaranIDStr == "" || pengeluaranNama == "" || pengeluaranTanggalStr == "" || pengeluaranNominalStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Semua field wajib harus diisi",
		})
		return
	}

	// Convert kategori_pengeluaran_id to uint
	kategoriPengeluaranID, err := strconv.ParseUint(kategoriPengeluaranIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Kategori pengeluaran ID tidak valid",
		})
		return
	}

	// Convert nominal to float64
	pengeluaranNominal, err := strconv.ParseFloat(pengeluaranNominalStr, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nominal pengeluaran tidak valid",
		})
		return
	}

	// Validasi nominal
	if pengeluaranNominal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nominal pengeluaran harus lebih dari 0",
		})
		return
	}

	// Parsing tanggal dari string ke time.Time
	pengeluaranTanggal, err := time.Parse("2006-01-02", pengeluaranTanggalStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Format tanggal tidak valid. Gunakan format YYYY-MM-DD",
		})
		return
	}

	// Validasi tanggal tidak boleh lebih besar dari hari ini
	if pengeluaranTanggal.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tanggal pengeluaran tidak boleh lebih besar dari hari ini",
		})
		return
	}

	// Check if kategori pengeluaran exists
	var kategori models.KategoriPengeluaran
	if err := pc.db.First(&kategori, uint(kategoriPengeluaranID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi kategori pengeluaran",
			})
		}
		return
	}

	// Handle file upload untuk pengeluaran_bukti
	pengeluaranBuktiFilename := ""
	if _, header, err := c.Request.FormFile("pengeluaran_bukti"); err == nil && header != nil {
		// Gunakan helper untuk handle upload
		filename, err := helper.HandleFileImageUpload(c, "pengeluaran_bukti", "")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload bukti pengeluaran",
				"details": err.Error(),
			})
			return
		}
		pengeluaranBuktiFilename = filename
	}

	// Buat pengeluaran baru
	pengeluaran := models.Pengeluaran{
		KategoriPengeluaranID: uint(kategoriPengeluaranID),
		PengeluaranNama:       pengeluaranNama,
		PengeluaranTanggal:    pengeluaranTanggal,
		PengeluaranNominal:    pengeluaranNominal,
		PengeluaranBukti:      pengeluaranBuktiFilename,
		CreatedAt:             time.Now(),
		UpdatedAt:             time.Now(),
	}

	if err := pc.db.Create(&pengeluaran).Error; err != nil {
		// Jika gagal create, hapus file yang sudah diupload
		if pengeluaranBuktiFilename != "" {
			helper.DeleteOldPhoto(pengeluaranBuktiFilename, "pengeluaran_bukti")
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat pengeluaran",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori
	if err := pc.db.Preload("KategoriPengeluaran").First(&pengeluaran, pengeluaran.PengeluaranID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data pengeluaran yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Pengeluaran berhasil dibuat",
		"data":    pengeluaran,
	})
}

// ✅ READ - Mendapatkan semua pengeluaran
func (pc *PengeluaranController) GetAllPengeluaran(c *gin.Context) {
	var pengeluaran []models.Pengeluaran

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
	query := pc.db.Model(&models.Pengeluaran{}).Preload("KategoriPengeluaran")

	// Apply filters
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("pengeluaran_nama LIKE ?", "%"+searchSafe+"%")
	}

	if kategoriID != "" {
		kategoriIDSafe, err := strconv.ParseUint(kategoriID, 10, 32)
		if err == nil {
			query = query.Where("kategori_pengeluaran_id = ?", kategoriIDSafe)
		}
	}

	if tanggalFrom != "" {
		if tanggalFromSafe, err := time.Parse("2006-01-02", tanggalFrom); err == nil {
			query = query.Where("DATE(pengeluaran_tanggal) >= ?", tanggalFromSafe.Format("2006-01-02"))
		}
	}

	if tanggalTo != "" {
		if tanggalToSafe, err := time.Parse("2006-01-02", tanggalTo); err == nil {
			query = query.Where("DATE(pengeluaran_tanggal) <= ?", tanggalToSafe.Format("2006-01-02"))
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		// Limit(limit).
		Order("pengeluaran_tanggal DESC, created_at DESC").
		Find(&pengeluaran).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data pengeluaran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pengeluaran,
		// "pagination": gin.H{
		// 	"page":  page,
		// 	"limit": limit,
		// 	"total": total,
		// },
	})
}

// ✅ READ - Mendapatkan pengeluaran by ID
func (pc *PengeluaranController) GetPengeluaranByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pengeluaranID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pengeluaran tidak valid",
		})
		return
	}

	var pengeluaran models.Pengeluaran
	if err := pc.db.Preload("KategoriPengeluaran").First(&pengeluaran, pengeluaranID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data pengeluaran",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": pengeluaran,
	})
}

// ✅ UPDATE - Mengupdate pengeluaran
func (pc *PengeluaranController) UpdatePengeluaran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pengeluaranID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pengeluaran tidak valid",
		})
		return
	}

	var pengeluaran models.Pengeluaran
	if err := pc.db.First(&pengeluaran, pengeluaranID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan pengeluaran",
			})
		}
		return
	}

	// Binding manual untuk form data
	kategoriPengeluaranIDStr := c.PostForm("kategori_pengeluaran_id")
	pengeluaranNama := strings.TrimSpace(c.PostForm("pengeluaran_nama"))
	pengeluaranTanggalStr := strings.TrimSpace(c.PostForm("pengeluaran_tanggal"))
	pengeluaranNominalStr := c.PostForm("pengeluaran_nominal")

	// Update fields menggunakan map
	updates := make(map[string]interface{})

	// Handle kategori_pengeluaran_id jika diupdate
	if kategoriPengeluaranIDStr != "" {
		kategoriPengeluaranID, err := strconv.ParseUint(kategoriPengeluaranIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori pengeluaran ID tidak valid",
			})
			return
		}

		// Validasi kategori pengeluaran jika diupdate
		var kategori models.KategoriPengeluaran
		if err := pc.db.First(&kategori, uint(kategoriPengeluaranID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Kategori pengeluaran tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi kategori pengeluaran",
				})
			}
			return
		}
		updates["kategori_pengeluaran_id"] = uint(kategoriPengeluaranID)
	}

	// Handle pengeluaran_nama jika diupdate
	if pengeluaranNama != "" {
		updates["pengeluaran_nama"] = pengeluaranNama
	}

	// Handle pengeluaran_tanggal jika diupdate
	if pengeluaranTanggalStr != "" {
		// Parsing tanggal dari string ke time.Time
		pengeluaranTanggal, err := time.Parse("2006-01-02", pengeluaranTanggalStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Format tanggal tidak valid. Gunakan format YYYY-MM-DD",
			})
			return
		}

		// Validasi tanggal jika diupdate
		if pengeluaranTanggal.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tanggal pengeluaran tidak boleh lebih besar dari hari ini",
			})
			return
		}

		updates["pengeluaran_tanggal"] = pengeluaranTanggal
	}

	// Handle pengeluaran_nominal jika diupdate
	if pengeluaranNominalStr != "" {
		pengeluaranNominal, err := strconv.ParseFloat(pengeluaranNominalStr, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nominal pengeluaran tidak valid",
			})
			return
		}

		// Validasi nominal jika diupdate
		if pengeluaranNominal <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nominal pengeluaran harus lebih dari 0",
			})
			return
		}
		updates["pengeluaran_nominal"] = pengeluaranNominal
	}

	// Handle file upload untuk pengeluaran_bukti
	pengeluaranBuktiFilename := pengeluaran.PengeluaranBukti // Simpan filename lama dulu
	if _, header, err := c.Request.FormFile("pengeluaran_bukti"); err == nil && header != nil {
		// Ada file baru yang diupload
		filename, err := helper.HandleFileImageUpload(c, "pengeluaran_bukti", pengeluaran.PengeluaranBukti)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload bukti pengeluaran",
				"details": err.Error(),
			})
			return
		}
		pengeluaranBuktiFilename = filename
	}
	
	// Selalu update pengeluaran_bukti (bisa filename baru atau tetap yang lama)
	updates["pengeluaran_bukti"] = pengeluaranBuktiFilename
	updates["updated_at"] = time.Now()

	// Eksekusi update hanya jika ada field yang diupdate
	if len(updates) > 0 {
		if err := pc.db.Model(&pengeluaran).Updates(updates).Error; err != nil {
			// Jika gagal update, hapus file baru yang sudah diupload
			if pengeluaranBuktiFilename != pengeluaran.PengeluaranBukti {
				helper.DeleteOldPhoto(pengeluaranBuktiFilename, "pengeluaran_bukti")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal mengupdate pengeluaran",
				"details": err.Error(),
			})
			return
		}
	}

	// Reload dengan data terbaru termasuk kategori
	if err := pc.db.Preload("KategoriPengeluaran").First(&pengeluaran, pengeluaranID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data pengeluaran yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pengeluaran berhasil diupdate",
		"data":    pengeluaran,
	})
}

// ✅ DELETE - Menghapus pengeluaran
func (pc *PengeluaranController) DeletePengeluaran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	pengeluaranID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID pengeluaran tidak valid",
		})
		return
	}

	var pengeluaran models.Pengeluaran
	if err := pc.db.First(&pengeluaran, pengeluaranID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan pengeluaran",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := pc.db.Delete(&pengeluaran).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus pengeluaran",
			"details": err.Error(),
		})
		return
	}

	// Hapus file foto jika ada
	if pengeluaran.PengeluaranBukti != "" {
		helper.DeleteOldPhoto(pengeluaran.PengeluaranBukti, "pengeluaran_bukti")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Pengeluaran berhasil dihapus",
	})
}

// ✅ GET - Statistik pengeluaran
func (pc *PengeluaranController) GetStatistikPengeluaran(c *gin.Context) {
	type StatistikResult struct {
		TotalPengeluaran int64   `json:"total_pengeluaran"`
		RataRataBulanan  float64 `json:"rata_rata_bulanan"`
		BulanIni         int64   `json:"bulan_ini"`
		MingguIni        int64   `json:"minggu_ini"`
	}

	var statistik StatistikResult

	// Hitung total pengeluaran (AMAN)
	pc.db.Model(&models.Pengeluaran{}).Count(&statistik.TotalPengeluaran)

	// Hitung pengeluaran bulan ini (AMAN)
	awalBulan := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	pc.db.Model(&models.Pengeluaran{}).
		Where("pengeluaran_tanggal >= ?", awalBulan).
		Count(&statistik.BulanIni)

	// Hitung pengeluaran minggu ini (AMAN)
	awalMinggu := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1)
	pc.db.Model(&models.Pengeluaran{}).
		Where("pengeluaran_tanggal >= ?", awalMinggu).
		Count(&statistik.MingguIni)

	// Hitung rata-rata bulanan (AMAN)
	var totalNominal float64
	pc.db.Model(&models.Pengeluaran{}).
		Select("COALESCE(SUM(pengeluaran_nominal), 0)").
		Row().
		Scan(&totalNominal)

	if statistik.TotalPengeluaran > 0 {
		// Asumsi data selama 12 bulan
		statistik.RataRataBulanan = totalNominal / 12
	}

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ GET - Total nominal pengeluaran per kategori
func (pc *PengeluaranController) GetTotalNominalPerKategori(c *gin.Context) {
	type TotalPerKategori struct {
		KategoriPengeluaranNama string  `json:"kategori_pengeluaran_nama"`
		TotalNominal           float64 `json:"total_nominal"`
		Persentase             float64 `json:"persentase"`
	}

	var results []TotalPerKategori

	// Query total nominal per kategori (AMAN)
	if err := pc.db.
		Model(&models.Pengeluaran{}).
		Select("kategori_pengeluaran.kategori_pengeluaran_nama, SUM(pengeluaran_nominal) as total_nominal").
		Joins("LEFT JOIN kategori_pengeluarans ON kategori_pengeluarans.kategori_pengeluaran_id = pengeluarans.kategori_pengeluaran_id").
		Group("kategori_pengeluaran.kategori_pengeluaran_id, kategori_pengeluaran.kategori_pengeluaran_nama").
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

// ✅ GET - Laporan pengeluaran bulanan
func (pc *PengeluaranController) GetLaporanPengeluaranBulanan(c *gin.Context) {
	type LaporanBulanan struct {
		Bulan           string  `json:"bulan"`
		Tahun           int     `json:"tahun"`
		BulanAngka      int     `json:"bulan_angka"`
		TotalPengeluaran float64 `json:"total_pengeluaran"`
		JumlahTransaksi int64   `json:"jumlah_transaksi"`
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

		// Hitung total pengeluaran per bulan (AMAN)
		pc.db.Model(&models.Pengeluaran{}).
			Where("pengeluaran_tanggal BETWEEN ? AND ?", awalBulan, akhirBulan).
			Count(&jumlah)

		// Hitung jumlah transaksi per bulan (AMAN)
		pc.db.Model(&models.Pengeluaran{}).
			Select("COALESCE(SUM(pengeluaran_nominal), 0)").
			Where("pengeluaran_tanggal BETWEEN ? AND ?", awalBulan, akhirBulan).
			Row().
			Scan(&total)

		laporan = append(laporan, LaporanBulanan{
			Bulan:           tanggal.Format("January 2006"),
			Tahun:           tahun,
			BulanAngka:      bulan,
			TotalPengeluaran: total,
			JumlahTransaksi: jumlah,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": laporan,
	})
}


// ✅ GET - Mendapatkan gambar bukti pengeluaran
func (pc *PengeluaranController) GetPengeluaranBuktiImage(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	// Gunakan helper function GetFileByFileName
	file, err := helper.GetFileByFileName("pengeluaran_bukti", filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File bukti pengeluaran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal membuka file",
				"details": err.Error(),
			})
		}
		return
	}
	defer file.Close()

	// Dapatkan file info untuk Content-Type
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendapatkan info file",
		})
		return
	}

	// Set header yang sesuai
	ext := filepath.Ext(filename)
	contentType := helper.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Serve file
	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}