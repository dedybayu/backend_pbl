package controllers

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"rt-management/helper"
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TagihanIuranController struct {
	db *gorm.DB
}

func NewTagihanIuranController(db *gorm.DB) *TagihanIuranController {
	return &TagihanIuranController{db: db}
}

// Struct untuk request
type CreateTagihanIuranRequest struct {
	TagihanIuran        string  `json:"tagihan_iuran" binding:"required"`
	TagihanIuranNominal float64 `json:"tagihan_iuran_nominal" binding:"required"`
	WargaID             uint    `json:"warga_id" binding:"required"`
}

type UpdateTagihanIuranRequest struct {
	TagihanIuran        string  `json:"tagihan_iuran"`
	TagihanIuranNominal float64 `json:"tagihan_iuran_nominal"`
	WargaID             uint    `json:"warga_id"`
}

type UpdateStatusVerifikasiRequest struct {
	TagihanIuranStatusVerifikasi string `json:"tagihan_iuran_status_verifikasi" binding:"required,oneof=diterima ditolak menunggu"`
	Keterangan                   string `json:"keterangan"`
}

// TagihanIuranResponse represents the response structure for tagihan iuran
// @Description Response untuk data tagihan iuran
type TagihanIuranResponse struct {
	Success bool                `json:"success" example:"true"`
	Message string              `json:"message,omitempty" example:"Tagihan iuran berhasil dibuat"`
	Data    models.TagihanIuran `json:"data,omitempty"`
}

// TagihanIuranListResponse represents the response structure for list of tagihan iuran
// @Description Response untuk list data tagihan iuran
type TagihanIuranListResponse struct {
	Success bool                  `json:"success" example:"true"`
	Data    []models.TagihanIuran `json:"data"`
}

// ErrorResponse represents the error response structure
// @Description Response untuk error
type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Error   string `json:"error" example:"ID tagihan iuran tidak valid"`
	Details string `json:"details,omitempty" example:"strconv.ParseUint: parsing \"abc\": invalid syntax"`
}


// GetAll godoc
// @Summary Get all tagihan iuran
// @Description Mendapatkan semua data tagihan iuran dengan filter opsional
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param status query string false "Filter by status (tertagih/lunas)"
// @Param status_verifikasi query string false "Filter by status verifikasi (menunggu/diterima/ditolak)"
// @Param warga_id query string false "Filter by warga ID"
// @Security BearerAuth
// @Success 200 {object} TagihanIuranListResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran [get]
func (tic *TagihanIuranController) GetAll(c *gin.Context) {
	var tagihan []models.TagihanIuran
	
	// Get query parameters untuk filter
	status := c.Query("status")
	statusVerifikasi := c.Query("status_verifikasi")
	wargaID := c.Query("warga_id")
	
	// Build query dengan preload warga
	query := tic.db.Preload("Warga").Order("created_at DESC")
	
	// Apply filters jika ada
	if status != "" {
		query = query.Where("tagihan_iuran_status = ?", status)
	}
	
	if statusVerifikasi != "" {
		query = query.Where("tagihan_iuran_status_verifikasi = ?", statusVerifikasi)
	}
	
	if wargaID != "" {
		wID, err := strconv.ParseUint(wargaID, 10, 32)
		if err == nil {
			query = query.Where("warga_id = ?", wID)
		}
	}
	
	if err := query.Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengambil data tagihan iuran",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagihan,
	})
}

// GetByID godoc
// @Summary Get tagihan iuran by ID
// @Description Mendapatkan data tagihan iuran berdasarkan ID
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param id path string true "Tagihan Iuran ID"
// @Security BearerAuth
// @Success 200 {object} TagihanIuranResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/{id} [get]
func (tic *TagihanIuranController) GetByID(c *gin.Context) {
	id := c.Param("id")
	
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID tagihan iuran tidak valid",
		})
		return
	}
	
	var tagihan models.TagihanIuran
	if err := tic.db.Preload("Warga").First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Gagal mengambil data tagihan iuran",
			})
		}
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagihan,
	})
}

// CreateTagihan godoc
// @Summary Create new tagihan iuran
// @Description Membuat tagihan iuran baru (Hanya admin/ketua RT)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param request body CreateTagihanIuranRequest true "Data tagihan iuran"
// @Security BearerAuth
// @Success 201 {object} TagihanIuranResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran [post]
func (tic *TagihanIuranController) CreateTagihan(c *gin.Context) {
	var req CreateTagihanIuranRequest
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}
	
	// Validasi warga exists
	var warga models.Warga
	if err := tic.db.First(&warga, req.WargaID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Warga tidak ditemukan",
		})
		return
	}
	
	// Create tagihan iuran
	tagihan := models.TagihanIuran{
		TagihanIuran:                 req.TagihanIuran,
		TagihanIuranNominal:          req.TagihanIuranNominal,
		TagihanIuranStatus:           "tertagih",
		TagihanIuranStatusVerifikasi: "menunggu",
		WargaID:                      req.WargaID,
		CreatedAt:                    time.Now(),
		UpdatedAt:                    time.Now(),
	}
	
	if err := tic.db.Create(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal membuat tagihan iuran",
			"details": err.Error(),
		})
		return
	}
	
	// Reload with warga data
	tic.db.Preload("Warga").First(&tagihan, tagihan.ID)
	
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tagihan iuran berhasil dibuat",
		"data":    tagihan,
	})
}

