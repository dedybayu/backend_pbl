// controllers/user_controller.go
package controllers

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"rt-management/models"

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
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	LevelID  uint   `json:"level_id" binding:"required"`
}

type UpdateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	LevelID  uint   `json:"level_id"`
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
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	req.Username = sanitizeInput(req.Username)
	req.Password = strings.TrimSpace(req.Password) // Password tidak di-sanitize berlebihan

	// ✅ Validasi username format
	if !isValidUsername(req.Username) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username must be 3-20 characters and contain only letters, numbers, and underscores",
		})
		return
	}

	// ✅ Validasi password strength
	if !isValidPassword(req.Password) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Password must be at least 8 characters and contain both letters and numbers",
		})
		return
	}

	// ✅ Validasi level ID (prevent invalid level injection)
	if req.LevelID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Level ID is required",
		})
		return
	}

	// ✅ Check if level exists - SAFE: parameterized query
	var level models.Level
	if err := uc.db.First(&level, req.LevelID).Error; err != nil {
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
	if err := uc.db.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Username already exists",
		})
		return
	}

	// ✅ Hash password dengan cost yang aman
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to secure password",
		})
		return
	}

	user := models.User{
		Username: req.Username,
		Password: string(hashedPassword),
		LevelID:  req.LevelID,
	}

	// ✅ SAFE: GORM create dengan parameterized queries
	if err := uc.db.Create(&user).Error; err != nil {
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

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// ✅ Sanitize input
	if req.Username != "" {
		req.Username = sanitizeInput(req.Username)
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

	// ✅ Check if level exists if provided
	if req.LevelID != 0 {
		var level models.Level
		// ✅ SAFE: parameterized query
		if err := uc.db.First(&level, req.LevelID).Error; err != nil {
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
		user.LevelID = req.LevelID
	}

	// ✅ Update username dengan validasi
	if req.Username != "" {
		if !isValidUsername(req.Username) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username must be 3-20 characters and contain only letters, numbers, and underscores",
			})
			return
		}

		// ✅ Check if username already exists (excluding current user) - SAFE: parameterized query
		var existingUser models.User
		if err := uc.db.Where("username = ? AND user_id != ?", req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Username already exists",
			})
			return
		}
		user.Username = req.Username
	}

	// ✅ Update password dengan validasi
	if req.Password != "" {
		if !isValidPassword(req.Password) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Password must be at least 8 characters and contain both letters and numbers",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to secure password",
			})
			return
		}
		user.Password = string(hashedPassword)
	}

	// ✅ SAFE: GORM Save dengan parameterized queries
	if err := uc.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update user",
		})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"message": "User deleted successfully",
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