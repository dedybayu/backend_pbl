package controllers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
	"rt-management/models"
	"rt-management/helper" // Import helper
	"log"
	"golang.org/x/crypto/bcrypt"

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
	// Warga Data
	KeluargaID        uint   `form:"keluarga_id" binding:"required" example:"1"`
	RumahID           uint   `form:"rumah_id" example:"1"` // ✅ TAMBAHKAN RumahID
	WargaNama         string `form:"warga_nama" binding:"required" example:"John Doe"`
	WargaNIK          string `form:"warga_nik" binding:"required" example:"3201010101010001"`
	NoTelp            string `form:"no_telp" binding:"required" example:"081234567890"` // ✅ GABUNGAN: warga_no_tlp + user_no_telp
	WargaTempatLahir  string `form:"warga_tempat_lahir" example:"Jakarta"`
	WargaTanggalLahir string `form:"warga_tanggal_lahir" example:"1990-01-01"`
	WargaJenisKelamin string `form:"warga_jenis_kelamin" binding:"required" example:"L"`
	WargaStatusAktif  string `form:"warga_status_aktif" example:"aktif"`
	WargaStatusHidup  string `form:"warga_status_hidup" example:"hidup"`
	AgamaID           uint   `form:"agama_id" example:"1"`
	PekerjaanID       uint   `form:"pekerjaan_id" example:"1"`
	
	// User Data (untuk create user otomatis)
	Username    string `form:"username" binding:"required" example:"johndoe"`
	Password    string `form:"password" binding:"required" example:"password123"`
	UserAlamat  string `form:"user_alamat" binding:"required" example:"Jl. Contoh No. 123"`
	UserEmail   string `form:"user_email" binding:"required" example:"johndoe@example.com"`
	// FotoProfile dihapus dari form binding karena akan dihandle sebagai file
}

type UpdateWargaRequest struct {
	// Warga Data
	KeluargaID        uint   `form:"keluarga_id" example:"1"`
	RumahID           uint   `form:"rumah_id" example:"1"` // ✅ TAMBAHKAN RumahID
	WargaNama         string `form:"warga_nama" example:"John Doe Updated"`
	WargaNIK          string `form:"warga_nik" example:"3201010101010001"`
	NoTelp            string `form:"no_telp" example:"081234567890"` // ✅ GABUNGAN: warga_no_tlp + user_no_telp
	WargaTempatLahir  string `form:"warga_tempat_lahir" example:"Jakarta"`
	WargaTanggalLahir string `form:"warga_tanggal_lahir" example:"1990-01-01"`
	WargaJenisKelamin string `form:"warga_jenis_kelamin" example:"L"`
	WargaStatusAktif  string `form:"warga_status_aktif" example:"aktif"`
	WargaStatusHidup  string `form:"warga_status_hidup" example:"hidup"`
	AgamaID           uint   `form:"agama_id" example:"1"`
	PekerjaanID       uint   `form:"pekerjaan_id" example:"1"`
	
	// User Data (opsional untuk update)
	Username    string `form:"username" example:"johndoe_updated"`
	Password    string `form:"password" example:"newpassword123"` // Opsional
	UserAlamat  string `form:"user_alamat" example:"Jl. Contoh No. 456"`
	UserEmail   string `form:"user_email" example:"johndoe.updated@example.com"`
	// FotoProfile dihapus dari form binding karena akan dihandle sebagai file
}

// ✅ Security validation functions (tetap sama)
func isValidNIK(nik string) bool {
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
	matched, _ := regexp.MatchString(`^(\+62|62|0)8[1-9][0-9]{6,9}$`, phone)
	return matched
}

func isValidName(name string) bool {
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
		"hidup":     true,
		"meninggal": true,
	}
	return validStatuses[status]
}

func isValidWargaID(id string) bool {
	parsedID, err := strconv.ParseUint(id, 10, 32)
	return err == nil && parsedID > 0
}

func isValidEmailWarga(email string) bool {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func sanitizeString(input string) string {
	reg := regexp.MustCompile(`[<>"'%;()&+*|=/\\]`)
	sanitized := reg.ReplaceAllString(input, "")
	return strings.TrimSpace(sanitized)
}

func hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedPassword), nil
}

