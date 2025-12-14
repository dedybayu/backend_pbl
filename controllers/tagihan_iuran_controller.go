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

type TagihanIuranController struct {
	db *gorm.DB
}

func NewTagihanIuranController(db *gorm.DB) *TagihanIuranController {
	return &TagihanIuranController{db: db}
}

// Request structs
type CreateTagihanIuranRequest struct {
	TagihanIuran        string  `form:"tagihan_iuran" binding:"required"`
	TagihanIuranNominal float64 `form:"tagihan_iuran_nominal" binding:"required"`
	WargaID             uint    `form:"warga_id" binding:"required"`
}

type UpdateTagihanIuranRequest struct {
	TagihanIuran        string  `form:"tagihan_iuran"`
	TagihanIuranNominal float64 `form:"tagihan_iuran_nominal"`
	WargaID             uint    `form:"warga_id"`
}

type BayarIuranRequest struct {
	TagihanIuranStatus string  `form:"tagihan_iuran_status" binding:"required,oneof=lunas tertagih"`
	JumlahBayar        float64 `form:"jumlah_bayar" binding:"required"`
	MetodePembayaran   string  `form:"metode_pembayaran"`
	Keterangan         string  `form:"keterangan"`
}

// CreateTagihanIuran godoc
// @Summary Membuat tagihan iuran baru
// @Description Membuat data tagihan iuran baru
// @Tags Tagihan Iuran
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param tagihan_iuran formData string true "Nama tagihan iuran"
// @Param tagihan_iuran_nominal formData number true "Nominal tagihan"
// @Param warga_id formData integer true "ID Warga"
// @Success 201 {object} map[string]interface{} "Tagihan iuran berhasil dibuat"
// @Failure 400 {object} map[string]interface{} "Data tidak valid"
// @Failure 409 {object} map[string]interface{} "Tagihan iuran sudah ada"
// @Failure 500 {object} map[string]interface{} "Gagal membuat tagihan iuran"
// @Router /api/tagihan-iuran [post]
func (tic *TagihanIuranController) CreateTagihanIuran(c *gin.Context) {
	var req CreateTagihanIuranRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.TagihanIuran = strings.TrimSpace(req.TagihanIuran)

	// Validasi required fields
	if req.TagihanIuran == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nama tagihan iuran harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.TagihanIuran) < 2 || len(req.TagihanIuran) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nama tagihan iuran harus 2-100 karakter",
		})
		return
	}

	// Validasi nominal
	if req.TagihanIuranNominal <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nominal tagihan harus lebih dari 0",
		})
		return
	}

	// Validasi warga_id
	var warga models.Warga
	if err := tic.db.First(&warga, req.WargaID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Warga tidak ditemukan",
		})
		return
	}

	// Check if tagihan iuran dengan nama yang sama sudah ada untuk warga yang sama
	var existingTagihan models.TagihanIuran
	if err := tic.db.Where("tagihan_iuran = ? AND warga_id = ?", req.TagihanIuran, req.WargaID).First(&existingTagihan).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Tagihan iuran dengan nama tersebut sudah ada untuk warga ini",
		})
		return
	}

	// Buat tagihan iuran baru
	tagihan := models.TagihanIuran{
		TagihanIuran:        req.TagihanIuran,
		TagihanIuranNominal: req.TagihanIuranNominal,
		TagihanIuranStatus:  "tertagih", // default status
		WargaID:             req.WargaID,
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := tic.db.Create(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal membuat tagihan iuran",
			"details": err.Error(),
		})
		return
	}

	// Preload warga data untuk response
	tic.db.Preload("Warga").First(&tagihan, tagihan.ID)

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "Tagihan iuran berhasil dibuat",
		"data":    tagihan,
	})
}

