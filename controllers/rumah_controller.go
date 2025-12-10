package controllers

import (
	"net/http"
	"strconv"
	// "time"

	"rt-management/models"
	// "rt-management/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RumahController struct {
	db *gorm.DB
}

func NewRumahController(db *gorm.DB) *RumahController {
	return &RumahController{db: db}
}

// Request structs
type CreateRumahRequest struct {
	RumahAlamat string `form:"rumah_alamat" binding:"required"`
	RumahStatus string `form:"rumah_status"`
}

type UpdateRumahRequest struct {
	RumahAlamat string `form:"rumah_alamat"`
	RumahStatus string `form:"rumah_status"`
}

// CreateRumah membuat data rumah baru
// @Summary      Membuat rumah baru
// @Description  Membuat data rumah baru dengan alamat dan status
// @Tags         rumah
// @Accept       multipart/form-data
// @Produce      json
// @Param        rumah_alamat formData string true "Alamat rumah"
// @Param        rumah_status formData string false "Status rumah" Enums(tersedia, ditempati)
// @Success      201 {object} map[string]interface{} "Rumah berhasil dibuat"
// @Failure      400 {object} map[string]interface{} "Data request tidak valid"
// @Failure      500 {object} map[string]interface{} "Gagal membuat rumah"
// @Router       /rumah [post]
func (rc *RumahController) CreateRumah(c *gin.Context) {
	var req CreateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Validasi alamat
	if req.RumahAlamat == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Alamat rumah harus diisi",
		})
		return
	}

	// Set default status
	if req.RumahStatus == "" {
		req.RumahStatus = "tersedia"
	}

	// Validasi status
	if !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	rumah := models.Rumah{
		RumahAlamat: req.RumahAlamat,
		RumahStatus: req.RumahStatus,
	}

	if err := rc.db.Create(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Rumah berhasil dibuat",
		"data":    rumah,
	})
}

// GetAllRumah mengambil semua data rumah dengan filter opsional
// @Summary      Mengambil semua data rumah
// @Description  Mengambil daftar rumah dengan filter opsional berdasarkan status atau ID warga
// @Tags         rumah
// @Accept       json
// @Produce      json
// @Param        page    query int    false "Nomor halaman (dinonaktifkan secara default)"
// @Param        limit   query int    false "Jumlah item per halaman (dinonaktifkan secara default, gunakan 0 untuk tanpa batas)"
// @Param        status  query string false "Filter berdasarkan status" Enums(tersedia, ditempati)
// @Param        warga_id query string false "Filter berdasarkan ID warga"
// @Param        all     query bool   false "Ambil semua data tanpa pagination"
// @Success      200 {object} map[string]interface{} "Daftar rumah"
// @Failure      500 {object} map[string]interface{} "Gagal mengambil data rumah"
// @Router       /rumah [get]
func (rc *RumahController) GetAllRumah(c *gin.Context) {
	var rumah []models.Rumah

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0"))
	allRecords, _ := strconv.ParseBool(c.Query("all"))
	
	status := c.Query("status")
	wargaID := c.Query("warga_id")

	query := rc.db.Model(&models.Rumah{})

	if status != "" {
		query = query.Where("rumah_status = ?", status)
	}

	// Filter rumah berdasarkan warga penghuni
	if wargaID != "" {
		query = query.Joins("JOIN wargas ON wargas.rumah_id = rumahs.rumah_id").
			Where("wargas.warga_id = ?", wargaID)
	}

	// Hitung total untuk informasi saja
	var total int64
	query.Count(&total)

	// Tentukan apakah akan menggunakan pagination atau tidak
	usePagination := !allRecords && limit > 0

	if usePagination {
		offset := (page - 1) * limit
		if offset < 0 {
			offset = 0
		}
		query = query.Offset(offset).Limit(limit)
	}

	if err := query.Find(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data rumah",
		})
		return
	}

	// Tambahkan data warga penghuni
	type RumahWithWarga struct {
		models.Rumah
		Warga models.Warga `json:"warga"`
	}

	var results []RumahWithWarga

	for _, r := range rumah {
		var warga models.Warga
		rc.db.Where("rumah_id = ?", r.RumahID).First(&warga)

		results = append(results, RumahWithWarga{
			Rumah: r,
			Warga: warga,
		})
	}

	response := gin.H{
		"data": results,
		"total": total,
	}

	// Hanya tambahkan informasi pagination jika digunakan
	if usePagination {
		response["pagination"] = gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"has_next": page*limit < int(total),
			"has_prev": page > 1,
		}
	} else {
		response["message"] = "Semua data berhasil diambil (pagination dinonaktifkan)"
	}

	c.JSON(http.StatusOK, response)
}