// UpdateTagihan godoc
// @Summary Update tagihan iuran
// @Description Mengupdate data tagihan iuran (Hanya admin/ketua RT)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param id path string true "Tagihan Iuran ID"
// @Param request body UpdateTagihanIuranRequest true "Data update tagihan iuran"
// @Security BearerAuth
// @Success 200 {object} TagihanIuranResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/{id} [put]
func (tic *TagihanIuranController) UpdateTagihan(c *gin.Context) {
	id := c.Param("id")
	
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID tagihan iuran tidak valid",
		})
		return
	}
	
	var req UpdateTagihanIuranRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}
	
	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Gagal menemukan tagihan iuran",
			})
		}
		return
	}
	
	// Cek apakah sudah lunas - tidak bisa update jika sudah lunas
	if tagihan.TagihanIuranStatus == "lunas" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tagihan yang sudah lunas tidak dapat diubah",
		})
		return
	}
	
	// Update fields jika ada
	if req.TagihanIuran != "" {
		tagihan.TagihanIuran = req.TagihanIuran
	}
	if req.TagihanIuranNominal > 0 {
		tagihan.TagihanIuranNominal = req.TagihanIuranNominal
	}
	if req.WargaID > 0 {
		// Validasi warga exists
		var warga models.Warga
		if err := tic.db.First(&warga, req.WargaID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Warga tidak ditemukan",
			})
			return
		}
		tagihan.WargaID = req.WargaID
	}
	
	tagihan.UpdatedAt = time.Now()
	
	if err := tic.db.Save(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengupdate tagihan iuran",
			"details": err.Error(),
		})
		return
	}
	
	// Reload with warga data
	tic.db.Preload("Warga").First(&tagihan, tagihan.ID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tagihan iuran berhasil diupdate",
		"data":    tagihan,
	})
}

// DeleteTagihan godoc
// @Summary Delete tagihan iuran
// @Description Menghapus data tagihan iuran (Hanya admin/ketua RT)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param id path string true "Tagihan Iuran ID"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "success: true, message: 'Tagihan iuran berhasil dihapus'"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/{id} [delete]
func (tic *TagihanIuranController) DeleteTagihan(c *gin.Context) {
	id := c.Param("id")
	
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID tagihan iuran tidak valid",
		})
		return
	}
	
	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Gagal menemukan tagihan iuran",
			})
		}
		return
	}
	
	// Hapus file bukti jika ada
	if tagihan.TagihanBukti != "" {
		helper.DeleteOldPhoto(tagihan.TagihanBukti, "bukti_iuran")
	}
	
	if err := tic.db.Delete(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghapus tagihan iuran",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tagihan iuran berhasil dihapus",
	})
}

// GetByWargaID godoc
// @Summary Get tagihan iuran by warga ID
// @Description Mendapatkan data tagihan iuran berdasarkan ID warga
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param warga_id path string true "Warga ID"
// @Param status query string false "Filter by status (tertagih/lunas)"
// @Param status_verifikasi query string false "Filter by status verifikasi (menunggu/diterima/ditolak)"
// @Security BearerAuth
// @Success 200 {object} TagihanIuranListResponse
// @Failure 400 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/warga/{warga_id} [get]
func (tic *TagihanIuranController) GetByWargaID(c *gin.Context) {
	wargaID := c.Param("warga_id")
	
	wID, err := strconv.ParseUint(wargaID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID warga tidak valid",
		})
		return
	}
	
	var tagihan []models.TagihanIuran
	
	query := tic.db.Preload("Warga").
		Where("warga_id = ?", wID).
		Order("created_at DESC")
	
	// Optional filters
	if status := c.Query("status"); status != "" {
		query = query.Where("tagihan_iuran_status = ?", status)
	}
	
	if statusVerifikasi := c.Query("status_verifikasi"); statusVerifikasi != "" {
		query = query.Where("tagihan_iuran_status_verifikasi = ?", statusVerifikasi)
	}
	
	if err := query.Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengambil data tagihan iuran",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagihan,
	})
}