// GetAllTagihanIuran godoc
// @Summary Mendapatkan semua tagihan iuran
// @Description Mengambil seluruh data tagihan iuran dengan pagination dan filter
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Nomor halaman" minimum(1) default(1)
// @Param limit query int false "Jumlah data per halaman" minimum(1) maximum(100) default(10)
// @Param search query string false "Kata kunci pencarian"
// @Param status query string false "Filter status (tertagih/lunas)"
// @Param warga_id query int false "Filter berdasarkan ID warga"
// @Success 200 {object} map[string]interface{} "Data tagihan iuran berhasil diambil"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data tagihan iuran"
// @Router /api/tagihan-iuran [get]
func (tic *TagihanIuranController) GetAllTagihanIuran(c *gin.Context) {
	var tagihan []models.TagihanIuran

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	search := strings.TrimSpace(c.Query("search"))
	status := strings.TrimSpace(c.Query("status"))
	wargaID := c.Query("warga_id")

	// Build query dengan preload warga
	query := tic.db.Preload("Warga").Model(&models.TagihanIuran{})

	// Apply search filter
	if search != "" {
		query = query.Where("tagihan_iuran LIKE ?", "%"+search+"%")
	}

	// Apply status filter
	if status != "" && (status == "tertagih" || status == "lunas") {
		query = query.Where("tagihan_iuran_status = ?", status)
	}

	// Apply warga_id filter
	if wargaID != "" {
		if wargaIDUint, err := strconv.ParseUint(wargaID, 10, 32); err == nil {
			query = query.Where("warga_id = ?", wargaIDUint)
		}
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Execute query dengan pagination
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengambil data tagihan iuran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagihan,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
			"pages": (int(total) + limit - 1) / limit,
		},
	})
}

// GetTagihanIuranByID godoc
// @Summary Mendapatkan tagihan iuran berdasarkan ID
// @Description Mendapatkan detail tagihan iuran berdasarkan ID
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Tagihan Iuran"
// @Success 200 {object} map[string]interface{} "Data tagihan iuran berhasil diambil"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Tagihan iuran tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data tagihan iuran"
// @Router /api/tagihan-iuran/{id} [get]
func (tic *TagihanIuranController) GetTagihanIuranByID(c *gin.Context) {
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

// UpdateTagihanIuran godoc
// @Summary Mengupdate tagihan iuran
// @Description Mengupdate data tagihan iuran berdasarkan ID
// @Tags Tagihan Iuran
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Tagihan Iuran"
// @Param tagihan_iuran formData string false "Nama tagihan iuran"
// @Param tagihan_iuran_nominal formData number false "Nominal tagihan"
// @Param warga_id formData integer false "ID Warga"
// @Success 200 {object} map[string]interface{} "Tagihan iuran berhasil diupdate"
// @Failure 400 {object} map[string]interface{} "Data tidak valid"
// @Failure 404 {object} map[string]interface{} "Tagihan iuran tidak ditemukan"
// @Failure 409 {object} map[string]interface{} "Tagihan iuran dengan nama tersebut sudah ada"
// @Failure 500 {object} map[string]interface{} "Gagal mengupdate tagihan iuran"
// @Router /api/tagihan-iuran/{id} [put]
func (tic *TagihanIuranController) UpdateTagihanIuran(c *gin.Context) {
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

	var req UpdateTagihanIuranRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Hanya update field yang diisi
	updates := make(map[string]interface{} )
	
	if req.TagihanIuran != "" {
		req.TagihanIuran = strings.TrimSpace(req.TagihanIuran)
		
		if len(req.TagihanIuran) < 2 || len(req.TagihanIuran) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Nama tagihan iuran harus 2-100 karakter",
			})
			return
		}
		
		// Check duplicate jika nama diubah
		if req.TagihanIuran != tagihan.TagihanIuran {
			var existingTagihan models.TagihanIuran
			if err := tic.db.Where("tagihan_iuran = ? AND warga_id = ? AND id != ?", 
				req.TagihanIuran, tagihan.WargaID, tagihanID).First(&existingTagihan).Error; err == nil {
				c.JSON(http.StatusConflict, gin.H{
					"success": false,
					"error":   "Tagihan iuran dengan nama tersebut sudah ada untuk warga ini",
				})
				return
			}
		}
		
		updates["tagihan_iuran"] = req.TagihanIuran
	}
	
	if req.TagihanIuranNominal > 0 {
		updates["tagihan_iuran_nominal"] = req.TagihanIuranNominal
	} else if req.TagihanIuranNominal < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Nominal tagihan harus lebih dari 0",
		})
		return
	}
	
	if req.WargaID > 0 {
		// Validasi warga_id
		var warga models.Warga
		if err := tic.db.First(&warga, req.WargaID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Warga tidak ditemukan",
			})
			return
		}
		updates["warga_id"] = req.WargaID
	}
	
	if len(updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Tidak ada data yang diupdate",
		})
		return
	}
	
	updates["updated_at"] = time.Now()

	// Update menggunakan map untuk menghindari zero value issues
	if err := tic.db.Model(&tagihan).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengupdate tagihan iuran",
			"details": err.Error(),
		})
		return
	}

	// Reload data dengan preload
	tic.db.Preload("Warga").First(&tagihan, tagihan.ID)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Tagihan iuran berhasil diupdate",
		"data":    tagihan,
	})
}