// GetRumahByID mengambil data rumah berdasarkan ID
// @Summary      Mengambil data rumah berdasarkan ID
// @Description  Mengambil detail rumah termasuk data warga berdasarkan ID rumah
// @Tags         rumah
// @Accept       json
// @Produce      json
// @Param        id path string true "ID Rumah"
// @Success      200 {object} map[string]interface{} "Detail rumah"
// @Failure      404 {object} map[string]interface{} "Rumah tidak ditemukan"
// @Router       /api/rumah/{id} [get]
func (rc *RumahController) GetRumahByID(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rumah tidak ditemukan",
		})
		return
	}

	var warga models.Warga
	rc.db.Where("rumah_id = ?", rumah.RumahID).First(&warga)

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"rumah": rumah,
			"warga": warga,
		},
	})
}

// UpdateRumah memperbarui data rumah berdasarkan ID
// @Summary      Memperbarui data rumah
// @Description  Memperbarui informasi rumah berdasarkan ID
// @Tags         rumah
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "ID Rumah"
// @Param        rumah_alamat formData string false "Alamat rumah"
// @Param        rumah_status formData string false "Status rumah" Enums(tersedia, ditempati)
// @Success      200 {object} map[string]interface{} "Rumah berhasil diperbarui"
// @Failure      400 {object} map[string]interface{} "Data request tidak valid"
// @Failure      404 {object} map[string]interface{} "Rumah tidak ditemukan"
// @Failure      500 {object} map[string]interface{} "Gagal memperbarui rumah"
// @Router       /api/rumah/{id} [put]
func (rc *RumahController) UpdateRumah(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Rumah tidak ditemukan",
		})
		return
	}

	var req UpdateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	if req.RumahStatus != "" && !isValidRumahStatus(req.RumahStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status harus 'tersedia' atau 'ditempati'",
		})
		return
	}

	updates := map[string]interface{}{}
	if req.RumahAlamat != "" {
		updates["rumah_alamat"] = req.RumahAlamat
	}
	if req.RumahStatus != "" {
		updates["rumah_status"] = req.RumahStatus
	}

	if err := rc.db.Model(&rumah).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal memperbarui rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rumah berhasil diperbarui",
		"data":    rumah,
	})
}

// DeleteRumah menghapus data rumah berdasarkan ID
// @Summary      Menghapus data rumah
// @Description  Menghapus data rumah berdasarkan ID
// @Tags         rumah
// @Accept       json
// @Produce      json
// @Param        id path string true "ID Rumah"
// @Success      200 {object} map[string]interface{} "Rumah berhasil dihapus"
// @Failure      404 {object} map[string]interface{} "Rumah tidak ditemukan"
// @Failure      500 {object} map[string]interface{} "Gagal menghapus rumah"
// @Router       /api/rumah/{id} [delete]
func (rc *RumahController) DeleteRumah(c *gin.Context) {
	id := c.Param("id")

	var rumah models.Rumah
	if err := rc.db.First(&rumah, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Rumah tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan rumah",
			})
		}
		return
	}

	if err := rc.db.Delete(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rumah berhasil dihapus",
	})
}

// GetRumahByStatus mengambil data rumah berdasarkan status
// @Summary      Mengambil data rumah berdasarkan status
// @Description  Mengambil daftar rumah yang difilter berdasarkan status
// @Tags         rumah
// @Accept       json
// @Produce      json
// @Param        status path string true "Status rumah" Enums(tersedia, ditempati)
// @Success      200 {object} map[string]interface{} "Daftar rumah berdasarkan status"
// @Failure      400 {object} map[string]interface{} "Status tidak valid"
// @Failure      500 {object} map[string]interface{} "Gagal mengambil data rumah"
// @Router       /api/rumah/status/{status} [get]
func (rc *RumahController) GetRumahByStatus(c *gin.Context) {
	status := c.Param("status")

	if !isValidRumahStatus(status) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status tidak valid. Gunakan 'tersedia' atau 'ditempati'",
		})
		return
	}

	var rumah []models.Rumah
	if err := rc.db.Preload("Warga").Where("rumah_status = ?", status).Find(&rumah).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data rumah",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  rumah,
		"total": len(rumah),
	})
}

// ✅ Helper function untuk validasi status rumah
func isValidRumahStatus(status string) bool {
	return status == "tersedia" || status == "ditempati"
}