// UpdateStatusVerifikasi godoc
// @Summary Update status verifikasi tagihan iuran
// @Description Mengupdate status verifikasi tagihan iuran (Hanya admin/ketua RT)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Param id path string true "Tagihan Iuran ID"
// @Param request body UpdateStatusVerifikasiRequest true "Data update status verifikasi"
// @Security BearerAuth
// @Success 200 {object} TagihanIuranResponse
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/{id}/verifikasi [put]
func (tic *TagihanIuranController) UpdateStatusVerifikasi(c *gin.Context) {
	id := c.Param("id")
	
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID tagihan iuran tidak valid",
		})
		return
	}
	
	var req UpdateStatusVerifikasiRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}
	
	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Gagal menemukan tagihan iuran",
			})
		}
		return
	}
	
	// Update status verifikasi
	tagihan.TagihanIuranStatusVerifikasi = req.TagihanIuranStatusVerifikasi
	
	// Jika verifikasi "ditolak" dan status "lunas", ubah status kembali ke "tertagih"
	if req.TagihanIuranStatusVerifikasi == "ditolak" && tagihan.TagihanIuranStatus == "lunas" {
		tagihan.TagihanIuranStatus = "tertagih"
		// Jangan hapus bukti, biarkan untuk referensi
	}
	
	tagihan.UpdatedAt = time.Now()
	
	if err := tic.db.Save(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengupdate status verifikasi",
			"details": err.Error(),
		})
		return
	}
	
	// Reload with warga data
	tic.db.Preload("Warga").First(&tagihan, tagihan.ID)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Status verifikasi berhasil diupdate",
		"data":    tagihan,
	})
}

// BayarIuran godoc
// @Summary Bayar tagihan iuran
// @Description Membayar tagihan iuran dengan mengupload bukti pembayaran
// @Tags Tagihan Iuran
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Tagihan Iuran ID"
// @Param bukti_iuran formData file true "File bukti pembayaran (jpg, jpeg, png, gif, webp)"
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "success: true, message: 'Pembayaran berhasil. Bukti pembayaran sedang diverifikasi.', data: {...}}"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /tagihan-iuran/{id}/bayar [post]
func (tic *TagihanIuranController) BayarIuran(c *gin.Context) {
	id := c.Param("id")
	
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "ID tagihan iuran tidak valid",
		})
		return
	}
	
	var tagihan models.TagihanIuran
	if err := tic.db.Preload("Warga").First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Gagal menemukan tagihan iuran",
			})
		}
		return
	}
	
	// Cek apakah sudah lunas
	if tagihan.TagihanIuranStatus == "lunas" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tagihan iuran ini sudah lunas",
		})
		return
	}
	
	// Handle upload bukti pembayaran menggunakan helper
	buktiFileName, uploadError := helper.HandleFileImageUpload(c, "bukti_iuran", tagihan.TagihanBukti)
	if uploadError != nil && uploadError.Error() != "http: no such file" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Gagal mengupload bukti pembayaran",
			"details": uploadError.Error(),
		})
		return
	}
	
	// Validasi: bukti wajib diupload
	if uploadError != nil && uploadError.Error() == "http: no such file" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Bukti pembayaran wajib diupload",
		})
		return
	}
	
	// Update tagihan
	tagihan.TagihanIuranStatus = "lunas"
	tagihan.TagihanIuranStatusVerifikasi = "menunggu"
	if buktiFileName != "" {
		tagihan.TagihanBukti = buktiFileName
	}
	tagihan.UpdatedAt = time.Now()
	
	if err := tic.db.Save(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal memproses pembayaran",
			"details": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Pembayaran berhasil. Bukti pembayaran sedang diverifikasi.",
		"data": gin.H{
			"id":                           tagihan.ID,
			"tagihan_iuran":                tagihan.TagihanIuran,
			"tagihan_iuran_nominal":        tagihan.TagihanIuranNominal,
			"tagihan_iuran_status":         tagihan.TagihanIuranStatus,
			"tagihan_iuran_status_verifikasi": tagihan.TagihanIuranStatusVerifikasi,
			"tagihan_bukti":                tagihan.TagihanBukti,
			"bukti_url":                    "/api/files/iuran/" + tagihan.TagihanBukti,
			"warga":                        tagihan.Warga,
			"updated_at":                   tagihan.UpdatedAt,
		},
	})
}

// GetFileBukti godoc
// @Summary Get file bukti pembayaran
// @Description Mendapatkan file bukti pembayaran tagihan iuran
// @Tags Tagihan Iuran
// @Accept json
// @Produce image/*,application/octet-stream
// @Param filename path string true "Nama file bukti"
// @Success 200 {file} file "File bukti pembayaran"
// @Failure 400 {object} ErrorResponse
// @Failure 404 {object} ErrorResponse
// @Failure 500 {object} ErrorResponse
// @Router /api/tagihan-iuran/files/{filename} [get]
func (tic *TagihanIuranController) GetFileBukti(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nama file tidak valid",
		})
		return
	}
	
	file, err := helper.GetFileByFileName("bukti_iuran", filename)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "File tidak ditemukan",
		})
		return
	}
	defer file.Close()
	
	// Get file info
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal membaca file",
		})
		return
	}
	
	// Determine content type
	ext := filepath.Ext(filename)
	contentType := helper.GetContentType(ext)
	
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", "inline; filename=\""+filename+"\"")
	
	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}