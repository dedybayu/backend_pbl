// controllers/keluarga_controller.go
package controllers

import (
	"log"
	"net/http"
	"regexp"
	"rt-management/models"
	"strconv"
	"strings"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type KeluargaController struct {
	db *gorm.DB
}

func NewKeluargaController(db *gorm.DB) *KeluargaController {
	if db == nil {
		log.Fatal("❌ Database connection is nil in KeluargaController")
	}
	return &KeluargaController{db: db}
}

type CreateKeluargaRequest struct {
	KeluargaNama   string `json:"keluarga_nama" binding:"required"`
	KeluargaStatus string `json:"keluarga_status"`
}

type UpdateKeluargaRequest struct {
	KeluargaNama   string `json:"keluarga_nama"`
	KeluargaStatus string `json:"keluarga_status"`
}

// ✅ Security validation functions
func isValidKeluargaName(name string) bool {
	// Nama keluarga harus 2-100 karakter, hanya huruf, angka, spasi, dan karakter umum
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	matched, _ := regexp.MatchString("^[a-zA-Z0-9\\s\\-.,()]+$", name)
	return matched
}

func isValidStatus(status string) bool {
	// Hanya menerima status yang valid
	validStatuses := map[string]bool{
		"aktif":    true,
		"nonaktif": true,
	}
	return validStatuses[status]
}

func isValidKeluargaID(id string) bool {
	// Validasi ID adalah angka positif
	parsedID, err := strconv.ParseUint(id, 10, 32)
	return err == nil && parsedID > 0
}

func sanitizeSearchQuery(query string) string {
	// Sanitize khusus untuk search query
	reg := regexp.MustCompile(`[<>"'%;()&+*|=/\\]`)
	sanitized := reg.ReplaceAllString(query, "")
	return strings.TrimSpace(sanitized)
}

// GetAllKeluarga returns all keluarga dengan security checks
func (kc *KeluargaController) GetAllKeluarga(c *gin.Context) {
	var families []models.Keluarga

	log.Println("🔄 Fetching all families from database...")

	// ✅ SAFE: GORM menggunakan parameterized queries
	if err := kc.db.Preload("Wargas").Find(&families).Error; err != nil {
		log.Printf("❌ Error fetching families: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch families",
		})
		return
	}

	log.Printf("✅ Successfully fetched %d families", len(families))
	c.JSON(http.StatusOK, gin.H{
		"data":  families,
		"count": len(families),
	})
}

// GetKeluargaByID returns keluarga by ID dengan security validation
func (kc *KeluargaController) GetKeluargaByID(c *gin.Context) {
	keluargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidKeluargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Fetching family with ID: %s", keluargaID)

	var keluarga models.Keluarga
	// ✅ SAFE: GORM menggunakan prepared statements
	if err := kc.db.Preload("Wargas").First(&keluarga, keluargaID).Error; err != nil {
		log.Printf("❌ Error fetching family %s: %v", keluargaID, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Family not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch family",
			})
		}
		return
	}

	log.Printf("✅ Successfully fetched family: %s", keluarga.KeluargaNama)
	c.JSON(http.StatusOK, keluarga)
}

// GetKeluargaWithDetails returns keluarga dengan detail lengkap
func (kc *KeluargaController) GetKeluargaWithDetails(c *gin.Context) {
	keluargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidKeluargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Fetching family details with ID: %s", keluargaID)

	var keluarga models.Keluarga
	// ✅ SAFE: Semua menggunakan parameterized queries
	if err := kc.db.
		Preload("Wargas", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Agama").Preload("Pekerjaan")
		}).
		First(&keluarga, keluargaID).Error; err != nil {
		log.Printf("❌ Error fetching family details %s: %v", keluargaID, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Family not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch family details",
			})
		}
		return
	}

	log.Printf("✅ Successfully fetched family details: %s", keluarga.KeluargaNama)
	c.JSON(http.StatusOK, keluarga)
}

