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

type MutasiKeluargaController struct {
	db *gorm.DB
}

func NewMutasiKeluargaController(db *gorm.DB) *MutasiKeluargaController {
	return &MutasiKeluargaController{db: db}
}

// Request structs
type CreateMutasiKeluargaRequest struct {
	KeluargaID          uint      `json:"keluarga_id" binding:"required"`
	MutasiKeluargaJenis string    `json:"mutasi_keluarga_jenis" binding:"required"`
	MutasiKeluargaAlasan string   `json:"mutasi_keluarga_alasan"`
	MutasiKeluargaTanggal time.Time `json:"mutasi_keluarga_tanggal" binding:"required"`
}

type UpdateMutasiKeluargaRequest struct {
	KeluargaID          uint      `json:"keluarga_id"`
	MutasiKeluargaJenis string    `json:"mutasi_keluarga_jenis"`
	MutasiKeluargaAlasan string   `json:"mutasi_keluarga_alasan"`
	MutasiKeluargaTanggal time.Time `json:"mutasi_keluarga_tanggal"`
}

// ✅ CREATE - Membuat mutasi keluarga baru
func (mc *MutasiKeluargaController) CreateMutasiKeluarga(c *gin.Context) {
	var req CreateMutasiKeluargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.MutasiKeluargaJenis = strings.TrimSpace(req.MutasiKeluargaJenis)
	req.MutasiKeluargaAlasan = strings.TrimSpace(req.MutasiKeluargaAlasan)

	// Validasi required fields
	if req.MutasiKeluargaJenis == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Jenis mutasi harus diisi",
		})
		return
	}

	// Validasi jenis mutasi
	if !isValidJenisMutasi(req.MutasiKeluargaJenis) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Jenis mutasi harus 'masuk' atau 'keluar'",
		})
		return
	}

	// Validasi tanggal tidak boleh lebih besar dari hari ini
	if req.MutasiKeluargaTanggal.After(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tanggal mutasi tidak boleh lebih besar dari hari ini",
		})
		return
	}

	// Check if keluarga exists
	var keluarga models.Keluarga
	if err := mc.db.First(&keluarga, req.KeluargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Keluarga tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi keluarga",
			})
		}
		return
	}

	// Buat mutasi keluarga baru
	mutasi := models.MutasiKeluarga{
		KeluargaID:          req.KeluargaID,
		MutasiKeluargaJenis: req.MutasiKeluargaJenis,
		MutasiKeluargaAlasan: req.MutasiKeluargaAlasan,
		MutasiKeluargaTanggal: req.MutasiKeluargaTanggal,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := mc.db.Create(&mutasi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat mutasi keluarga",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data keluarga
	if err := mc.db.Preload("Keluarga").First(&mutasi, mutasi.MutasiKeluargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data mutasi keluarga yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Mutasi keluarga berhasil dibuat",
		"data":    mutasi,
	})
}

// ✅ READ - Mendapatkan semua mutasi keluarga
func (mc *MutasiKeluargaController) GetAllMutasiKeluarga(c *gin.Context) {
	var mutasi []models.MutasiKeluarga

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	keluargaID := c.Query("keluarga_id")
	jenis := c.Query("jenis")
	tanggalFrom := c.Query("tanggal_from")
	tanggalTo := c.Query("tanggal_to")

	// Build query dengan GORM (AMAN - parameterized queries)
	query := mc.db.Model(&models.MutasiKeluarga{}).Preload("Keluarga")

	// Apply filters
	if keluargaID != "" {
		keluargaIDSafe, err := strconv.ParseUint(keluargaID, 10, 32)
		if err == nil {
			query = query.Where("keluarga_id = ?", keluargaIDSafe)
		}
	}

	if jenis != "" && isValidJenisMutasi(jenis) {
		query = query.Where("mutasi_keluarga_jenis = ?", jenis)
	}

	if tanggalFrom != "" {
		if tanggalFromSafe, err := time.Parse("2006-01-02", tanggalFrom); err == nil {
			query = query.Where("DATE(mutasi_keluarga_tanggal) >= ?", tanggalFromSafe.Format("2006-01-02"))
		}
	}

	if tanggalTo != "" {
		if tanggalToSafe, err := time.Parse("2006-01-02", tanggalTo); err == nil {
			query = query.Where("DATE(mutasi_keluarga_tanggal) <= ?", tanggalToSafe.Format("2006-01-02"))
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("mutasi_keluarga_tanggal DESC, created_at DESC").
		Find(&mutasi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data mutasi keluarga",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": mutasi,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan mutasi keluarga by ID
func (mc *MutasiKeluargaController) GetMutasiKeluargaByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	mutasiID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID mutasi keluarga tidak valid",
		})
		return
	}

	var mutasi models.MutasiKeluarga
	if err := mc.db.Preload("Keluarga").First(&mutasi, mutasiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Mutasi keluarga tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data mutasi keluarga",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": mutasi,
	})
}

