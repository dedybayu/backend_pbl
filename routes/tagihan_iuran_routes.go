// routes/pemasukan_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)
func SetupTagihanIuranRoutes(api *gin.RouterGroup, tagihanIuranController *controllers.TagihanIuranController, authMiddleware *middleware.AuthMiddleware) {
	tagihan := api.Group("/tagihan-iuran")
	{
		// Public routes (butuh auth)
		tagihan.GET("", authMiddleware.RequireLevel(1, 2), tagihanIuranController.GetAllTagihanIuran)
		tagihan.GET("/dropdown", authMiddleware.RequireLevel(1, 2), tagihanIuranController.GetTagihanIuranDropdown)
		tagihan.GET("/:id", authMiddleware.RequireLevel(1, 2), tagihanIuranController.GetTagihanIuranByID)
		
		// Admin only routes
		adminTagihan := tagihan.Group("")
		adminTagihan.Use(authMiddleware.RequireLevel(1))
		{
			adminTagihan.POST("", tagihanIuranController.CreateTagihanIuran)
			adminTagihan.PUT("/:id", tagihanIuranController.UpdateTagihanIuran)
			adminTagihan.DELETE("/:id", tagihanIuranController.DeleteTagihanIuran)
		}
	}
}