// DeleteTagihanIuran godoc
// @Summary Menghapus tagihan iuran
// @Description Menghapus tagihan iuran berdasarkan ID
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Tagihan Iuran"
// @Success 200 {object} map[string]interface{} "Tagihan iuran berhasil dihapus"
// @Failure 400 {object} map[string]interface{} "ID tidak valid"
// @Failure 404 {object} map[string]interface{} "Tagihan iuran tidak ditemukan"
// @Failure 500 {object} map[string]interface{} "Gagal menghapus tagihan iuran"
// @Router /api/tagihan-iuran/{id} [delete]
func (tic *TagihanIuranController) DeleteTagihanIuran(c *gin.Context) {
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

// GetTagihanIuranDropdown godoc
// @Summary Mendapatkan dropdown tagihan iuran
// @Description Mengambil data untuk dropdown tagihan iuran (id dan nama saja)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Data dropdown berhasil diambil"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil data dropdown"
// @Router /api/tagihan-iuran/dropdown [get]
func (tic *TagihanIuranController) GetTagihanIuranDropdown(c *gin.Context) {
	var tagihan []struct {
		ID           uint   `json:"id"`
		TagihanIuran string `json:"tagihan_iuran"`
	}

	if err := tic.db.
		Model(&models.TagihanIuran{}).
		Select("id, tagihan_iuran").
		Order("tagihan_iuran ASC").
		Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal mengambil data dropdown tagihan iuran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tagihan,
	})
}

// BayarIuran godoc
// @Summary Melakukan pembayaran iuran
// @Description Melakukan pembayaran untuk tagihan iuran berdasarkan ID
// @Tags Tagihan Iuran
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID Tagihan Iuran"
// @Param tagihan_iuran_status formData string true "Status pembayaran (lunas/tertagih)" Enums(lunas, tertagih)
// @Param jumlah_bayar formData number true "Jumlah yang dibayar"
// @Param metode_pembayaran formData string false "Metode pembayaran"
// @Param keterangan formData string false "Keterangan pembayaran"
// @Success 200 {object} map[string]interface{} "Pembayaran berhasil diproses"
// @Failure 400 {object} map[string]interface{} "Data tidak valid"
// @Failure 404 {object} map[string]interface{} "Tagihan iuran tidak ditemukan"
// @Failure 422 {object} map[string]interface{} "Jumlah bayar tidak sesuai"
// @Failure 500 {object} map[string]interface{} "Gagal memproses pembayaran"
// @Router /api/tagihan-iuran/{id}/bayar [post]
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

	var req BayarIuranRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Data request tidak valid",
			"details": err.Error(),
		})
		return
	}

	// Validasi jumlah bayar
	if req.JumlahBayar <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Jumlah bayar harus lebih dari 0",
		})
		return
	}

	// Jika status lunas, pastikan jumlah bayar sesuai dengan nominal tagihan
	if req.TagihanIuranStatus == "lunas" {
		if req.JumlahBayar < tagihan.TagihanIuranNominal {
			c.JSON(http.StatusUnprocessableEntity, gin.H{
				"success": false,
				"error":   "Jumlah bayar kurang dari nominal tagihan",
				"detail": gin.H{
					"nominal_tagihan": tagihan.TagihanIuranNominal,
					"jumlah_bayar":    req.JumlahBayar,
					"kekurangan":      tagihan.TagihanIuranNominal - req.JumlahBayar,
				},
			})
			return
		}
		
		// Jika bayar lebih, tetap dianggap lunas (kelebihan sebagai kelebihan bayar)
		// Anda bisa mencatat kelebihan bayar di keterangan
		if req.JumlahBayar > tagihan.TagihanIuranNominal {
			if req.Keterangan == "" {
				req.Keterangan = "Kelebihan bayar: " + strconv.FormatFloat(req.JumlahBayar-tagihan.TagihanIuranNominal, 'f', 2, 64)
			} else {
				req.Keterangan += " (Kelebihan bayar: " + strconv.FormatFloat(req.JumlahBayar-tagihan.TagihanIuranNominal, 'f', 2, 64) + ")"
			}
		}
	}

	// Update status tagihan
	tagihan.TagihanIuranStatus = req.TagihanIuranStatus
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
		"message": "Pembayaran iuran berhasil diproses",
		"data": gin.H{
			"tagihan_iuran":     tagihan,
			"jumlah_bayar":      req.JumlahBayar,
			"status":            req.TagihanIuranStatus,
			"metode_pembayaran": req.MetodePembayaran,
			"keterangan":        req.Keterangan,
			"tanggal_bayar":     time.Now(),
			"kembalian":         req.JumlahBayar - tagihan.TagihanIuranNominal,
		},
	})
}