// ✅ UPDATE - Mengupdate mutasi keluarga
func (mc *MutasiKeluargaController) UpdateMutasiKeluarga(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	mutasiID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID mutasi keluarga tidak valid",
		})
		return
	}

	var mutasi models.MutasiKeluarga
	if err := mc.db.First(&mutasi, mutasiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Mutasi keluarga tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan mutasi keluarga",
			})
		}
		return
	}

	var req UpdateMutasiKeluargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.MutasiKeluargaJenis != "" {
		req.MutasiKeluargaJenis = strings.TrimSpace(req.MutasiKeluargaJenis)
	}
	if req.MutasiKeluargaAlasan != "" {
		req.MutasiKeluargaAlasan = strings.TrimSpace(req.MutasiKeluargaAlasan)
	}

	// Validasi jenis mutasi jika diupdate
	if req.MutasiKeluargaJenis != "" && !isValidJenisMutasi(req.MutasiKeluargaJenis) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Jenis mutasi harus 'masuk' atau 'keluar'",
		})
		return
	}

	// Validasi tanggal jika diupdate
	if !req.MutasiKeluargaTanggal.IsZero() {
		if req.MutasiKeluargaTanggal.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Tanggal mutasi tidak boleh lebih besar dari hari ini",
			})
			return
		}
	}

	// Validasi keluarga jika diupdate
	if req.KeluargaID != 0 {
		var keluarga models.Keluarga
		if err := mc.db.First(&keluarga, req.KeluargaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Keluarga tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi keluarga",
				})
			}
			return
		}
	}

	// Update fields menggunakan map (AMAN - GORM Updates dengan map)
	updates := make(map[string]interface{})
	
	if req.KeluargaID != 0 {
		updates["keluarga_id"] = req.KeluargaID
	}
	if req.MutasiKeluargaJenis != "" {
		updates["mutasi_keluarga_jenis"] = req.MutasiKeluargaJenis
	}
	if req.MutasiKeluargaAlasan != "" {
		updates["mutasi_keluarga_alasan"] = req.MutasiKeluargaAlasan
	}
	if !req.MutasiKeluargaTanggal.IsZero() {
		updates["mutasi_keluarga_tanggal"] = req.MutasiKeluargaTanggal
	}
	
	updates["updated_at"] = time.Now()

	if err := mc.db.Model(&mutasi).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate mutasi keluarga",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk keluarga
	if err := mc.db.Preload("Keluarga").First(&mutasi, mutasiID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data mutasi keluarga yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mutasi keluarga berhasil diupdate",
		"data":    mutasi,
	})
}

// ✅ DELETE - Menghapus mutasi keluarga
func (mc *MutasiKeluargaController) DeleteMutasiKeluarga(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	mutasiID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID mutasi keluarga tidak valid",
		})
		return
	}

	var mutasi models.MutasiKeluarga
	if err := mc.db.First(&mutasi, mutasiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Mutasi keluarga tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan mutasi keluarga",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := mc.db.Delete(&mutasi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus mutasi keluarga",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Mutasi keluarga berhasil dihapus",
	})
}

