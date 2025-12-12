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
	Username   string `form:"username"`
	UserNama   string `form:"nama"`       // Sesuai field di model
	UserAlamat string `form:"alamat"`     // Sesuai field di model
	UserNoTelp string `form:"no_telp"`    // Sesuai field di model
	UserEmail  string `form:"email"`      // Sesuai field di model
}

type UpdatePasswordRequest struct {
	CurrentPassword string `form:"current_password" binding:"required"`
	NewPassword     string `form:"new_password" binding:"required"`
	ConfirmPassword string `form:"confirm_password" binding:"required"`
}

// GetProfile godoc
// @Summary      Mendapatkan data profile user
// @Description  Mendapatkan data lengkap profile user yang sedang login
// @Tags         profile
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]string  "Profile retrieved successfully"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "User not found"
// @Failure      500  {object}  map[string]string  "Failed to fetch profile"
// @Router       /profile [get]
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
	// Preload semua relasi yang diperlukan
	if err := pc.db.Preload("Level").Preload("Warga").First(&user, userID).Error; err != nil {
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

// UpdateProfile godoc
// @Summary      Mengupdate data profile
// @Description  Mengupdate data profile user (tanpa password)
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        nama      formData  string  false  "Nama lengkap user"
// @Param        username  formData  string  false  "Username user"
// @Param        alamat    formData  string  false  "Alamat user"
// @Param        no_telp   formData  string  false  "Nomor telepon user"
// @Param        email     formData  string  false  "Email user"
// @Param        foto_profile  formData  file    false  "Foto profile user"
// @Success      200  {object}  map[string]string  "Profile updated successfully"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "User not found"
// @Failure      409  {object}  map[string]string  "Username/email already exists"
// @Failure      500  {object}  map[string]string  "Failed to update profile"
// @Router       /profile [put]
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
	userNama := sanitizeInput(strings.TrimSpace(c.PostForm("nama")))
	userAlamat := sanitizeInput(strings.TrimSpace(c.PostForm("alamat")))
	userNoTelp := sanitizeInput(strings.TrimSpace(c.PostForm("no_telp")))
	userEmail := sanitizeInput(strings.TrimSpace(c.PostForm("email")))

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
			c.JSON(http.StatusConflict, gin.H{
				"error": "Username already exists",
			})
			return
		}
		updates["username"] = username
	}

	// Update email dengan validasi
	if userEmail != "" {
		if !isValidEmail(userEmail) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid email format",
			})
			return
		}

		// Check if email already exists (excluding current user)
		var count int64
		if err := pc.db.Model(&models.User{}).Where("user_email = ? AND user_id != ?", userEmail, userID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to check email availability",
			})
			return
		}

		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Email already exists",
			})
			return
		}
		updates["user_email"] = userEmail
	}

	// Update nama jika ada
	if userNama != "" {
		if len(userNama) < 2 || len(userNama) > 100 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Name must be between 2 and 100 characters",
			})
			return
		}
		updates["user_nama"] = userNama
	}

	// Update alamat jika ada
	if userAlamat != "" {
		if len(userAlamat) > 255 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Address is too long (max 255 characters)",
			})
			return
		}
		updates["user_alamat"] = userAlamat
	}

	// Update nomor telepon jika ada
	if userNoTelp != "" {
		if !isValidPhoneNumber(userNoTelp) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Invalid phone number format",
			})
			return
		}
		updates["user_no_telp"] = userNoTelp
	}

	// Handle file upload untuk foto_profile
	fotoProfileFilename := user.FotoProfile
	if file, header, err := c.Request.FormFile("foto_profile"); err == nil && header != nil {
		defer file.Close()
		
		fmt.Printf("✅ DEBUG: File upload detected - %s (%d bytes)\n", header.Filename, header.Size)
		
		filename, err := helper.HandleFileImageUpload(c, "foto_profile", user.FotoProfile)
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

	// Update foto_profile
	if fotoProfileFilename != user.FotoProfile {
		updates["foto_profile"] = fotoProfileFilename
	}

	// Eksekusi update hanya jika ada field yang diupdate
	if len(updates) > 0 {
		if err := pc.db.Model(&user).Updates(updates).Error; err != nil {
			// Jika gagal update, hapus file baru yang sudah diupload
			if fotoProfileFilename != user.FotoProfile {
				helper.DeleteOldPhoto(fotoProfileFilename, "profile_foto")
			}
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to update profile",
				"details": err.Error(),
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

// UpdatePassword godoc
// @Summary      Mengupdate password user
// @Description  Mengupdate password user dengan validasi password lama
// @Tags         profile
// @Accept       multipart/form-data
// @Produce      json
// @Security     BearerAuth
// @Param        current_password  formData  string  true  "Password saat ini"
// @Param        new_password      formData  string  true  "Password baru"
// @Param        confirm_password  formData  string  true  "Konfirmasi password baru"
// @Success      200  {object}  map[string]string  "Password updated successfully"
// @Failure      400  {object}  map[string]string  "Invalid request"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "User not found"
// @Failure      500  {object}  map[string]string  "Failed to update password"
// @Router       /profile/change-password [put]
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

// GetFotoProfileImage godoc
// @Summary      Mendapatkan gambar foto profile
// @Description  Mengambil file gambar foto profile berdasarkan nama file
// @Tags         profile
// @Produce      image/jpeg,image/png,image/gif
// @Security     BearerAuth
// @Param        filename  path      string  true  "Nama file gambar"
// @Success      200  {file}    binary  "File gambar"
// @Failure      400  {object}  map[string]string  "Nama file tidak valid"
// @Failure      401  {object}  map[string]string  "Unauthorized"
// @Failure      404  {object}  map[string]string  "File tidak ditemukan"
// @Failure      500  {object}  map[string]string  "Gagal membuka file"
// @Router       /profile/image/{filename} [get]
func (pc *ProfileController) GetFotoProfileImage(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	// Gunakan helper function GetFileByFileName
	file, err := helper.GetFileByFileName("foto_profile", filename)
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