// GetStatistikTagihanIuran godoc
// @Summary Mendapatkan statistik tagihan iuran
// @Description Mendapatkan statistik tagihan iuran (total, lunas, tertagih, total nominal)
// @Tags Tagihan Iuran
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param warga_id query int false "Filter berdasarkan ID warga"
// @Success 200 {object} map[string]interface{} "Statistik berhasil diambil"
// @Failure 500 {object} map[string]interface{} "Gagal mengambil statistik"
// @Router /api/tagihan-iuran/statistik [get]
func (tic *TagihanIuranController) GetStatistikTagihanIuran(c *gin.Context) {
	wargaID := c.Query("warga_id")

	var statistik struct {
		Total                int64   `json:"total"`
		TotalLunas           int64   `json:"total_lunas"`
		TotalTertagih        int64   `json:"total_tertagih"`
		TotalNominal         float64 `json:"total_nominal"`
		TotalLunasNominal    float64 `json:"total_lunas_nominal"`
		TotalTertagihNominal float64 `json:"total_tertagih_nominal"`
	}

	// Query base
	query := tic.db.Model(&models.TagihanIuran{})

	// Filter by warga_id jika ada
	if wargaID != "" {
		if wargaIDUint, err := strconv.ParseUint(wargaID, 10, 32); err == nil {
			query = query.Where("warga_id = ?", wargaIDUint)
		}
	}

	// Hitung total
	if err := query.Count(&statistik.Total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total tagihan",
		})
		return
	}

	// Hitung total lunas
	if err := tic.db.Model(&models.TagihanIuran{}).
		Where("tagihan_iuran_status = ?", "lunas").
		Count(&statistik.TotalLunas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total lunas",
		})
		return
	}

	// Hitung total tertagih
	if err := tic.db.Model(&models.TagihanIuran{}).
		Where("tagihan_iuran_status = ?", "tertagih").
		Count(&statistik.TotalTertagih).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total tertagih",
		})
		return
	}

	// Hitung total nominal
	if err := tic.db.Model(&models.TagihanIuran{}).
		Select("COALESCE(SUM(tagihan_iuran_nominal), 0) as total_nominal").
		Scan(&statistik.TotalNominal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total nominal",
		})
		return
	}

	// Hitung total nominal lunas
	if err := tic.db.Model(&models.TagihanIuran{}).
		Select("COALESCE(SUM(tagihan_iuran_nominal), 0) as total_lunas_nominal").
		Where("tagihan_iuran_status = ?", "lunas").
		Scan(&statistik.TotalLunasNominal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total nominal lunas",
		})
		return
	}

	// Hitung total nominal tertagih
	if err := tic.db.Model(&models.TagihanIuran{}).
		Select("COALESCE(SUM(tagihan_iuran_nominal), 0) as total_tertagih_nominal").
		Where("tagihan_iuran_status = ?", "tertagih").
		Scan(&statistik.TotalTertagihNominal).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Gagal menghitung total nominal tertagih",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    statistik,
	})
}