// GetAllWarga godoc
// @Summary Get all residents
// @Description Get a list of all residents with their details
// @Tags warga
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "List of residents"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga [get]
func (wc *WargaController) GetAllWarga(c *gin.Context) {
	var wargas []models.Warga
	
	log.Println("🔄 Fetching all residents from database...")
	
	if err := wc.db.
		Preload("User").
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
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

// GetTotalWarga godoc
// @Summary Get total number of residents
// @Description Get the total count of residents
// @Tags warga
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Total residents count"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/total [get]
func (wc *WargaController) GetTotalWarga(c *gin.Context) {
	var total int64
	
	log.Println("🔄 Fetching total residents count...")
	
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

// GetWargaByID godoc
// @Summary Get resident by ID
// @Description Get resident details by ID
// @Tags warga
// @Accept json
// @Produce json
// @Param id path string true "Resident ID"
// @Success 200 {object} models.Warga "Resident details"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 404 {object} map[string]string "Resident not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/{id} [get]
func (wc *WargaController) GetWargaByID(c *gin.Context) {
	wargaID := c.Param("id")

	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Fetching resident with ID: %s", wargaID)

	var warga models.Warga

	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
		Preload("TagihanIurans").
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

// GetWargaByKeluarga godoc
// @Summary Get residents by family ID
// @Description Get all residents belonging to a specific family
// @Tags warga
// @Accept json
// @Produce json
// @Param keluarga_id path string true "Family ID"
// @Success 200 {object} map[string]interface{} "List of residents in family"
// @Failure 400 {object} map[string]string "Invalid family ID format"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/keluarga/{keluarga_id} [get]
func (wc *WargaController) GetWargaByKeluarga(c *gin.Context) {
	keluargaID := c.Param("keluarga_id")

	if !isValidWargaID(keluargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid family ID format",
		})
		return
	}

	log.Printf("🔄 Fetching residents for family ID: %s", keluargaID)

	var wargas []models.Warga
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
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

// CreateWarga godoc
// @Summary Create new resident with user account
// @Description Create a new resident and automatically create a user account for them
// @Tags warga
// @Accept multipart/form-data
// @Produce json
// @Param keluarga_id formData integer true "Family ID"
// @Param rumah_id formData integer false "House ID"
// @Param warga_nama formData string true "Resident name"
// @Param warga_nik formData string true "Resident NIK (16 digits)"
// @Param no_telp formData string true "Phone number (used for both resident and user)"
// @Param warga_tempat_lahir formData string false "Place of birth"
// @Param warga_tanggal_lahir formData string false "Date of birth (YYYY-MM-DD)"
// @Param warga_jenis_kelamin formData string true "Gender (L/P)"
// @Param warga_status_aktif formData string false "Active status (aktif/nonaktif)"
// @Param warga_status_hidup formData string false "Life status (hidup/meninggal)"
// @Param agama_id formData integer false "Religion ID"
// @Param pekerjaan_id formData integer false "Occupation ID"
// @Param username formData string true "Username for user account"
// @Param password formData string true "Password for user account"
// @Param user_alamat formData string true "User address"
// @Param user_email formData string true "User email"
// @Param foto_profile formData file false "Profile photo file"
// @Success 201 {object} map[string]interface{} "Resident created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 409 {object} map[string]string "Conflict (NIK or username already exists)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga [post]
func (wc *WargaController) CreateWarga(c *gin.Context) {
	var req CreateWargaRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	req.WargaNama = sanitizeString(req.WargaNama)
	req.WargaNIK = strings.TrimSpace(req.WargaNIK)
	req.NoTelp = strings.TrimSpace(req.NoTelp)
	req.WargaTempatLahir = sanitizeString(req.WargaTempatLahir)
	req.WargaTanggalLahir = strings.TrimSpace(req.WargaTanggalLahir)
	req.Username = sanitizeString(strings.TrimSpace(req.Username))
	req.UserAlamat = sanitizeString(req.UserAlamat)
	req.UserEmail = strings.TrimSpace(req.UserEmail)

	// ✅ Validasi data warga
	if !isValidName(req.WargaNama) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Name must be 2-100 characters and contain only letters and spaces",
		})
		return
	}

	if !isValidNIK(req.WargaNIK) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "NIK must be exactly 16 digits",
		})
		return
	}

	// ✅ Validasi nomor telepon (digunakan untuk warga dan user)
	if !isValidPhoneNumber(req.NoTelp) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid phone number format",
		})
		return
	}

	if !isValidGender(req.WargaJenisKelamin) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gender must be 'L' or 'P'",
		})
		return
	}

	// ✅ Validasi data user
	if len(req.Username) < 3 || len(req.Username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username must be 3-50 characters",
		})
		return
	}

	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 6 characters",
		})
		return
	}

	if !isValidEmailWarga(req.UserEmail) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid email format",
		})
		return
	}

	// ✅ Parsing tanggal lahir
	wargaTanggalLahir := time.Time{}
	if req.WargaTanggalLahir != "" {
		parsedDate, err := time.Parse("2006-01-02", req.WargaTanggalLahir)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}
		if parsedDate.After(time.Now()) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Date of birth cannot be in the future",
			})
			return
		}
		wargaTanggalLahir = parsedDate
	}

	// Set default values
	if req.WargaStatusAktif == "" {
		req.WargaStatusAktif = "aktif"
	}
	if req.WargaStatusHidup == "" {
		req.WargaStatusHidup = "hidup"
	}

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

	// ✅ Cek duplikasi NIK
	var existingWargaWithNIK models.Warga
	if err := wc.db.Where("warga_nik = ?", req.WargaNIK).First(&existingWargaWithNIK).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "NIK already exists",
		})
		return
	}

	// ✅ Cek duplikasi username
	var existingUser models.User
	if err := wc.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// ✅ Cek duplikasi email
	if err := wc.db.Where("user_email = ?", req.UserEmail).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Email already exists",
		})
		return
	}

	// ✅ Validasi referensi
	// Check keluarga exists
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

	// Check agama exists (jika provided)
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

	// Check pekerjaan exists (jika provided)
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

	// Check rumah exists (jika provided)
	if req.RumahID != 0 {
		var rumah models.Rumah
		if err := wc.db.First(&rumah, req.RumahID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "House not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate house",
				})
			}
			return
		}
	}

	// ✅ Mulai transaction
	tx := wc.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction failed",
			})
		}
	}()

	// ✅ Handle file upload untuk foto_profile
	fotoProfileFilename := ""
	fotoProfile, err := helper.HandleFileImageUpload(c, "foto_profile", "")
	if err != nil && err.Error() != "http: no such file" {
		tx.Rollback()
		log.Printf("❌ Error uploading profile photo: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to upload profile photo: " + err.Error(),
		})
		return
	} else if err == nil {
		fotoProfileFilename = fotoProfile
	}

	// ✅ Hash password
	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	// ✅ 1. Buat user terlebih dahulu
	// LevelID 6 diasumsikan sebagai level "Warga" (sesuaikan dengan database Anda)
	newUser := models.User{
		Username:    req.Username,
		Password:    hashedPassword,
		UserNama:    req.WargaNama, // Gunakan nama warga sebagai nama user
		UserAlamat:  req.UserAlamat,
		UserNoTelp:  req.NoTelp, // ✅ Gunakan nomor telepon yang sama
		UserEmail:   req.UserEmail,
		LevelID:     6, // Level untuk warga, sesuaikan dengan database
		FotoProfile: fotoProfileFilename,
	}

	if err := tx.Create(&newUser).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Error creating user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create user account",
		})
		return
	}

	log.Printf("✅ User created: %s (ID: %d)", newUser.Username, newUser.UserID)

	// ✅ 2. Buat warga dengan user_id yang sudah dibuat
	warga := models.Warga{
		UserID:            newUser.UserID,
		KeluargaID:        req.KeluargaID,
		RumahID:           req.RumahID,
		WargaNama:         req.WargaNama,
		WargaNIK:          req.WargaNIK,
		WargaNoTlp:        req.NoTelp, // ✅ Gunakan nomor telepon yang sama
		WargaTempatLahir:  req.WargaTempatLahir,
		WargaTanggalLahir: wargaTanggalLahir,
		WargaJenisKelamin: req.WargaJenisKelamin,
		WargaStatusAktif:  req.WargaStatusAktif,
		WargaStatusHidup:  req.WargaStatusHidup,
		AgamaID:           req.AgamaID,
		PekerjaanID:       req.PekerjaanID,
	}

	if err := tx.Create(&warga).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Error creating resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create resident",
		})
		return
	}

	// ✅ 3. Update status rumah jika warga ditempatkan di rumah
	if req.RumahID != 0 {
		if err := tx.Model(&models.Rumah{}).
			Where("rumah_id = ?", req.RumahID).
			Update("rumah_status", "ditempati").Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Error updating house status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update house status",
			})
			return
		}
	}

	// ✅ Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save data",
		})
		return
	}

	// ✅ Reload warga dengan data terbaru
	if err := wc.db.
		Preload("User").
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
		First(&warga, warga.WargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load created resident",
		})
		return
	}

	log.Printf("✅ Successfully created resident with user account: %s (WargaID: %d, UserID: %d)", 
		warga.WargaNama, warga.WargaID, warga.UserID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Resident and user account created successfully",
		"data":    warga,
		"user": gin.H{
			"user_id":   newUser.UserID,
			"username":  newUser.Username,
			"email":     newUser.UserEmail,
		},
	})
}

