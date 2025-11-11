// controllers/warga_controller.go
package controllers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"rt-management/models"
	"log"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type WargaController struct {
	db *gorm.DB
}

func NewWargaController(db *gorm.DB) *WargaController {
	if db == nil {
		log.Fatal("❌ Database connection is nil in WargaController")
	}
	return &WargaController{db: db}
}

type CreateWargaRequest struct {
	KeluargaID        uint      `json:"keluarga_id" binding:"required"`
	WargaNama         string    `json:"warga_nama" binding:"required"`
	WargaNIK          string    `json:"warga_nik" binding:"required"`
	WargaNoTlp        string    `json:"warga_no_tlp"`
	WargaTempatLahir  string    `json:"warga_tempat_lahir"`
	WargaTanggalLahir time.Time `json:"warga_tanggal_lahir"`
	WargaJenisKelamin string    `json:"warga_jenis_kelamin" binding:"required"`
	WargaStatusAktif  string    `json:"warga_status_aktif"`
	WargaStatusHidup  string    `json:"warga_status_hidup"`
	AgamaID           uint      `json:"agama_id"`
	PekerjaanID       uint      `json:"pekerjaan_id"`
}

type UpdateWargaRequest struct {
	KeluargaID        uint      `json:"keluarga_id"`
	WargaNama         string    `json:"warga_nama"`
	WargaNIK          string    `json:"warga_nik"`
	WargaNoTlp        string    `json:"warga_no_tlp"`
	WargaTempatLahir  string    `json:"warga_tempat_lahir"`
	WargaTanggalLahir time.Time `json:"warga_tanggal_lahir"`
	WargaJenisKelamin string    `json:"warga_jenis_kelamin"`
	WargaStatusAktif  string    `json:"warga_status_aktif"`
	WargaStatusHidup  string    `json:"warga_status_hidup"`
	AgamaID           uint      `json:"agama_id"`
	PekerjaanID       uint      `json:"pekerjaan_id"`
}

// ✅ Security validation functions
func isValidNIK(nik string) bool {
	// NIK harus 16 digit angka
	if len(nik) != 16 {
		return false
	}
	matched, _ := regexp.MatchString("^[0-9]{16}$", nik)
	return matched
}

func isValidPhoneNumber(phone string) bool {
	if phone == "" {
		return true // Optional field
	}
	// Format nomor telepon Indonesia
	matched, _ := regexp.MatchString(`^(\+62|62|0)8[1-9][0-9]{6,9}$`, phone)
	return matched
}

func isValidName(name string) bool {
	// Nama harus 2-100 karakter, hanya huruf dan spasi
	if len(name) < 2 || len(name) > 100 {
		return false
	}
	matched, _ := regexp.MatchString("^[a-zA-Z\\s]+$", name)
	return matched
}

func isValidGender(gender string) bool {
	validGenders := map[string]bool{
		"L": true,
		"P": true,
	}
	return validGenders[gender]
}

func isValidStatusAktif(status string) bool {
	validStatuses := map[string]bool{
		"aktif":    true,
		"nonaktif": true,
	}
	return validStatuses[status]
}

func isValidStatusHidup(status string) bool {
	validStatuses := map[string]bool{
		"hidup":      true,
		"meninggal":  true,
	}
	return validStatuses[status]
}

func isValidWargaID(id string) bool {
	// Validasi ID adalah angka positif
	parsedID, err := strconv.ParseUint(id, 10, 32)
	return err == nil && parsedID > 0
}

func sanitizeString(input string) string {
	// Remove potentially dangerous characters
	reg := regexp.MustCompile(`[<>"'%;()&+*|=/\\]`)
	sanitized := reg.ReplaceAllString(input, "")
	return strings.TrimSpace(sanitized)
}

// GetAllWarga returns all warga dengan security checks
func (wc *WargaController) GetAllWarga(c *gin.Context) {
	var wargas []models.Warga
	
	log.Println("🔄 Fetching all residents from database...")
	
	// ✅ SAFE: GORM menggunakan parameterized queries
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumahs").
		Find(&wargas).Error; err != nil {
		log.Printf("❌ Error fetching residents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch residents",
		})
		return
	}

	log.Printf("✅ Successfully fetched %d residents", len(wargas))
	c.JSON(http.StatusOK, gin.H{
		"data":  wargas,
		"count": len(wargas),
	})
}