// CreateKeluarga creates new keluarga dengan comprehensive security checks
func (kc *KeluargaController) CreateKeluarga(c *gin.Context) {
	var req CreateKeluargaRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	req.KeluargaNama = strings.TrimSpace(req.KeluargaNama)
	req.KeluargaStatus = strings.TrimSpace(req.KeluargaStatus)

	// ✅ Validasi nama keluarga
	if !isValidKeluargaName(req.KeluargaNama) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Family name must be 2-100 characters and contain only letters, numbers, spaces, and common punctuation",
		})
		return
	}

	// Set default status jika tidak provided
	if req.KeluargaStatus == "" {
		req.KeluargaStatus = "aktif"
	}

	// ✅ Validasi status
	if !isValidStatus(req.KeluargaStatus) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status must be 'aktif' or 'nonaktif'",
		})
		return
	}

	log.Printf("🔄 Creating new family: %s", req.KeluargaNama)

	keluarga := models.Keluarga{
		KeluargaNama:   req.KeluargaNama,
		KeluargaStatus: req.KeluargaStatus,
	}

	// ✅ SAFE: GORM Create dengan parameterized queries
	if err := kc.db.Create(&keluarga).Error; err != nil {
		log.Printf("❌ Error creating family: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create family",
		})
		return
	}

	log.Printf("✅ Successfully created family: %s (ID: %d)", keluarga.KeluargaNama, keluarga.KeluargaID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Family created successfully",
		"data":    keluarga,
	})
}

// UpdateKeluarga updates keluarga data dengan security checks
func (kc *KeluargaController) UpdateKeluarga(c *gin.Context) {
	keluargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidKeluargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Updating family with ID: %s", keluargaID)

	var req UpdateKeluargaRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	if req.KeluargaNama != "" {
		req.KeluargaNama = strings.TrimSpace(req.KeluargaNama)
	}
	if req.KeluargaStatus != "" {
		req.KeluargaStatus = strings.TrimSpace(req.KeluargaStatus)
	}

	var keluarga models.Keluarga
	// ✅ SAFE: GORM First dengan parameterized query
	if err := kc.db.First(&keluarga, keluargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Family not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch family"})
		}
		return
	}

	// ✅ Update fields dengan validasi
	if req.KeluargaNama != "" {
		if !isValidKeluargaName(req.KeluargaNama) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Family name must be 2-100 characters and contain only letters, numbers, spaces, and common punctuation",
			})
			return
		}
		keluarga.KeluargaNama = req.KeluargaNama
	}

	if req.KeluargaStatus != "" {
		// ✅ Validasi status
		if !isValidStatus(req.KeluargaStatus) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status must be 'aktif' or 'nonaktif'",
			})
			return
		}
		keluarga.KeluargaStatus = req.KeluargaStatus
	}

	// ✅ SAFE: GORM Save dengan parameterized queries
	if err := kc.db.Save(&keluarga).Error; err != nil {
		log.Printf("❌ Error updating family: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update family",
		})
		return
	}

	// ✅ Reload dengan data terbaru - SAFE
	if err := kc.db.Preload("Wargas").First(&keluarga, keluargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated family",
		})
		return
	}

	log.Printf("✅ Successfully updated family: %s", keluarga.KeluargaNama)
	c.JSON(http.StatusOK, gin.H{
		"message": "Family updated successfully",
		"data":    keluarga,
	})
}

