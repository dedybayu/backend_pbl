// routes/auth_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupAuthRoutes(router *gin.Engine, authController *controllers.AuthController, authMiddleware *middleware.AuthMiddleware) {
	auth := router.Group("/auth")
	{
		auth.POST("/login", authController.Login)
		auth.POST("/logout", authMiddleware.Auth(), authController.Logout)
		auth.POST("/refresh", authMiddleware.Auth(), authController.RefreshToken)
	}
}