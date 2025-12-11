// controllers/user_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"rt-management/helper"
	"rt-management/models"
	"strconv"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type UserController struct {
	db *gorm.DB
}

func NewUserController(db *gorm.DB) *UserController {
	return &UserController{db: db}
}

// CreateUserRequest untuk Swagger documentation
// @Description Request untuk membuat user baru
type CreateUserRequest struct {
	Username    string `form:"username" binding:"required" example:"john_doe"`
	Password    string `form:"password" binding:"required" example:"Password123"`
	UserNama    string `form:"user_nama" binding:"required" example:"John Doe"`
	UserAlamat  string `form:"user_alamat" binding:"required" example:"Jl. Contoh No. 123"`
	UserNoTelp  string `form:"user_no_telp" binding:"required" example:"081234567890"`
	UserEmail   string `form:"user_email" binding:"required" example:"john@example.com"`
	LevelID     uint   `form:"level_id" binding:"required" example:"1"`
	FotoProfile string `form:"foto_profile" example:"profile.jpg"`
}

// UpdateUserRequest untuk Swagger documentation
// @Description Request untuk mengupdate user
type UpdateUserRequest struct {
	Username    string `form:"username" example:"john_doe_updated"`
	Password    string `form:"password" example:"NewPassword123"`
	UserNama    string `form:"user_nama" example:"John Doe Updated"`
	UserAlamat  string `form:"user_alamat" example:"Jl. Baru No. 456"`
	UserNoTelp  string `form:"user_no_telp" example:"081298765432"`
	UserEmail   string `form:"user_email" example:"john.updated@example.com"`
	LevelID     uint   `form:"level_id" example:"2"`
	FotoProfile string `form:"foto_profile" example:"profile_updated.jpg"`
}

// UserResponse untuk Swagger documentation
// @Description Response untuk user data
type UserResponse struct {
	UserID      uint           `json:"user_id" example:"1"`
	Username    string         `json:"username" example:"john_doe"`
	UserNama    string         `json:"user_nama" example:"John Doe"`
	UserAlamat  string         `json:"user_alamat" example:"Jl. Contoh No. 123"`
	UserNoTelp  string         `json:"user_no_telp" example:"081234567890"`
	UserEmail   string         `json:"user_email" example:"john@example.com"`
	LevelID     uint           `json:"level_id" example:"1"`
	FotoProfile string         `json:"foto_profile" example:"profile.jpg"`
	CreatedAt   string         `json:"created_at" example:"2024-01-01T00:00:00Z"`
	UpdatedAt   string         `json:"updated_at" example:"2024-01-01T00:00:00Z"`
	Level       models.Level   `json:"level"`
	Warga       *models.Warga  `json:"warga,omitempty"` // hanya tampil jika tidak nil
}

// ✅ Security validation functions
func isValidUsername(username string) bool {
	// Username harus 3-20 karakter, hanya huruf, angka, underscore
	matched, _ := regexp.MatchString("^[a-zA-Z0-9_]{3,20}$", username)
	return matched
}

func isValidPassword(password string) bool {
	// Password minimal 8 karakter, mengandung huruf dan angka
	if len(password) < 8 {
		return false
	}
	hasLetter, _ := regexp.MatchString("[a-zA-Z]", password)
	hasNumber, _ := regexp.MatchString("[0-9]", password)
	return hasLetter && hasNumber
}

func isValidEmail(email string) bool {
	// Validasi email format sederhana
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return emailRegex.MatchString(email)
}

func isValidPhone(phone string) bool {
	// Validasi nomor telepon (minimal 10 digit, maksimal 15 digit, hanya angka)
	phoneRegex := regexp.MustCompile(`^[0-9]{10,15}$`)
	return phoneRegex.MatchString(phone)
}

func sanitizeInput(input string) string {
	// Remove potentially dangerous characters
	reg := regexp.MustCompile(`[<>"'%;()&+]`)
	sanitized := reg.ReplaceAllString(input, "")
	return strings.TrimSpace(sanitized)
}

