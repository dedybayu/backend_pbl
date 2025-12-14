package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupTagihanIuranRoutes(api *gin.RouterGroup, tagihanIuranController *controllers.TagihanIuranController, authMiddleware *middleware.AuthMiddleware) {
	tagihan := api.Group("/tagihan-iuran")
	{
		// Routes yang bisa diakses semua level (1-6)
		tagihan.GET("", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetAllTagihanIuran)
		tagihan.GET("/dropdown", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetTagihanIuranDropdown)
		tagihan.GET("/statistik", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetStatistikTagihanIuran)
		tagihan.GET("/:id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetTagihanIuranByID)
		tagihan.POST("/:id/bayar", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.BayarIuran) // Warga bisa bayar sendiri
		
		// Routes khusus admin (level 1)
		adminTagihan := tagihan.Group("")
		adminTagihan.Use(authMiddleware.RequireLevel(1))
		{
			adminTagihan.POST("", tagihanIuranController.CreateTagihanIuran)
			adminTagihan.PUT("/:id", tagihanIuranController.UpdateTagihanIuran)
			adminTagihan.DELETE("/:id", tagihanIuranController.DeleteTagihanIuran)
		}
	}
}