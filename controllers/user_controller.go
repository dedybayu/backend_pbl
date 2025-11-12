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

type CreateUserRequest struct {
	Username    string `form:"username" binding:"required"`
	Password    string `form:"password" binding:"required"`
	LevelID     uint   `form:"level_id" binding:"required"`
	FotoProfile string `form:"foto_profile"` // Tetap string untuk filename
}

type UpdateUserRequest struct {
	Username    string `form:"username"`
	Password    string `form:"password"`
	LevelID     uint   `form:"level_id"`
	FotoProfile string `form:"foto_profile"`
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

// GetAllUsers returns all users with security checks
func (uc *UserController) GetAllUsers(c *gin.Context) {
	var users []models.User

	// ✅ SAFE: GORM menggunakan parameterized queries internally
	if err := uc.db.Preload("Level").Find(&users).Error; err != nil {
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

// GetUserByID returns user by ID with security validation
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
	if err := uc.db.Preload("Level").First(&user, userID).Error; err != nil {
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

// CreateUser creates new user with comprehensive security checks
func (uc *UserController) CreateUser(c *gin.Context) {
	// Binding manual untuk form data
	username := sanitizeInput(strings.TrimSpace(c.PostForm("username")))
	password := strings.TrimSpace(c.PostForm("password"))
	levelIDStr := c.PostForm("level_id")

	// Validasi required fields
	if username == "" || password == "" || levelIDStr == "" {
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

	// ✅ Validasi level ID (prevent invalid level injection)
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

	// ✅ Check if username already exists - SAFE: parameterized query
	var existingUser models.User
	if err := uc.db.Where("username = ?", username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username already exists",
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

	user := models.User{
		Username:    username,
		Password:    string(hashedPassword),
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
			"error": "Failed to create user",
		})
		return
	}

	// ✅ Reload user dengan data terbaru
	if err := uc.db.Preload("Level").First(&user, user.UserID).Error; err != nil {
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

// UpdateUser updates user data dengan security checks
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

		// ✅ Check if username already exists (excluding current user) - SAFE: parameterized query
		var existingUser models.User
		if err := uc.db.Where("username = ? AND user_id != ?", username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username already exists",
			})
			return
		}
		updates["username"] = username
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
				"error": "Failed to update user",
			})
			return
		}
	}

	// ✅ Reload user dengan data terbaru
	if err := uc.db.Preload("Level").First(&user, user.UserID).Error; err != nil {
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

// DeleteUser deletes user dengan security checks
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

// GetUserProfile returns current user profile (additional security feature)
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
	if err := uc.db.Preload("Level").First(&user, userID).Error; err != nil {
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

// GetTotalUser returns total number of users
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

// ✅ GET - Mendapatkan gambar foto profile
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