// UpdateWarga godoc
// @Summary Update resident and user account
// @Description Update resident information and optionally update associated user account
// @Tags warga
// @Accept multipart/form-data
// @Produce json
// @Param id path string true "Resident ID"
// @Param keluarga_id formData integer false "Family ID"
// @Param rumah_id formData integer false "House ID"
// @Param warga_nama formData string false "Resident name"
// @Param warga_nik formData string false "Resident NIK (16 digits)"
// @Param no_telp formData string false "Phone number (used for both resident and user)"
// @Param warga_tempat_lahir formData string false "Place of birth"
// @Param warga_tanggal_lahir formData string false "Date of birth (YYYY-MM-DD)"
// @Param warga_jenis_kelamin formData string false "Gender (L/P)"
// @Param warga_status_aktif formData string false "Active status (aktif/nonaktif)"
// @Param warga_status_hidup formData string false "Life status (hidup/meninggal)"
// @Param agama_id formData integer false "Religion ID"
// @Param pekerjaan_id formData integer false "Occupation ID"
// @Param username formData string false "Username for user account"
// @Param password formData string false "Password for user account (optional)"
// @Param user_alamat formData string false "User address"
// @Param user_email formData string false "User email"
// @Param foto_profile formData file false "Profile photo file"
// @Success 200 {object} map[string]interface{} "Resident updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "Resident not found"
// @Failure 409 {object} map[string]string "Conflict (NIK or username already exists)"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/{id} [put]
func (wc *WargaController) UpdateWarga(c *gin.Context) {
	wargaID := c.Param("id")

	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Updating resident with ID: %s", wargaID)

	var req UpdateWargaRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
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
	if req.NoTelp != "" {
		req.NoTelp = strings.TrimSpace(req.NoTelp)
	}
	if req.WargaTempatLahir != "" {
		req.WargaTempatLahir = sanitizeString(req.WargaTempatLahir)
	}
	if req.WargaTanggalLahir != "" {
		req.WargaTanggalLahir = strings.TrimSpace(req.WargaTanggalLahir)
	}
	if req.Username != "" {
		req.Username = sanitizeString(strings.TrimSpace(req.Username))
	}
	if req.UserAlamat != "" {
		req.UserAlamat = sanitizeString(req.UserAlamat)
	}
	if req.UserEmail != "" {
		req.UserEmail = strings.TrimSpace(req.UserEmail)
	}

	// ✅ Get existing warga
	var warga models.Warga
	if err := wc.db.Preload("User").First(&warga, wargaID).Error; err != nil {
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

	// ✅ Simpan rumah_id sebelumnya untuk update status rumah
	previousRumahID := warga.RumahID

	// ✅ Mulai transaction
	tx := wc.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction failed",
			})
		}
	}()

	// ✅ Handle file upload untuk foto_profile
	oldFotoProfile := ""
	if warga.User != nil {
		oldFotoProfile = warga.User.FotoProfile
	}
	
	fotoProfile, err := helper.HandleFileImageUpload(c, "foto_profile", oldFotoProfile)
	if err != nil && err.Error() != "http: no such file" {
		tx.Rollback()
		log.Printf("❌ Error uploading profile photo: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to upload profile photo: " + err.Error(),
		})
		return
	}

	// ✅ Update warga data
	if req.KeluargaID != 0 {
		var keluarga models.Keluarga
		if err := tx.First(&keluarga, req.KeluargaID).Error; err != nil {
			tx.Rollback()
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

	// ✅ Update RumahID dengan validasi
	if req.RumahID != 0 && req.RumahID != warga.RumahID {
		var rumah models.Rumah
		if err := tx.First(&rumah, req.RumahID).Error; err != nil {
			tx.Rollback()
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "House not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate house",
				})
			}
			return
		}
		warga.RumahID = req.RumahID
	}

	if req.WargaNama != "" {
		if !isValidName(req.WargaNama) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Name must be 2-100 characters and contain only letters and spaces",
			})
			return
		}
		warga.WargaNama = req.WargaNama
	}

	if req.WargaNIK != "" {
		if !isValidNIK(req.WargaNIK) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "NIK must be exactly 16 digits",
			})
			return
		}
		// Check jika NIK already exists (excluding current resident)
		var existingWarga models.Warga
		if err := tx.Where("warga_nik = ? AND warga_id != ?", req.WargaNIK, wargaID).First(&existingWarga).Error; err == nil {
			tx.Rollback()
			c.JSON(http.StatusConflict, gin.H{
				"error": "NIK already exists",
			})
			return
		}
		warga.WargaNIK = req.WargaNIK
	}

	if req.NoTelp != "" {
		if !isValidPhoneNumber(req.NoTelp) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid phone number format",
			})
			return
		}
		warga.WargaNoTlp = req.NoTelp
	}

	if req.WargaTempatLahir != "" {
		warga.WargaTempatLahir = req.WargaTempatLahir
	}

	if req.WargaTanggalLahir != "" {
		wargaTanggalLahir, err := time.Parse("2006-01-02", req.WargaTanggalLahir)
		if err != nil {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid date format. Use YYYY-MM-DD format",
			})
			return
		}
		if wargaTanggalLahir.After(time.Now()) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Date of birth cannot be in the future",
			})
			return
		}
		warga.WargaTanggalLahir = wargaTanggalLahir
	}

	if req.WargaJenisKelamin != "" {
		if !isValidGender(req.WargaJenisKelamin) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Gender must be 'L' or 'P'",
			})
			return
		}
		warga.WargaJenisKelamin = req.WargaJenisKelamin
	}

	if req.WargaStatusAktif != "" {
		if !isValidStatusAktif(req.WargaStatusAktif) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status aktif must be 'aktif' or 'nonaktif'",
			})
			return
		}
		warga.WargaStatusAktif = req.WargaStatusAktif
	}

	if req.WargaStatusHidup != "" {
		if !isValidStatusHidup(req.WargaStatusHidup) {
			tx.Rollback()
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status hidup must be 'hidup' atau 'meninggal'",
			})
			return
		}
		warga.WargaStatusHidup = req.WargaStatusHidup
	}

	if req.AgamaID != 0 {
		var agama models.Agama
		if err := tx.First(&agama, req.AgamaID).Error; err != nil {
			tx.Rollback()
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
		if err := tx.First(&pekerjaan, req.PekerjaanID).Error; err != nil {
			tx.Rollback()
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

	// ✅ Update warga
	if err := tx.Save(&warga).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Error updating resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update resident",
		})
		return
	}

	// ✅ Update user data jika ada perubahan
	if warga.User != nil {
		userUpdates := make(map[string]interface{})
		
		if req.Username != "" && req.Username != warga.User.Username {
			// Check jika username sudah digunakan oleh user lain
			var existingUser models.User
			if err := tx.Where("username = ? AND user_id != ?", req.Username, warga.User.UserID).First(&existingUser).Error; err == nil {
				tx.Rollback()
				c.JSON(http.StatusConflict, gin.H{
					"error": "Username already exists",
				})
				return
			}
			userUpdates["username"] = req.Username
		}

		if req.Password != "" {
			hashedPassword, err := hashPassword(req.Password)
			if err != nil {
				tx.Rollback()
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to hash password",
				})
				return
			}
			userUpdates["password"] = hashedPassword
		}

		if req.UserAlamat != "" {
			userUpdates["user_alamat"] = req.UserAlamat
		}

		if req.NoTelp != "" && req.NoTelp != warga.User.UserNoTelp {
			userUpdates["user_no_telp"] = req.NoTelp
		}

		if req.UserEmail != "" && req.UserEmail != warga.User.UserEmail {
			if !isValidEmailWarga(req.UserEmail) {
				tx.Rollback()
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Invalid email format",
				})
				return
			}
			// Check jika email sudah digunakan oleh user lain
			var existingUser models.User
			if err := tx.Where("user_email = ? AND user_id != ?", req.UserEmail, warga.User.UserID).First(&existingUser).Error; err == nil {
				tx.Rollback()
				c.JSON(http.StatusConflict, gin.H{
					"error": "Email already exists",
				})
				return
			}
			userUpdates["user_email"] = req.UserEmail
		}

		if req.WargaNama != "" && req.WargaNama != warga.User.UserNama {
			userUpdates["user_nama"] = req.WargaNama
		}

		// ✅ Update foto_profile jika ada file baru
		if fotoProfile != "" && fotoProfile != oldFotoProfile {
			userUpdates["foto_profile"] = fotoProfile
		}

		// ✅ Update user jika ada perubahan
		if len(userUpdates) > 0 {
			if err := tx.Model(&models.User{}).
				Where("user_id = ?", warga.UserID).
				Updates(userUpdates).Error; err != nil {
				tx.Rollback()
				log.Printf("❌ Error updating user: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to update user account",
				})
				return
			}
		}
	}

	// ✅ Update status rumah berdasarkan perubahan
	if previousRumahID != 0 && previousRumahID != warga.RumahID {
		if err := tx.Model(&models.Rumah{}).
			Where("rumah_id = ?", previousRumahID).
			Update("rumah_status", "tersedia").Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Error updating previous house status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update house status",
			})
			return
		}
	}
	
	if warga.RumahID != 0 {
		if err := tx.Model(&models.Rumah{}).
			Where("rumah_id = ?", warga.RumahID).
			Update("rumah_status", "ditempati").Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Error updating new house status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update house status",
			})
			return
		}
	}

	// ✅ Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to save updates",
		})
		return
	}

	// ✅ Reload dengan data terbaru
	if err := wc.db.
		Preload("User").
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
		First(&warga, wargaID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated resident",
		})
		return
	}

	log.Printf("✅ Successfully updated resident: %s", warga.WargaNama)
	c.JSON(http.StatusOK, gin.H{
		"message": "Resident and user account updated successfully",
		"data":    warga,
	})
}

