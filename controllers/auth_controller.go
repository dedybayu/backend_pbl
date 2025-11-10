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

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token string      `json:"token"`
	User  models.User `json:"user"`
	Level models.Level `json:"level"`
}

type LogoutResponse struct {
	Message string `json:"message"`
}

type ProfileResponse struct {
	User  models.User `json:"user"`
	Level models.Level `json:"level"`
}

// Login handles user authentication
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		User:  user,
		Level: user.Level,
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

// GetProfile returns current user profile
func (ac *AuthController) GetProfile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := ac.db.Preload("Level").First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	response := ProfileResponse{
		User:  user,
		Level: user.Level,
	}

	c.JSON(http.StatusOK, response)
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
		"token": token,
		"message": "Token refreshed successfully",
	})
}