// ✅ GET - Mendapatkan mutasi keluarga by keluarga ID
func (mc *MutasiKeluargaController) GetMutasiByKeluargaID(c *gin.Context) {
	keluargaID := c.Param("keluarga_id")

	// Validasi ID (AMAN - dikonversi ke uint)
	keluargaIDSafe, err := strconv.ParseUint(keluargaID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID keluarga tidak valid",
		})
		return
	}

	// Check if keluarga exists
	var keluarga models.Keluarga
	if err := mc.db.First(&keluarga, keluargaIDSafe).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Keluarga tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi keluarga",
			})
		}
		return
	}

	var mutasi []models.MutasiKeluarga

	// Query mutasi by keluarga ID (AMAN - parameterized)
	if err := mc.db.
		Preload("Keluarga").
		Where("keluarga_id = ?", keluargaIDSafe).
		Order("mutasi_keluarga_tanggal DESC").
		Find(&mutasi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data mutasi keluarga",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  mutasi,
		"total": len(mutasi),
		"keluarga": gin.H{
			"keluarga_id":   keluarga.KeluargaID,
			"keluarga_nama": keluarga.KeluargaNama,
		},
	})
}

// ✅ GET - Statistik mutasi keluarga
func (mc *MutasiKeluargaController) GetStatistikMutasiKeluarga(c *gin.Context) {
	type StatistikResult struct {
		TotalMutasi int64 `json:"total_mutasi"`
		Masuk       int64 `json:"masuk"`
		Keluar      int64 `json:"keluar"`
		BulanIni    int64 `json:"bulan_ini"`
	}

	var statistik StatistikResult

	// Hitung total mutasi (AMAN)
	mc.db.Model(&models.MutasiKeluarga{}).Count(&statistik.TotalMutasi)

	// Hitung mutasi masuk (AMAN)
	mc.db.Model(&models.MutasiKeluarga{}).
		Where("mutasi_keluarga_jenis = ?", "masuk").
		Count(&statistik.Masuk)

	// Hitung mutasi keluar (AMAN)
	mc.db.Model(&models.MutasiKeluarga{}).
		Where("mutasi_keluarga_jenis = ?", "keluar").
		Count(&statistik.Keluar)

	// Hitung mutasi bulan ini (AMAN)
	awalBulan := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	mc.db.Model(&models.MutasiKeluarga{}).
		Where("mutasi_keluarga_tanggal >= ?", awalBulan).
		Count(&statistik.BulanIni)

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ GET - Laporan mutasi keluarga per bulan
func (mc *MutasiKeluargaController) GetLaporanMutasiBulanan(c *gin.Context) {
	type LaporanBulanan struct {
		Bulan        string `json:"bulan"`
		Tahun        int    `json:"tahun"`
		BulanAngka   int    `json:"bulan_angka"`
		TotalMasuk   int64  `json:"total_masuk"`
		TotalKeluar  int64  `json:"total_keluar"`
		TotalMutasi  int64  `json:"total_mutasi"`
	}

	var laporan []LaporanBulanan

	// Hitung 6 bulan terakhir
	sekarang := time.Now()
	for i := 0; i < 6; i++ {
		tanggal := sekarang.AddDate(0, -i, 0)
		bulan := int(tanggal.Month())
		tahun := tanggal.Year()

		var masuk int64
		var keluar int64
		var total int64

		awalBulan := time.Date(tahun, time.Month(bulan), 1, 0, 0, 0, 0, time.UTC)
		akhirBulan := awalBulan.AddDate(0, 1, -1)

		// Hitung mutasi masuk per bulan (AMAN)
		mc.db.Model(&models.MutasiKeluarga{}).
			Where("mutasi_keluarga_jenis = ? AND mutasi_keluarga_tanggal BETWEEN ? AND ?", "masuk", awalBulan, akhirBulan).
			Count(&masuk)

		// Hitung mutasi keluar per bulan (AMAN)
		mc.db.Model(&models.MutasiKeluarga{}).
			Where("mutasi_keluarga_jenis = ? AND mutasi_keluarga_tanggal BETWEEN ? AND ?", "keluar", awalBulan, akhirBulan).
			Count(&keluar)

		total = masuk + keluar

		laporan = append(laporan, LaporanBulanan{
			Bulan:        tanggal.Format("January 2006"),
			Tahun:        tahun,
			BulanAngka:   bulan,
			TotalMasuk:   masuk,
			TotalKeluar:  keluar,
			TotalMutasi:  total,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"data": laporan,
	})
}

// ✅ Helper function untuk validasi jenis mutasi
func isValidJenisMutasi(jenis string) bool {
	return jenis == "masuk" || jenis == "keluar"
}