// DeleteWarga godoc
// @Summary Delete resident and associated user account
// @Description Delete resident and automatically delete the associated user account
// @Tags warga
// @Accept json
// @Produce json
// @Param id path string true "Resident ID"
// @Success 200 {object} map[string]interface{} "Resident and user deleted successfully"
// @Failure 400 {object} map[string]string "Invalid ID format"
// @Failure 404 {object} map[string]string "Resident not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/{id} [delete]
func (wc *WargaController) DeleteWarga(c *gin.Context) {
	wargaID := c.Param("id")

	if !isValidWargaID(wargaID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid resident ID format",
		})
		return
	}

	log.Printf("🔄 Deleting resident with ID: %s", wargaID)

	var warga models.Warga
	if err := wc.db.Preload("User").First(&warga, wargaID).Error; err != nil {
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

	// ✅ Mulai transaction
	tx := wc.db.Begin()
	if tx.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to start transaction",
		})
		return
	}

	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Transaction failed",
			})
		}
	}()

	// ✅ Update rumah menjadi tersedia jika warga punya rumah
	if warga.RumahID != 0 {
		if err := tx.Model(&models.Rumah{}).
			Where("rumah_id = ?", warga.RumahID).
			Update("rumah_status", "tersedia").Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Error updating house status: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update house status",
			})
			return
		}
	}

	// ✅ Hapus foto profile jika ada
	if warga.User != nil && warga.User.FotoProfile != "" {
		if err := helper.DeleteOldPhoto(warga.User.FotoProfile, "foto_profile"); err != nil {
			log.Printf("⚠️ Warning: Failed to delete profile photo: %v", err)
			// Lanjutkan proses delete meskipun gagal hapus foto
		}
	}

	// ✅ Hapus warga
	if err := tx.Delete(&warga).Error; err != nil {
		tx.Rollback()
		log.Printf("❌ Error deleting resident: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete resident",
		})
		return
	}

	// ✅ Hapus user terkait jika ada
	if warga.User != nil {
		if err := tx.Delete(&warga.User).Error; err != nil {
			tx.Rollback()
			log.Printf("❌ Error deleting user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to delete user account",
			})
			return
		}
		log.Printf("✅ User deleted: %s (ID: %d)", warga.User.Username, warga.User.UserID)
	}

	// ✅ Commit transaction
	if err := tx.Commit().Error; err != nil {
		log.Printf("❌ Error committing transaction: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete data",
		})
		return
	}

	log.Printf("✅ Successfully deleted resident and user account: %s (ID: %d)", warga.WargaNama, warga.WargaID)
	c.JSON(http.StatusOK, gin.H{
		"message": "Resident and user account deleted successfully",
		"deleted_warga_id": warga.WargaID,
		"deleted_user_id":  warga.UserID,
	})
}