// GetTotalWarga returns total number of warga
func (wc *WargaController) GetTotalWarga(c *gin.Context) {
	var total int64
	
	log.Println("🔄 Fetching total residents count...")
	
	// ✅ SAFE: Parameterized query
	if err := wc.db.Model(&models.Warga{}).Count(&total).Error; err != nil {
		log.Printf("❌ Error fetching total residents: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch total residents",
		})
		return
	}

	log.Printf("✅ Total residents: %d", total)
	c.JSON(http.StatusOK, gin.H{
		"total_warga": total,
		"message":     "Total residents retrieved successfully",
	})
}

// GetWargaByID returns warga by ID dengan security validation
func (wc *WargaController) GetWargaByID(c *gin.Context) {
	wargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Fetching resident with ID: %s", wargaID)

	var warga models.Warga
	// ✅ SAFE: GORM menggunakan prepared statements
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumahs").
		First(&warga, wargaID).Error; err != nil {
		log.Printf("❌ Error fetching resident %s: %v", wargaID, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resident not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch resident",
			})
		}
		return
	}

	log.Printf("✅ Successfully fetched resident: %s", warga.WargaNama)
	c.JSON(http.StatusOK, warga)
}

// GetWargaByKeluarga returns warga by keluarga ID
func (wc *WargaController) GetWargaByKeluarga(c *gin.Context) {
	keluargaID := c.Param("keluarga_id")

	// ✅ Validasi ID input
	if !isValidWargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Fetching residents for family ID: %s", keluargaID)

	var wargas []models.Warga
	// ✅ SAFE: Parameterized query
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Where("keluarga_id = ?", keluargaID).
		Find(&wargas).Error; err != nil {
		log.Printf("❌ Error fetching residents for family %s: %v", keluargaID, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch residents for this family",
		})
		return
	}

	log.Printf("✅ Successfully fetched %d residents for family ID: %s", len(wargas), keluargaID)
	c.JSON(http.StatusOK, gin.H{
		"data":  wargas,
		"count": len(wargas),
	})
}

// CreateWarga creates new warga dengan comprehensive security checks
func (wc *WargaController) CreateWarga(c *gin.Context) {
	var req CreateWargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	req.WargaNama = sanitizeString(req.WargaNama)
	req.WargaNIK = strings.TrimSpace(req.WargaNIK)
	req.WargaNoTlp = strings.TrimSpace(req.WargaNoTlp)
	req.WargaTempatLahir = sanitizeString(req.WargaTempatLahir)

	// ✅ Validasi nama
	if !isValidName(req.WargaNama) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name must be 2-100 characters and contain only letters and spaces",
		})
		return
	}

	// ✅ Validasi NIK
	if !isValidNIK(req.WargaNIK) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "NIK must be exactly 16 digits",
		})
		return
	}

	// ✅ Validasi nomor telepon
	if !isValidPhoneNumber(req.WargaNoTlp) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid phone number format",
		})
		return
	}

	// ✅ Validasi jenis kelamin
	if !isValidGender(req.WargaJenisKelamin) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gender must be 'L' or 'P'",
		})
		return
	}

	// Set default values jika tidak provided
	if req.WargaStatusAktif == "" {
		req.WargaStatusAktif = "aktif"
	}
	if req.WargaStatusHidup == "" {
		req.WargaStatusHidup = "hidup"
	}

	// ✅ Validasi status
	if !isValidStatusAktif(req.WargaStatusAktif) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status aktif must be 'aktif' or 'nonaktif'",
		})
		return
	}
	if !isValidStatusHidup(req.WargaStatusHidup) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status hidup must be 'hidup' or 'meninggal'",
		})
		return
	}

	// ✅ Check if keluarga exists
	var keluarga models.Keluarga
	if err := wc.db.First(&keluarga, req.KeluargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Family not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate family",
			})
		}
		return
	}

	// ✅ Check if agama exists (jika provided)
	if req.AgamaID != 0 {
		var agama models.Agama
		if err := wc.db.First(&agama, req.AgamaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Religion not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate religion",
				})
			}
			return
		}
	}

	// ✅ Check if pekerjaan exists (jika provided)
	if req.PekerjaanID != 0 {
		var pekerjaan models.Pekerjaan
		if err := wc.db.First(&pekerjaan, req.PekerjaanID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Occupation not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate occupation",
				})
			}
			return
		}
	}

	// ✅ Check if NIK already exists
	var existingWarga models.Warga
	if err := wc.db.Where("warga_nik = ?", req.WargaNIK).First(&existingWarga).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "NIK already exists",
		})
		return
	}

	log.Printf("🔄 Creating new resident: %s", req.WargaNama)

	warga := models.Warga{
		KeluargaID:        req.KeluargaID,
		WargaNama:         req.WargaNama,
		WargaNIK:          req.WargaNIK,
		WargaNoTlp:        req.WargaNoTlp,
		WargaTempatLahir:  req.WargaTempatLahir,
		WargaTanggalLahir: req.WargaTanggalLahir,
		WargaJenisKelamin: req.WargaJenisKelamin,
		WargaStatusAktif:  req.WargaStatusAktif,
		WargaStatusHidup:  req.WargaStatusHidup,
		AgamaID:           req.AgamaID,
		PekerjaanID:       req.PekerjaanID,
	}

	// ✅ SAFE: GORM create dengan parameterized queries
	if err := wc.db.Create(&warga).Error; err != nil {
		log.Printf("❌ Error creating resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create resident",
		})
		return
	}

	// ✅ Reload warga dengan data terbaru
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		First(&warga, warga.WargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load created resident",
		})
		return
	}

	log.Printf("✅ Successfully created resident: %s (ID: %d)", warga.WargaNama, warga.WargaID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Resident created successfully",
		"data":    warga,
	})
}

