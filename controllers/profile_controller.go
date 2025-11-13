// controllers/profile_controller.go
package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"rt-management/helper"
	"rt-management/models"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type ProfileController struct {
	db *gorm.DB
}

func NewProfileController(db *gorm.DB) *ProfileController {
	return &ProfileController{db: db}
}

// Request structs
type UpdateProfileRequest struct {
	Username    string `form:"username"`
	FotoProfile string `form:"foto_profile"`
}

type UpdatePasswordRequest struct {
	CurrentPassword string `form:"current_password" binding:"required"`
	NewPassword     string `form:"new_password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" binding:"required"`
}

// GetProfile mendapatkan data profile user yang sedang login
func (pc *ProfileController) GetProfile(c *gin.Context) {
	// Ambil userID dari context (setelah authentication)
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var user models.User
	if err := pc.db.Preload("Level").First(&user, userID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "User not found",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch profile",
			})
		}
		return
	}

	// Jangan expose password hash
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile retrieved successfully",
		"data":    user,
	})
}

// UpdateProfile mengupdate data profile (tanpa password)
func (pc *ProfileController) UpdateProfile(c *gin.Context) {
	// Ambil userID dari context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var user models.User
	if err := pc.db.First(&user, userID).Error; err != nil {
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

	// Update fields menggunakan map
	updates := make(map[string]interface{})

	// Update username dengan validasi
	if username != "" {
		if !isValidUsername(username) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username must be 3-20 characters and contain only letters, numbers, and underscores",
			})
			return
		}

		// Check if username already exists (excluding current user)
		var count int64
		if err := pc.db.Model(&models.User{}).Where("username = ? AND user_id != ?", username, userID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check username availability",
			})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username already exists",
			})
			return
		}
		updates["username"] = username
	}

	// Handle file upload untuk foto_profile
	fotoProfileFilename := user.FotoProfile
	if file, header, err := c.Request.FormFile("foto_profile"); err == nil && header != nil {
		defer file.Close()
		
		fmt.Printf("✅ DEBUG: File upload detected - %s (%d bytes)\n", header.Filename, header.Size)
		
		filename, err := helper.HandleFileImageUpload(c, "profile_foto", user.FotoProfile)
		if err != nil {
			fmt.Printf("❌ DEBUG: Upload failed: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto profile",
				"details": err.Error(),
			})
			return
		}
		fotoProfileFilename = filename
		fmt.Printf("✅ DEBUG: Upload success, new filename: %s\n", filename)
	} else {
		if err != nil && err != http.ErrMissingFile {
			fmt.Printf("🔍 DEBUG: FormFile error: %v\n", err)
		}
	}

	// Selalu update foto_profile
	updates["foto_profile"] = fotoProfileFilename

	// Eksekusi update hanya jika ada field yang diupdate
	if len(updates) > 0 {
		if err := pc.db.Model(&user).Updates(updates).Error; err != nil {
			// Jika gagal update, hapus file baru yang sudah diupload
			if fotoProfileFilename != user.FotoProfile {
				helper.DeleteOldPhoto(fotoProfileFilename, "profile_foto")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profile",
			})
			return
		}
	}

	// Reload user dengan data terbaru
	if err := pc.db.Preload("Level").First(&user, user.UserID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to load updated profile",
		})
		return
	}

	// Jangan expose password hash
	user.Password = ""

	c.JSON(http.StatusOK, gin.H{
		"message": "Profile updated successfully",
		"data":    user,
	})
}

// UpdatePassword mengupdate password user
func (pc *ProfileController) UpdatePassword(c *gin.Context) {
	// Ambil userID dari context
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Unauthorized",
		})
		return
	}

	var req UpdatePasswordRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)
	req.ConfirmPassword = strings.TrimSpace(req.ConfirmPassword)

	// Validasi new password dan confirm password
	if req.NewPassword != req.ConfirmPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "New password and confirm password do not match",
		})
		return
	}

	// Validasi new password strength
	if !isValidPassword(req.NewPassword) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "New password must be at least 8 characters and contain both letters and numbers",
		})
		return
	}

	// Validasi new password tidak sama dengan current password
	if req.NewPassword == req.CurrentPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "New password must be different from current password",
		})
		return
	}

	var user models.User
	if err := pc.db.First(&user, userID).Error; err != nil {
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

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Current password is incorrect",
		})
		return
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to secure new password",
		})
		return
	}

	// Update password
	if err := pc.db.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update password",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Password updated successfully",
	})
}

// GetFotoProfileImage mendapatkan gambar foto profile
func (pc *ProfileController) GetFotoProfileImage(c *gin.Context) {
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