// GetWargaStats godoc
// @Summary Get resident statistics
// @Description Get various statistics about residents
// @Tags warga
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{} "Resident statistics"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/stats [get]
func (wc *WargaController) GetWargaStats(c *gin.Context) {
	var stats struct {
		TotalWarga        int64 `json:"total_warga"`
		WargaAktif        int64 `json:"warga_aktif"`
		WargaNonaktif     int64 `json:"warga_nonaktif"`
		WargaLakiLaki     int64 `json:"warga_laki_laki"`
		WargaPerempuan    int64 `json:"warga_perempuan"`
		WargaHidup        int64 `json:"warga_hidup"`
		WargaMeninggal    int64 `json:"warga_meninggal"`
		WargaDenganRumah  int64 `json:"warga_dengan_rumah"`
		WargaTanpaRumah   int64 `json:"warga_tanpa_rumah"`
	}

	wc.db.Model(&models.Warga{}).Count(&stats.TotalWarga)
	wc.db.Model(&models.Warga{}).Where("warga_status_aktif = ?", "aktif").Count(&stats.WargaAktif)
	wc.db.Model(&models.Warga{}).Where("warga_status_aktif = ?", "nonaktif").Count(&stats.WargaNonaktif)
	wc.db.Model(&models.Warga{}).Where("warga_jenis_kelamin = ?", "L").Count(&stats.WargaLakiLaki)
	wc.db.Model(&models.Warga{}).Where("warga_jenis_kelamin = ?", "P").Count(&stats.WargaPerempuan)
	wc.db.Model(&models.Warga{}).Where("warga_status_hidup = ?", "hidup").Count(&stats.WargaHidup)
	wc.db.Model(&models.Warga{}).Where("warga_status_hidup = ?", "meninggal").Count(&stats.WargaMeninggal)
	wc.db.Model(&models.Warga{}).Where("rumah_id IS NOT NULL").Count(&stats.WargaDenganRumah)
	wc.db.Model(&models.Warga{}).Where("rumah_id IS NULL").Count(&stats.WargaTanpaRumah)

	log.Printf("📊 Resident stats: Total=%d, Active=%d, WithHouse=%d, WithoutHouse=%d", 
		stats.TotalWarga, stats.WargaAktif, stats.WargaDenganRumah, stats.WargaTanpaRumah)

	c.JSON(http.StatusOK, stats)
}

