// middleware/auth_middleware.go
package middleware

import (
	"net/http"
	"strings"
	"rt-management/utils"
	"log"

	"github.com/gin-gonic/gin"
)

type AuthMiddleware struct {
	jwtUtils *utils.JWTUtils
}

func NewAuthMiddleware(jwtUtils *utils.JWTUtils) *AuthMiddleware {
	return &AuthMiddleware{jwtUtils: jwtUtils}
}

// Auth middleware untuk validasi JWT token
func (m *AuthMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip auth untuk public endpoints
		if m.isPublicEndpoint(c.Request.URL.Path) {
			c.Next()
			return
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}

		tokenString := parts[1]
		
		// Log attempt untuk security monitoring
		log.Printf("🔐 Authentication attempt for path: %s", c.Request.URL.Path)
		
		claims, err := m.jwtUtils.ValidateToken(tokenString)
		if err != nil {
			log.Printf("❌ Authentication failed: %v", err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Set user data ke context
		c.Set("userID", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("levelID", claims.LevelID)
		
		log.Printf("✅ User authenticated: %s (ID: %d, Level: %d)", 
			claims.Username, claims.UserID, claims.LevelID)
		
		c.Next()
	}
}

// RequireLevel middleware untuk check level user
func (m *AuthMiddleware) RequireLevel(allowedLevels ...uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		levelID, exists := c.Get("levelID")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "User level not found"})
			c.Abort()
			return
		}

		userLevelID := levelID.(uint)
		for _, allowedLevel := range allowedLevels {
			if userLevelID == allowedLevel {
				c.Next()
				return
			}
		}

		username, _ := c.Get("username")
		log.Printf("🚫 Access denied for user %s to %s (Required levels: %v, User level: %d)", 
			username, c.Request.URL.Path, allowedLevels, userLevelID)
		
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Insufficient permissions",
			"message": "You don't have permission to access this resource",
		})
		c.Abort()
	}
}

// OptionalAuth middleware yang tidak mewajibkan auth, tapi tetap set user data jika ada
func (m *AuthMiddleware) OptionalAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.Next()
			return
		}

		tokenString := parts[1]
		if claims, err := m.jwtUtils.ValidateToken(tokenString); err == nil {
			c.Set("userID", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("levelID", claims.LevelID)
		}

		c.Next()
	}
}

// isPublicEndpoint check apakah endpoint tidak memerlukan auth
func (m *AuthMiddleware) isPublicEndpoint(path string) bool {
	publicEndpoints := []string{
		"/auth/login",
		"/health",
		"/api/docs",
	}

	for _, endpoint := range publicEndpoints {
		if path == endpoint {
			return true
		}
	}

	return false
}