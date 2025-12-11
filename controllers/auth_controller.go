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

type AuthUserResponse struct {
	UserID      uint            `json:"user_id" example:"1"`
	Username    string          `json:"username" example:"admin"`
	LevelID     uint            `json:"level_id" example:"1"`
	UserNama    string          `json:"user_nama" example:"Administrator"`
	UserAlamat  string          `json:"user_alamat" example:"Jl. Contoh No. 123"`
	UserNoTelp  string          `json:"user_no_telp" example:"081234567890"`
	UserEmail   string          `json:"user_email" example:"admin@example.com"`
	FotoProfile string          `json:"foto_profile" example:"profile.jpg"`
	Level       models.Level    `json:"level"`
	Warga       *models.Warga   `json:"warga,omitempty"` // hanya tampil jika tidak nil
}

type LoginRequest struct {
	Username string `json:"username" form:"username" xml:"username" binding:"required" example:"admin"`
	Password string `json:"password" form:"password" xml:"password" binding:"required" example:"password123"`
}

type LoginResponse struct {
	Token     string       `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	User      AuthUserResponse `json:"user"`
	Level     models.Level `json:"level"`
	LevelKode string       `json:"level_kode" example:"ADM"`
}

type LogoutResponse struct {
	Message string `json:"message" example:"Logout successful"`
}

type ProfileResponse struct {
	User  models.User  `json:"user"`
	Level models.Level `json:"level"`
}

type RefreshTokenResponse struct {
	Token   string `json:"token" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	Message string `json:"message" example:"Token refreshed successfully"`
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with username and password
// @Tags         auth
// @Accept       json
// @Accept       multipart/form-data
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        request body LoginRequest true "Login credentials"
// @Success      200  {object}  LoginResponse  "Login successful"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      401  {object}  map[string]interface{}  "Invalid credentials"
// @Failure      500  {object}  map[string]interface{}  "Failed to generate token"
// @Router       /auth/login [post]
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
		User: AuthUserResponse{
			UserID:      user.UserID,
			Username:    user.Username,
			LevelID:     user.LevelID,
			UserNama:    user.UserNama,
			UserAlamat:  user.UserAlamat,
			UserNoTelp:  user.UserNoTelp,
			UserEmail:   user.UserEmail,
			FotoProfile: user.FotoProfile,
			Level:       user.Level,
			Warga:       user.Warga, // pointer, nil jika tidak punya
		},
		Level:     user.Level,
		LevelKode: user.Level.LevelKode,
	}

	c.JSON(http.StatusOK, response)
}

// UniversalLogin godoc
// @Summary      Universal login
// @Description  Alternative login endpoint that accepts multiple content types
// @Tags         auth
// @Accept       multipart/form-data
// @Accept       x-www-form-urlencoded
// @Produce      json
// @Param        username formData string true "Username"
// @Param        password formData string true "Password"
// @Success      200  {object}  LoginResponse  "Login successful"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      401  {object}  map[string]interface{}  "Invalid credentials"
// @Failure      500  {object}  map[string]interface{}  "Failed to generate token"
// @Router       /auth/universal-login [post]
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
		User: AuthUserResponse{
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

// Logout godoc
// @Summary      Logout user
// @Description  Invalidate user's JWT token
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  LogoutResponse  "Logout successful"
// @Failure      400  {object}  map[string]interface{}  "Invalid authorization header"
// @Router       /auth/logout [post]
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

// RefreshToken godoc
// @Summary      Refresh JWT token
// @Description  Generate a new JWT token using current valid token
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  RefreshTokenResponse  "Token refreshed"
// @Failure      401  {object}  map[string]interface{}  "User not authenticated"
// @Failure      500  {object}  map[string]interface{}  "Failed to generate token"
// @Router       /auth/refresh-token [post]
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

// Profile godoc
// @Summary      Get user profile
// @Description  Get authenticated user's profile information
// @Tags         auth
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  ProfileResponse  "User profile"
// @Failure      401  {object}  map[string]interface{}  "User not authenticated"
// @Failure      404  {object}  map[string]interface{}  "User not found"
// @Failure      500  {object}  map[string]interface{}  "Failed to get user profile"
// @Router       /auth/profile [get]
func (ac *AuthController) Profile(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var user models.User
	if err := ac.db.
		Preload("Level").
		Preload("Warga").
		Preload("Warga.Keluarga").
		Preload("Warga.Agama").
		Preload("Warga.Pekerjaan").
		Preload("Warga.Rumah").
		Preload("Warga.TagihanIurans").
		First(&user, userID).Error; err != nil {

		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user profile"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"user":  user,
		"level": user.Level,
	})
}