func isValidUserID(id string) bool {
	// Validasi ID adalah angka positif
	parsedID, err := strconv.ParseUint(id, 10, 32)
	return err == nil && parsedID > 0
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get a list of all users with their level information
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [get]
func (uc *UserController) GetAllUsers(c *gin.Context) {
	var users []models.User

	// ✅ SAFE: GORM menggunakan parameterized queries internally
	if err := uc.db.
		Preload("Level").
		Preload("Warga"). // Preload Warga

		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch users",
		})
		return
	}

	// ✅ Sanitize response data (optional, untuk extra security)
	for i := range users {
		users[i].Password = "" // Jangan expose password hash
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  users,
		"count": len(users),
	})
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get detailed user information by user ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} models.User "Success response"
// @Failure 400 {object} map[string]string "Invalid user ID"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{id} [get]
func (uc *UserController) GetUserByID(c *gin.Context) {
	userID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidUserID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var user models.User
	// ✅ SAFE: GORM menggunakan prepared statements
	if err := uc.db.
		Preload("Level").
		Preload("Warga").

		First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user",
			})
		}
		return
	}

	// ✅ Jangan expose password hash
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with all required fields
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param username formData string true "Username"
// @Param password formData string true "Password"
// @Param user_nama formData string true "Full name"
// @Param user_alamat formData string true "Address"
// @Param user_no_telp formData string true "Phone number"
// @Param user_email formData string true "Email address"
// @Param level_id formData int true "Level ID"
// @Param foto_profile formData file false "Profile photo"
// @Success 201 {object} map[string]interface{} "User created successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users [post]
func (uc *UserController) CreateUser(c *gin.Context) {
	// Binding manual untuk form data
	username := sanitizeInput(strings.TrimSpace(c.PostForm("username")))
	password := strings.TrimSpace(c.PostForm("password"))
	userNama := sanitizeInput(strings.TrimSpace(c.PostForm("user_nama")))
	userAlamat := sanitizeInput(strings.TrimSpace(c.PostForm("user_alamat")))
	userNoTelp := strings.TrimSpace(c.PostForm("user_no_telp"))
	userEmail := strings.TrimSpace(c.PostForm("user_email"))
	levelIDStr := c.PostForm("level_id")

	// Validasi required fields
	if username == "" || password == "" || userNama == "" || userAlamat == "" || userNoTelp == "" || userEmail == "" || levelIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Semua field wajib harus diisi",
		})
		return
	}

	// Convert level_id to uint
	levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Level ID tidak valid",
		})
		return
	}

	// ✅ Validasi username format
	if !isValidUsername(username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username must be 3-20 characters and contain only letters, numbers, and underscores",
		})
		return
	}

	// ✅ Validasi password strength
	if !isValidPassword(password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters and contain both letters and numbers",
		})
		return
	}

	// ✅ Validasi email format
	if !isValidEmail(userEmail) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email format tidak valid",
		})
		return
	}

	// ✅ Validasi nomor telepon
	if !isValidPhone(userNoTelp) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nomor telepon harus 10-15 digit angka",
		})
		return
	}

	// ✅ Validasi level ID
	if levelID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Level ID is required",
		})
		return
	}

	// ✅ Check if level exists - SAFE: parameterized query
	var level models.Level
	if err := uc.db.First(&level, uint(levelID)).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Level not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to validate level",
			})
		}
		return
	}

	// ✅ Check if username already exists
	var existingUser models.User
	if err := uc.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// ✅ Check if email already exists
	if err := uc.db.Where("user_email = ?", userEmail).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Email already registered",
		})
		return
	}

	// ✅ Hash password dengan cost yang aman
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to secure password",
		})
		return
	}

	// Handle file upload untuk foto_profile
	fotoProfileFilename := ""
	if _, header, err := c.Request.FormFile("foto_profile"); err == nil && header != nil {
		// Gunakan helper untuk handle upload
		filename, err := helper.HandleFileImageUpload(c, "foto_profile", "")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto profile",
				"details": err.Error(),
			})
			return
		}
		fotoProfileFilename = filename
	}

	// Buat user dengan semua field
	user := models.User{
		Username:    username,
		Password:    string(hashedPassword),
		UserNama:    userNama,
		UserAlamat:  userAlamat,
		UserNoTelp:  userNoTelp,
		UserEmail:   userEmail,
		LevelID:     uint(levelID),
		FotoProfile: fotoProfileFilename,
	}

	// ✅ SAFE: GORM create dengan parameterized queries
	if err := uc.db.Create(&user).Error; err != nil {
		// Jika gagal create, hapus file yang sudah diupload
		if fotoProfileFilename != "" {
			helper.DeleteOldPhoto(fotoProfileFilename, "profile_foto")
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"details": err.Error(),
		})
		return
	}

	// ✅ Reload user dengan data terbaru
	if err := uc.db.
		Preload("Level").
		Preload("Warga").

		First(&user, user.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load created user",
		})
		return
	}

	// ✅ Jangan expose password hash di response
	user.Password = ""

	c.JSON(http.StatusCreated, gin.H{
		"message": "User created successfully",
		"data":    user,
	})
}