// SearchWarga godoc
// @Summary Search residents
// @Description Search residents by name or NIK
// @Tags warga
// @Accept json
// @Produce json
// @Param q query string true "Search query"
// @Success 200 {object} map[string]interface{} "Search results"
// @Failure 400 {object} map[string]string "Query parameter is required"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/search [get]
func (wc *WargaController) SearchWarga(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Query parameter 'q' is required",
		})
		return
	}

	query = sanitizeString(query)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid search query",
		})
		return
	}

	if len(query) > 50 {
		query = query[:50]
	}

	log.Printf("🔍 Searching residents with query: %s", query)

	var wargas []models.Warga
    
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
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

// GetWargaByRumah godoc
// @Summary Get resident by house ID
// @Description Get resident details by house ID
// @Tags warga
// @Accept json
// @Produce json
// @Param rumah_id path string true "House ID"
// @Success 200 {object} models.Warga "Resident details"
// @Failure 400 {object} map[string]string "Invalid house ID format"
// @Failure 404 {object} map[string]string "No resident found for this house"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /api/warga/rumah/{rumah_id} [get]
func (wc *WargaController) GetWargaByRumah(c *gin.Context) {
	rumahID := c.Param("rumah_id")

	if !isValidWargaID(rumahID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid house ID format",
		})
		return
	}

	log.Printf("🔄 Fetching resident for house ID: %s", rumahID)

	var warga models.Warga
	if err := wc.db.
		Preload("Keluarga").
		Preload("Agama").
		Preload("Pekerjaan").
		Preload("Rumah").
		Where("rumah_id = ?", rumahID).
		First(&warga).Error; err != nil {
		log.Printf("❌ Error fetching resident for house %s: %v", rumahID, err)
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "No resident found for this house",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch resident for this house",
			})
		}
		return
	}

	log.Printf("✅ Successfully fetched resident for house ID: %s", rumahID)
	c.JSON(http.StatusOK, warga)
}