// UpdateWarga updates warga data dengan security checks
func (wc *WargaController) UpdateWarga(c *gin.Context) {
	wargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Updating resident with ID: %s", wargaID)

	var req UpdateWargaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	if req.WargaNama != "" {
		req.WargaNama = sanitizeString(req.WargaNama)
	}
	if req.WargaNIK != "" {
		req.WargaNIK = strings.TrimSpace(req.WargaNIK)
	}
	if req.WargaNoTlp != "" {
		req.WargaNoTlp = strings.TrimSpace(req.WargaNoTlp)
	}
	if req.WargaTempatLahir != "" {
		req.WargaTempatLahir = sanitizeString(req.WargaTempatLahir)
	}

	var warga models.Warga
	// ✅ SAFE: GORM First dengan parameterized query
	if err := wc.db.First(&warga, wargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resident not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch resident",
			})
		}
		return
	}

	// ✅ Update fields dengan validasi
	if req.KeluargaID != 0 {
		var keluarga models.Keluarga
		if err := wc.db.First(&keluarga, req.KeluargaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Family not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate family",
				})
			}
			return
		}
		warga.KeluargaID = req.KeluargaID
	}

	if req.WargaNama != "" {
		if !isValidName(req.WargaNama) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Name must be 2-100 characters and contain only letters and spaces",
			})
			return
		}
		warga.WargaNama = req.WargaNama
	}

	if req.WargaNIK != "" {
		if !isValidNIK(req.WargaNIK) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "NIK must be exactly 16 digits",
			})
			return
		}
		// Check if NIK already exists (excluding current resident)
		var existingWarga models.Warga
		if err := wc.db.Where("warga_nik = ? AND warga_id != ?", req.WargaNIK, wargaID).First(&existingWarga).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "NIK already exists",
			})
			return
		}
		warga.WargaNIK = req.WargaNIK
	}

	if req.WargaNoTlp != "" {
		if !isValidPhoneNumber(req.WargaNoTlp) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid phone number format",
			})
			return
		}
		warga.WargaNoTlp = req.WargaNoTlp
	}

	if req.WargaTempatLahir != "" {
		warga.WargaTempatLahir = req.WargaTempatLahir
	}

	if !req.WargaTanggalLahir.IsZero() {
		warga.WargaTanggalLahir = req.WargaTanggalLahir
	}

	if req.WargaJenisKelamin != "" {
		if !isValidGender(req.WargaJenisKelamin) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Gender must be 'L' or 'P'",
			})
			return
		}
		warga.WargaJenisKelamin = req.WargaJenisKelamin
	}

	if req.WargaStatusAktif != "" {
		if !isValidStatusAktif(req.WargaStatusAktif) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status aktif must be 'aktif' or 'nonaktif'",
			})
			return
		}
		warga.WargaStatusAktif = req.WargaStatusAktif
	}

	if req.WargaStatusHidup != "" {
		if !isValidStatusHidup(req.WargaStatusHidup) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status hidup must be 'hidup' or 'meninggal'",
			})
			return
		}
		warga.WargaStatusHidup = req.WargaStatusHidup
	}

	if req.AgamaID != 0 {
		var agama models.Agama
		if err := wc.db.First(&agama, req.AgamaID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Religion not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate religion",
				})
			}
			return
		}
		warga.AgamaID = req.AgamaID
	}

	if req.PekerjaanID != 0 {
		var pekerjaan models.Pekerjaan
		if err := wc.db.First(&pekerjaan, req.PekerjaanID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Occupation not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate occupation",
				})
			}
			return
		}
		warga.PekerjaanID = req.PekerjaanID
	}

	// ✅ SAFE: GORM Save dengan parameterized queries
	if err := wc.db.Save(&warga).Error; err != nil {
		log.Printf("❌ Error updating resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update resident",
		})
		return
	}

	// ✅ Reload dengan data terbaru
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumahs").
		First(&warga, wargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated resident",
		})
		return
	}

	log.Printf("✅ Successfully updated resident: %s", warga.WargaNama)
	c.JSON(http.StatusOK, gin.H{
		"message": "Resident updated successfully",
		"data":    warga,
	})
}