// DeleteKeluarga deletes keluarga dengan security checks
func (kc *KeluargaController) DeleteKeluarga(c *gin.Context) {
	keluargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidKeluargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Deleting family with ID: %s", keluargaID)

	var keluarga models.Keluarga
	// ✅ SAFE: GORM First dengan parameterized query
	if err := kc.db.First(&keluarga, keluargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Family not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch family"})
		}
		return
	}

	// ✅ Check jika keluarga masih memiliki warga - SAFE: parameterized query
	var wargaCount int64
	if err := kc.db.Model(&models.Warga{}).Where("keluarga_id = ?", keluargaID).Count(&wargaCount).Error; err != nil {
		log.Printf("❌ Error checking family members: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check family members",
		})
		return
	}

	if wargaCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":        "Cannot delete family that still has members",
			"member_count": wargaCount,
		})
		return
	}

	// ✅ SAFE: GORM Delete dengan parameterized query
	if err := kc.db.Delete(&keluarga).Error; err != nil {
		log.Printf("❌ Error deleting family: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete family",
		})
		return
	}

	log.Printf("✅ Successfully deleted family: %s (ID: %d)", keluarga.KeluargaNama, keluarga.KeluargaID)
	c.JSON(http.StatusOK, gin.H{
		"message":           "Family deleted successfully",
		"deleted_family_id": keluarga.KeluargaID,
	})
}

// GetKeluargaStats returns statistics about families dengan security
func (kc *KeluargaController) GetKeluargaStats(c *gin.Context) {
	var stats struct {
		TotalKeluarga    int64 `json:"total_keluarga"`
		KeluargaAktif    int64 `json:"keluarga_aktif"`
		KeluargaNonaktif int64 `json:"keluarga_nonaktif"`
		TotalWarga       int64 `json:"total_warga"`
	}

	// ✅ SAFE: Semua count queries menggunakan parameterized queries internally
	kc.db.Model(&models.Keluarga{}).Count(&stats.TotalKeluarga)
	kc.db.Model(&models.Keluarga{}).Where("keluarga_status = ?", "aktif").Count(&stats.KeluargaAktif)
	kc.db.Model(&models.Keluarga{}).Where("keluarga_status = ?", "nonaktif").Count(&stats.KeluargaNonaktif)
	kc.db.Model(&models.Warga{}).Count(&stats.TotalWarga)

	log.Printf("📊 Family stats: Total=%d, Active=%d, Inactive=%d, Members=%d",
		stats.TotalKeluarga, stats.KeluargaAktif, stats.KeluargaNonaktif, stats.TotalWarga)

	c.JSON(http.StatusOK, stats)
}

// SearchKeluarga searches families by name dengan security enhancements
func (kc *KeluargaController) SearchKeluarga(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	// ✅ Sanitize search query
	query = sanitizeSearchQuery(query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid search query",
		})
		return
	}

	// ✅ Limit query length untuk prevent abuse
	if len(query) > 50 {
		query = query[:50]
	}

	log.Printf("🔍 Searching families with query: %s", query)

	var families []models.Keluarga

	// ✅ SAFE: Gunakan parameterized query
	if err := kc.db.
		Preload("Wargas").
		Where("keluarga_nama LIKE ?", "%"+query+"%").
		Find(&families).Error; err != nil {
		log.Printf("❌ Error searching families: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search families",
		})
		return
	}

	log.Printf("✅ Search completed: found %d families for query '%s'", len(families), query)

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": families,
		"count":   len(families),
	})
}

// GetKeluargaAktif returns only active families
func (kc *KeluargaController) GetKeluargaAktif(c *gin.Context) {
	var families []models.Keluarga

	// ✅ SAFE: Parameterized query
	if err := kc.db.
		Preload("Wargas").
		Where("keluarga_status = ?", "aktif").
		Find(&families).Error; err != nil {
		log.Printf("❌ Error fetching active families: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch active families",
		})
		return
	}

	log.Printf("✅ Fetched %d active families", len(families))
	c.JSON(http.StatusOK, gin.H{
		"data":  families,
		"count": len(families),
	})
}

// GetTotalKeluarga returns the total number of families
func (kc *KeluargaController) GetTotalKeluarga(c *gin.Context) {
	var total int64
	// ✅ SAFE: Parameterized query
	if err := kc.db.Model(&models.Keluarga{}).Count(&total).Error; err != nil {
		log.Printf("❌ Error fetching total families: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch total families",
		})
		return
	}

	log.Printf("✅ Fetched total families: %d", total)
	c.JSON(http.StatusOK, gin.H{
		"data":  total,
		"count": total,
	})
}
