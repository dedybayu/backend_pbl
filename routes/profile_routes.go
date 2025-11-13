// routes/profile_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupProfileRoutes(api *gin.RouterGroup, profileController *controllers.ProfileController, authMiddleware *middleware.AuthMiddleware) {
	// Profile routes group
	profile := api.Group("/profile")
	
	// Protected routes - require authentication
	authProfile := profile.Group("")
	authProfile.Use(authMiddleware.Auth()) // atau authMiddleware.RequireToken() sesuai implementasi Anda
	{
		authProfile.GET("", profileController.GetProfile)
		authProfile.PUT("", profileController.UpdateProfile)
		authProfile.PUT("/change-password", profileController.UpdatePassword) // atau /change-password
		authProfile.GET("/image/:filename", profileController.GetFotoProfileImage) // Perbaiki parameter
	}
}