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

// CreateRumah creates a new house
// @Summary      Create a new house
// @Description  Create a new house with address and status
// @Tags         houses
// @Accept       multipart/form-data
// @Produce      json
// @Param        rumah_alamat formData string true "House address"
// @Param        rumah_status formData string false "House status" Enums(tersedia, ditempati)
// @Success      201 {object} map[string]interface{} "House created successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request data"
// @Failure      500 {object} map[string]interface{} "Failed to create house"
// @Router       /rumah [post]
func (rc *RumahController) CreateRumah(c *gin.Context) {
	var req CreateRumahRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
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

// GetAllRumah retrieves all houses with optional filtering
// @Summary      Get all houses
// @Description  Get list of houses with optional filtering by status or warga ID
// @Tags         houses
// @Accept       json
// @Produce      json
// @Param        page    query int    false "Page number (disabled by default)"
// @Param        limit   query int    false "Items per page (disabled by default, use 0 for no limit)"
// @Param        status  query string false "Filter by status" Enums(tersedia, ditempati)
// @Param        warga_id query string false "Filter by warga ID"
// @Param        all     query bool   false "Get all records without pagination"
// @Success      200 {object} map[string]interface{} "List of houses"
// @Failure      500 {object} map[string]interface{} "Failed to get houses"
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
		response["message"] = "All records retrieved (pagination disabled)"
	}

	c.JSON(http.StatusOK, response)
}

// GetRumahByID retrieves a house by ID
// @Summary      Get house by ID
// @Description  Get house details including warga data by house ID
// @Tags         houses
// @Accept       json
// @Produce      json
// @Param        id path string true "House ID"
// @Success      200 {object} map[string]interface{} "House details"
// @Failure      404 {object} map[string]interface{} "House not found"
// @Router       /rumah/{id} [get]
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

// UpdateRumah updates a house by ID
// @Summary      Update a house
// @Description  Update house information by ID
// @Tags         houses
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path string true "House ID"
// @Param        rumah_alamat formData string false "House address"
// @Param        rumah_status formData string false "House status" Enums(tersedia, ditempati)
// @Success      200 {object} map[string]interface{} "House updated successfully"
// @Failure      400 {object} map[string]interface{} "Invalid request data"
// @Failure      404 {object} map[string]interface{} "House not found"
// @Failure      500 {object} map[string]interface{} "Failed to update house"
// @Router       /rumah/{id} [put]
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
			"error":   "Invalid request data",
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
			"error":   "Gagal update rumah",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Rumah berhasil diupdate",
		"data":    rumah,
	})
}

// DeleteRumah deletes a house by ID
// @Summary      Delete a house
// @Description  Delete house by ID
// @Tags         houses
// @Accept       json
// @Produce      json
// @Param        id path string true "House ID"
// @Success      200 {object} map[string]interface{} "House deleted successfully"
// @Failure      404 {object} map[string]interface{} "House not found"
// @Failure      500 {object} map[string]interface{} "Failed to delete house"
// @Router       /rumah/{id} [delete]
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

// GetRumahByStatus retrieves houses by status
// @Summary      Get houses by status
// @Description  Get list of houses filtered by status
// @Tags         houses
// @Accept       json
// @Produce      json
// @Param        status path string true "House status" Enums(tersedia, ditempati)
// @Success      200 {object} map[string]interface{} "List of houses by status"
// @Failure      400 {object} map[string]interface{} "Invalid status"
// @Failure      500 {object} map[string]interface{} "Failed to get houses"
// @Router       /rumah/status/{status} [get]
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