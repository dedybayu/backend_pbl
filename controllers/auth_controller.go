// controllers/auth_controller.go
package controllers

import (
	"net/http"
	"rt-management/models"
	"rt-management/utils"
	"strings"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type AuthController struct {
	db       *gorm.DB
	jwtUtils *utils.JWTUtils
}

func NewAuthController(db *gorm.DB, jwtUtils *utils.JWTUtils) *AuthController {
	return &AuthController{db: db, jwtUtils: jwtUtils}
}

type UserResponse struct {
	UserID   uint         `json:"user_id"`
	Username string       `json:"username"`
	// Nama     string       `json:"nama"`
	LevelID  uint         `json:"level_id"`
	Level    models.Level `json:"level"`
}

type LoginRequest struct {
	Username string `json:"username" form:"username" xml:"username" binding:"required"`
	Password string `json:"password" form:"password" xml:"password" binding:"required"`
}

type LoginResponse struct {
	Token     string       `json:"token"`
	User      UserResponse `json:"user"`
	Level     models.Level `json:"level"`
	LevelKode string       `json:"level_kode"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type ProfileResponse struct {
	User  models.User  `json:"user"`
	Level models.Level `json:"level"`
}

// Login handles user authentication - Support multiple content types
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	// Deteksi content type dan bind sesuai
	contentType := c.GetHeader("Content-Type")

	var err error
	if strings.Contains(contentType, "application/json") {
		// Handle JSON
		err = c.ShouldBindJSON(&req)
	} else if strings.Contains(contentType, "multipart/form-data") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") {
		// Handle form-data dan x-www-form-urlencoded
		err = c.ShouldBind(&req)
	} else {
		// Default: try both
		err = c.ShouldBind(&req)
		if err != nil {
			err = c.ShouldBindJSON(&req)
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
			"supported_formats": []string{
				"application/json",
				"multipart/form-data",
				"application/x-www-form-urlencoded",
			},
		})
		return
	}

	// Sanitize input
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Validasi input
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User
	if err := ac.db.Preload("Level").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := ac.jwtUtils.GenerateToken(user.UserID, user.Username, user.LevelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			// Nama:     user.Nama,
			LevelID: user.LevelID,
			Level:   user.Level,
		},
		Level:     user.Level,
		LevelKode: user.Level.LevelKode,
	}

	c.JSON(http.StatusOK, response)
}

// UniversalLogin - Alternatif: menggunakan ShouldBind yang support semua format
func (ac *AuthController) UniversalLogin(c *gin.Context) {
	var req LoginRequest

	// Gunakan ShouldBind yang support multiple content types
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
			"supported_content_types": []string{
				"application/json",
				"multipart/form-data",
				"application/x-www-form-urlencoded",
				"application/xml",
			},
		})
		return
	}

	// Sanitize input
	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	// Validasi input
	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User
	if err := ac.db.Preload("Level").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	token, err := ac.jwtUtils.GenerateToken(user.UserID, user.Username, user.LevelID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := LoginResponse{
		Token: token,
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			// Nama:     user.Nama,
			LevelID: user.LevelID,
			Level:   user.Level,
		},
		Level:     user.Level,
		LevelKode: user.Level.LevelKode,
	}

	c.JSON(http.StatusOK, response)
}

// Logout handles user logout by blacklisting the token
func (ac *AuthController) Logout(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization header required"})
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid authorization format"})
		return
	}

	tokenString := parts[1]

	// Tambahkan token ke blacklist
	if err := ac.jwtUtils.AddToBlacklist(tokenString); err != nil {
		// Jika gagal parse token, tetap return success
		c.JSON(http.StatusOK, LogoutResponse{
			Message: "Logout successful",
		})
		return
	}

	c.JSON(http.StatusOK, LogoutResponse{
		Message: "Logout successful",
	})
}

// RefreshToken generates new token (optional)
func (ac *AuthController) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, _ := c.Get("username")
	levelID, _ := c.Get("levelID")

	// Generate new token
	token, err := ac.jwtUtils.GenerateToken(
		userID.(uint),
		username.(string),
		levelID.(uint),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	// Blacklist old token
	authHeader := c.GetHeader("Authorization")
	if authHeader != "" {
		parts := strings.Split(authHeader, " ")
		if len(parts) == 2 && parts[0] == "Bearer" {
			ac.jwtUtils.AddToBlacklist(parts[1])
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Token refreshed successfully",
	})
}

// CheckAuthStatus - Check if user is authenticated
func (ac *AuthController) CheckAuthStatus(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"authenticated": false,
			"message":       "Not authenticated",
		})
		return
	}

	var user models.User
	if err := ac.db.Preload("Level").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"authenticated": false,
			"message":       "User not found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"authenticated": true,
		"user": gin.H{
			"user_id":  user.UserID,
			"username": user.Username,
			"level":    user.Level,
		},
	})
}
