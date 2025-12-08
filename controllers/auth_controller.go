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
	LevelID  uint         `json:"level_id"`
	UserNama string       `json:"user_nama"`
	UserAlamat string    `json:"user_alamat"`
	UserNoTelp string    `json:"user_no_telp"`
	UserEmail string    `json:"user_email"`
	Level    models.Level `json:"level"`
	Warga    *models.Warga `json:"warga,omitempty"` // hanya tampil jika tidak nil
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

// Login handles user authentication
func (ac *AuthController) Login(c *gin.Context) {
	var req LoginRequest

	contentType := c.GetHeader("Content-Type")

	var err error
	if strings.Contains(contentType, "application/json") {
		err = c.ShouldBindJSON(&req)
	} else if strings.Contains(contentType, "multipart/form-data") ||
		strings.Contains(contentType, "application/x-www-form-urlencoded") {
		err = c.ShouldBind(&req)
	} else {
		err = c.ShouldBind(&req)
		if err != nil {
			err = c.ShouldBindJSON(&req)
		}
	}

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User

	// PRELOAD LEVEL + WARGA (optional relationship)
	if err := ac.db.
		Preload("Level").
		Preload("Warga"). // penting
			Preload("Warga.Keluarga").
			Preload("Warga.Agama").
			Preload("Warga.Pekerjaan").
			Preload("Warga.Rumah").
			Preload("Warga.TagihanIurans").
		Where("username = ?", req.Username).
		First(&user).Error; err != nil {

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

	// Jika user tidak punya warga → user.Warga = nil (otomatis tidak tampil di JSON)
	response := LoginResponse{
		Token: token,
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			LevelID:  user.LevelID,
			UserNama: user.UserNama,
			UserAlamat: user.UserAlamat,
			UserNoTelp: user.UserNoTelp,
			UserEmail: user.UserEmail,
			Level:    user.Level,
			Warga:    user.Warga, // pointer, nil jika tidak punya
		},
		Level:     user.Level,
		LevelKode: user.Level.LevelKode,
	}

	c.JSON(http.StatusOK, response)
}

// UniversalLogin (opsional)
func (ac *AuthController) UniversalLogin(c *gin.Context) {
	var req LoginRequest

	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)

	if req.Username == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Username and password are required"})
		return
	}

	var user models.User
	if err := ac.db.Preload("Level").Preload("Warga").
		Where("username = ?", req.Username).First(&user).Error; err != nil {

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

	c.JSON(http.StatusOK, LoginResponse{
		Token: token,
		User: UserResponse{
			UserID:   user.UserID,
			Username: user.Username,
			LevelID:  user.LevelID,
			Level:    user.Level,
			Warga:    user.Warga,
		},
		Level:     user.Level,
		LevelKode: user.Level.LevelKode,
	})
}

// Logout
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

	ac.jwtUtils.AddToBlacklist(tokenString)

	c.JSON(http.StatusOK, LogoutResponse{
		Message: "Logout successful",
	})
}

// Refresh Token
func (ac *AuthController) RefreshToken(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	username, _ := c.Get("username")
	levelID, _ := c.Get("levelID")

	token, err := ac.jwtUtils.GenerateToken(
		userID.(uint),
		username.(string),
		levelID.(uint),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token":   token,
		"message": "Token refreshed successfully",
	})
}
