// routes/level_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupLevelRoutes(api *gin.RouterGroup, levelController *controllers.LevelController, authMiddleware *middleware.AuthMiddleware) {
	levels := api.Group("/levels")
	{
		levels.GET("", levelController.GetAllLevels)
		
		// Admin only routes
		adminLevels := levels.Group("")
		adminLevels.Use(authMiddleware.RequireLevel(1))
		{
			adminLevels.POST("", levelController.CreateLevel)
			adminLevels.GET("/:id", levelController.GetLevelByID)
			adminLevels.PUT("/:id", levelController.UpdateLevel)
			adminLevels.DELETE("/:id", levelController.DeleteLevel)
		}
	}
}