// UpdateUser godoc
// @Summary Update user data
// @Description Update existing user information
// @Tags Users
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param username formData string false "Username"
// @Param password formData string false "Password"
// @Param user_nama formData string false "Full name"
// @Param user_alamat formData string false "Address"
// @Param user_no_telp formData string false "Phone number"
// @Param user_email formData string false "Email address"
// @Param level_id formData int false "Level ID"
// @Param foto_profile formData file false "Profile photo"
// @Success 200 {object} map[string]interface{} "User updated successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{id} [put]
func (uc *UserController) UpdateUser(c *gin.Context) {
	userID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidUserID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var user models.User
	// ✅ SAFE: GORM First dengan parameterized query
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user",
			})
		}
		return
	}

	// Binding manual untuk form data
	username := sanitizeInput(strings.TrimSpace(c.PostForm("username")))
	password := strings.TrimSpace(c.PostForm("password"))
	userNama := sanitizeInput(strings.TrimSpace(c.PostForm("user_nama")))
	userAlamat := sanitizeInput(strings.TrimSpace(c.PostForm("user_alamat")))
	userNoTelp := strings.TrimSpace(c.PostForm("user_no_telp"))
	userEmail := strings.TrimSpace(c.PostForm("user_email"))
	levelIDStr := c.PostForm("level_id")

	// Update fields
	updates := make(map[string]interface{})

	// ✅ Check if level exists if provided
	if levelIDStr != "" {
		levelID, err := strconv.ParseUint(levelIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Level ID tidak valid",
			})
			return
		}

		var level models.Level
		// ✅ SAFE: parameterized query
		if err := uc.db.First(&level, uint(levelID)).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Level not found",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to validate level",
				})
			}
			return
		}
		updates["level_id"] = uint(levelID)
	}

	// ✅ Update username dengan validasi
	if username != "" {
		if !isValidUsername(username) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username must be 3-20 characters and contain only letters, numbers, and underscores",
			})
			return
		}

		// ✅ Check if username already exists (excluding current user)
		var existingUser models.User
		if err := uc.db.Where("username = ? AND user_id != ?", username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username already exists",
			})
			return
		}
		updates["username"] = username
	}

	// ✅ Update nama
	if userNama != "" {
		if len(userNama) < 2 || len(userNama) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama harus 2-100 karakter",
			})
			return
		}
		updates["user_nama"] = userNama
	}

	// ✅ Update alamat
	if userAlamat != "" {
		if len(userAlamat) < 5 || len(userAlamat) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Alamat harus 5-255 karakter",
			})
			return
		}
		updates["user_alamat"] = userAlamat
	}

	// ✅ Update no telp
	if userNoTelp != "" {
		if !isValidPhone(userNoTelp) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nomor telepon harus 10-15 digit angka",
			})
			return
		}
		updates["user_no_telp"] = userNoTelp
	}

	// ✅ Update email
	if userEmail != "" {
		if !isValidEmail(userEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email format tidak valid",
			})
			return
		}

		// ✅ Check if email already exists (excluding current user)
		var existingUser models.User
		if err := uc.db.Where("user_email = ? AND user_id != ?", userEmail, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Email already registered",
			})
			return
		}
		updates["user_email"] = userEmail
	}

	// ✅ Update password dengan validasi
	if password != "" {
		if !isValidPassword(password) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 8 characters and contain both letters and numbers",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to secure password",
			})
			return
		}
		updates["password"] = string(hashedPassword)
	}

	// Handle file upload untuk foto_profile
	fotoProfileFilename := user.FotoProfile // Simpan filename lama dulu
	if _, header, err := c.Request.FormFile("foto_profile"); err == nil && header != nil {
		// Ada file baru yang diupload
		filename, err := helper.HandleFileImageUpload(c, "foto_profile", user.FotoProfile)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto profile",
				"details": err.Error(),
			})
			return
		}
		fotoProfileFilename = filename
	}
	
	// Selalu update foto_profile (bisa filename baru atau tetap yang lama)
	updates["foto_profile"] = fotoProfileFilename

	// ✅ SAFE: GORM Updates dengan parameterized queries
	if len(updates) > 0 {
		if err := uc.db.Model(&user).Updates(updates).Error; err != nil {
			// Jika gagal update, hapus file baru yang sudah diupload
			if fotoProfileFilename != user.FotoProfile {
				helper.DeleteOldPhoto(fotoProfileFilename, "profile_foto")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to update user",
				"details": err.Error(),
			})
			return
		}
	}

	// ✅ Reload user dengan data terbaru
	if err := uc.db.
		Preload("Level").
		Preload("Warga").

		First(&user, user.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated user",
		})
		return
	}

	// ✅ Jangan expose password hash
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "User updated successfully",
		"data":    user,
	})
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user by ID
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} map[string]interface{} "User deleted successfully"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "User not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/{id} [delete]
func (uc *UserController) DeleteUser(c *gin.Context) {
	userID := c.Param("id")

	// ✅ Validasi ID input
	if !isValidUserID(userID) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid user ID format",
		})
		return
	}

	var user models.User
	// ✅ SAFE: GORM First dengan parameterized query
	if err := uc.db.First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch user",
			})
		}
		return
	}

	// ✅ Prevent self-deletion (optional security feature)
	currentUserID, exists := c.Get("userID")
	if exists && currentUserID.(uint) == user.UserID {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Cannot delete your own account",
		})
		return
	}

	// ✅ Cek apakah user memiliki relasi dengan Warga
	var warga models.Warga
	if err := uc.db.Where("user_id = ?", user.UserID).First(&warga).Error; err == nil {
		// User memiliki data warga, jangan dihapus atau beri warning
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User memiliki data warga yang terkait. Hapus data warga terlebih dahulu.",
		})
		return
	}

	// ✅ SAFE: GORM Delete dengan parameterized query
	if err := uc.db.Delete(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete user",
		})
		return
	}

	// ✅ Hapus foto profile jika ada
	if user.FotoProfile != "" {
		helper.DeleteOldPhoto(user.FotoProfile, "foto_profile")
	}

	c.JSON(http.StatusOK, gin.H{
		"message":         "User deleted successfully",
		"deleted_user_id": user.UserID,
	})
}

