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
	TagihanIuran string `form:"tagihan_iuran" binding:"required"`
}

type UpdateTagihanIuranRequest struct {
	TagihanIuran string `form:"tagihan_iuran" binding:"required"`
}

// ✅ CREATE - Membuat tagihan iuran baru
func (tic *TagihanIuranController) CreateTagihanIuran(c *gin.Context) {
	var req CreateTagihanIuranRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.TagihanIuran = strings.TrimSpace(req.TagihanIuran)

	// Validasi required fields
	if req.TagihanIuran == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama tagihan iuran harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.TagihanIuran) < 2 || len(req.TagihanIuran) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama tagihan iuran harus 2-100 karakter",
		})
		return
	}

	// Check if tagihan iuran dengan nama yang sama sudah ada
	var existingTagihan models.TagihanIuran
	if err := tic.db.Where("tagihan_iuran = ?", req.TagihanIuran).First(&existingTagihan).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tagihan iuran dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat tagihan iuran baru
	tagihan := models.TagihanIuran{
		TagihanIuran: req.TagihanIuran,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := tic.db.Create(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat tagihan iuran",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Tagihan iuran berhasil dibuat",
		"data":    tagihan,
	})
}

// ✅ READ - Mendapatkan semua tagihan iuran
func (tic *TagihanIuranController) GetAllTagihanIuran(c *gin.Context) {
	var tagihan []models.TagihanIuran

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := tic.db.Model(&models.TagihanIuran{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("tagihan_iuran LIKE ?", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data tagihan iuran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tagihan,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan tagihan iuran by ID
func (tic *TagihanIuranController) GetTagihanIuranByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tagihan iuran tidak valid",
		})
		return
	}

	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data tagihan iuran",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tagihan,
	})
}

// ✅ UPDATE - Mengupdate tagihan iuran
func (tic *TagihanIuranController) UpdateTagihanIuran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tagihan iuran tidak valid",
		})
		return
	}

	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan tagihan iuran",
			})
		}
		return
	}

	var req UpdateTagihanIuranRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.TagihanIuran = strings.TrimSpace(req.TagihanIuran)

	// Validasi required fields
	if req.TagihanIuran == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama tagihan iuran harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.TagihanIuran) < 2 || len(req.TagihanIuran) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama tagihan iuran harus 2-100 karakter",
		})
		return
	}

	// Check if tagihan iuran dengan nama yang sama sudah ada (exclude current)
	var existingTagihan models.TagihanIuran
	if err := tic.db.Where("tagihan_iuran = ? AND id != ?", req.TagihanIuran, tagihanID).First(&existingTagihan).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tagihan iuran dengan nama tersebut sudah ada",
		})
		return
	}

	// Update tagihan iuran
	tagihan.TagihanIuran = req.TagihanIuran
	tagihan.UpdatedAt = time.Now()

	if err := tic.db.Save(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate tagihan iuran",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tagihan iuran berhasil diupdate",
		"data":    tagihan,
	})
}

// ✅ DELETE - Menghapus tagihan iuran
func (tic *TagihanIuranController) DeleteTagihanIuran(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	tagihanID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID tagihan iuran tidak valid",
		})
		return
	}

	var tagihan models.TagihanIuran
	if err := tic.db.First(&tagihan, tagihanID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Tagihan iuran tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan tagihan iuran",
			})
		}
		return
	}

	// Delete menggunakan GORM Delete (AMAN)
	if err := tic.db.Delete(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus tagihan iuran",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Tagihan iuran berhasil dihapus",
	})
}

// ✅ GET - Dropdown tagihan iuran
func (tic *TagihanIuranController) GetTagihanIuranDropdown(c *gin.Context) {
	var tagihan []struct {
		ID           uint   `json:"id"`
		TagihanIuran string `json:"tagihan_iuran"`
	}

	// Query aman - hanya mengambil field yang diperlukan
	if err := tic.db.
		Model(&models.TagihanIuran{}).
		Select("id, tagihan_iuran").
		Order("tagihan_iuran ASC").
		Find(&tagihan).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data dropdown tagihan iuran",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": tagihan,
	})
}