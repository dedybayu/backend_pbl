// routes/user_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupUserRoutes(api *gin.RouterGroup, userController *controllers.UserController, authMiddleware *middleware.AuthMiddleware) {
	users := api.Group("/users")
	{
		// Public routes (butuh auth)
		users.GET("", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), userController.GetAllUsers)
		users.GET("/profile", authMiddleware.Auth(), userController.GetUserProfile) // ✅ NEW: Get current user profile
		users.GET("/image/:filename", authMiddleware.Auth(), userController.GetFotoProfileImage) 
		users.GET("/:id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), userController.GetUserByID)

		
		// Admin only routes
		adminUsers := users.Group("")
		adminUsers.Use(authMiddleware.RequireLevel(1))
		{
			adminUsers.POST("", userController.CreateUser)
			adminUsers.PUT("/:id", userController.UpdateUser)
			adminUsers.DELETE("/:id", userController.DeleteUser)
			
		}
	}
}