// DeleteWarga deletes warga dengan security checks
func (wc *WargaController) DeleteWarga(c *gin.Context) {
	wargaID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Deleting resident with ID: %s", wargaID)

	var warga models.Warga
	// ✅ SAFE: GORM First dengan parameterized query
	if err := wc.db.First(&warga, wargaID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Resident not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch resident",
			})
		}
		return
	}

	// ✅ Check jika warga masih memiliki rumah
	var rumahCount int64
	if err := wc.db.Model(&models.Rumah{}).Where("warga_id = ?", wargaID).Count(&rumahCount).Error; err != nil {
		log.Printf("❌ Error checking resident's houses: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check resident's houses",
		})
		return
	}

	if rumahCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete resident that still has houses",
			"house_count": rumahCount,
		})
		return
	}

	// ✅ SAFE: GORM Delete dengan parameterized query
	if err := wc.db.Delete(&warga).Error; err != nil {
		log.Printf("❌ Error deleting resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete resident",
		})
		return
	}

	log.Printf("✅ Successfully deleted resident: %s (ID: %d)", warga.WargaNama, warga.WargaID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Resident deleted successfully",
		"deleted_warga_id": warga.WargaID,
	})
}

// GetWargaStats returns statistics about warga
func (wc *WargaController) GetWargaStats(c *gin.Context) {
	var stats struct {
		TotalWarga       int64 `json:"total_warga"`
		WargaAktif       int64 `json:"warga_aktif"`
		WargaNonaktif    int64 `json:"warga_nonaktif"`
		WargaLakiLaki    int64 `json:"warga_laki_laki"`
		WargaPerempuan   int64 `json:"warga_perempuan"`
		WargaHidup       int64 `json:"warga_hidup"`
		WargaMeninggal   int64 `json:"warga_meninggal"`
	}

	// ✅ SAFE: Semua count queries menggunakan parameterized queries
	wc.db.Model(&models.Warga{}).Count(&stats.TotalWarga)
	wc.db.Model(&models.Warga{}).Where("warga_status_aktif = ?", "aktif").Count(&stats.WargaAktif)
	wc.db.Model(&models.Warga{}).Where("warga_status_aktif = ?", "nonaktif").Count(&stats.WargaNonaktif)
	wc.db.Model(&models.Warga{}).Where("warga_jenis_kelamin = ?", "L").Count(&stats.WargaLakiLaki)
	wc.db.Model(&models.Warga{}).Where("warga_jenis_kelamin = ?", "P").Count(&stats.WargaPerempuan)
	wc.db.Model(&models.Warga{}).Where("warga_status_hidup = ?", "hidup").Count(&stats.WargaHidup)
	wc.db.Model(&models.Warga{}).Where("warga_status_hidup = ?", "meninggal").Count(&stats.WargaMeninggal)

	log.Printf("📊 Resident stats: Total=%d, Active=%d, Male=%d, Female=%d", 
		stats.TotalWarga, stats.WargaAktif, stats.WargaLakiLaki, stats.WargaPerempuan)

	c.JSON(http.StatusOK, stats)
}

// SearchWarga searches warga by name or NIK
func (wc *WargaController) SearchWarga(c *gin.Context) {
    query := c.Query("q")
    if query == "" {
        c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
        return
    }

    // ✅ Sanitize search query
    query = sanitizeString(query)
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

    log.Printf("🔍 Searching residents with query: %s", query)

    var wargas []models.Warga
    
    // ✅ SAFE: Gunakan parameterized query
    if err := wc.db.
        Preload("Keluarga").
        Preload("Agama").
        Preload("Pekerjaan").
        Where("warga_nama LIKE ? OR warga_nik LIKE ?", "%"+query+"%", "%"+query+"%").
        Find(&wargas).Error; err != nil {
        log.Printf("❌ Error searching residents: %v", err)
        c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search residents",
		})
        return
    }

    log.Printf("✅ Search completed: found %d residents for query '%s'", len(wargas), query)

    c.JSON(http.StatusOK, gin.H{
        "query":   query,
        "results": wargas,
        "count":   len(wargas),
    })
}