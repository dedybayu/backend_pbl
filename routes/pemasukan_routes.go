// routes/pemasukan_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPemasukanRoutes(api *gin.RouterGroup, pemasukanController *controllers.PemasukanController, authMiddleware *middleware.AuthMiddleware) {
	pemasukan := api.Group("/pemasukan")
	{
		// Public routes (butuh auth)
		pemasukan.GET("", authMiddleware.RequireLevel(1, 2), pemasukanController.GetAllPemasukan)
		pemasukan.GET("/statistik", authMiddleware.RequireLevel(1, 2), pemasukanController.GetStatistikPemasukan)
		pemasukan.GET("/laporan", authMiddleware.RequireLevel(1, 2), pemasukanController.GetLaporanPemasukanBulanan)
		pemasukan.GET("/total-kategori", authMiddleware.RequireLevel(1, 2), pemasukanController.GetTotalNominalPerKategori)
		pemasukan.GET("/:id", authMiddleware.RequireLevel(1, 2), pemasukanController.GetPemasukanByID)
		
		// Admin only routes
		adminPemasukan := pemasukan.Group("")
		adminPemasukan.Use(authMiddleware.RequireLevel(1))
		{
			adminPemasukan.POST("", pemasukanController.CreatePemasukan)
			adminPemasukan.PUT("/:id", pemasukanController.UpdatePemasukan)
			adminPemasukan.DELETE("/:id", pemasukanController.DeletePemasukan)
		}
	}
}