// GetUserProfile godoc
// @Summary Get user profile
// @Description Get current authenticated user profile
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 401 {object} map[string]string "Unauthorized"
// @Failure 404 {object} map[string]string "User not found"
// @Router /users/profile [get]
func (uc *UserController) GetUserProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "User not authenticated",
		})
		return
	}

	var user models.User
	// ✅ SAFE: parameterized query
	if err := uc.db.
		Preload("Level").
		Preload("Warga").

		First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "User not found",
		})
		return
	}

	// ✅ Jangan expose sensitive data
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

// GetTotalUser godoc
// @Summary Get total users count
// @Description Get total number of registered users
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{} "Success response"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/total [get]
func (uc *UserController) GetTotalUser(c *gin.Context) {
	var total int64

	// ✅ SAFE: Parameterized query
	if err := uc.db.Model(&models.User{}).Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch total users",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_users": total,
		"message":     "Total users retrieved successfully",
	})
}

// GetFotoProfileImage godoc
// @Summary Get profile photo
// @Description Get user's profile photo by filename
// @Tags Users
// @Accept json
// @Produce image/*
// @Param filename path string true "Profile photo filename"
// @Success 200 {file} file "Profile photo file"
// @Failure 400 {object} map[string]string "Bad request"
// @Failure 404 {object} map[string]string "File not found"
// @Failure 500 {object} map[string]string "Internal server error"
// @Router /users/foto-profile/{filename} [get]
func (uc *UserController) GetFotoProfileImage(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	// Gunakan helper function GetFileByFileName
	file, err := helper.GetFileByFileName("profile_foto", filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File foto profile tidak ditemukan",
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

// SearchUsers godoc
// @Summary Search users
// @Description Search users by username, nama, email, or phone number
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param query query string false "Search query"
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} map[string]interface{} "Success response"
// @Router /users/search [get]
func (uc *UserController) SearchUsers(c *gin.Context) {
	query := sanitizeInput(strings.TrimSpace(c.Query("query")))
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	var users []models.User
	var total int64
	dbQuery := uc.db.Model(&models.User{}).
		Preload("Level").
		Preload("Warga").
		Preload("Warga.Keluarga").
		Preload("Warga.Agama").
		Preload("Warga.Pekerjaan").
		Preload("Warga.Rumah").
		Preload("Warga.TagihanIurans")

	if query != "" {
		searchPattern := "%" + query + "%"
		dbQuery = dbQuery.Where("username LIKE ? OR user_nama LIKE ? OR user_email LIKE ? OR user_no_telp LIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern)
	}

	// Hitung total
	dbQuery.Count(&total)

	// Ambil data dengan pagination
	offset := (page - 1) * limit
	if err := dbQuery.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search users",
		})
		return
	}

	// Jangan expose password
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{
		"data": users,
		"pagination": gin.H{
			"page":       page,
			"limit":      limit,
			"total":      total,
			"totalPages": (int(total) + limit - 1) / limit,
		},
	})
}