// routes/pengeluaran_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupPengeluaranRoutes(api *gin.RouterGroup, pengeluaranController *controllers.PengeluaranController, authMiddleware *middleware.AuthMiddleware) {
	pengeluaran := api.Group("/pengeluaran")
	{
		// Public routes (butuh auth)
		pengeluaran.GET("", authMiddleware.RequireLevel(1, 2), pengeluaranController.GetAllPengeluaran)
		pengeluaran.GET("/statistik", authMiddleware.RequireLevel(1, 2), pengeluaranController.GetStatistikPengeluaran)
		pengeluaran.GET("/laporan", authMiddleware.RequireLevel(1, 2), pengeluaranController.GetLaporanPengeluaranBulanan)
		pengeluaran.GET("/total-kategori", authMiddleware.RequireLevel(1, 2), pengeluaranController.GetTotalNominalPerKategori)
		pengeluaran.GET("/:id", authMiddleware.RequireLevel(1, 2), pengeluaranController.GetPengeluaranByID)
		
		// Admin only routes
		adminPengeluaran := pengeluaran.Group("")
		adminPengeluaran.Use(authMiddleware.RequireLevel(1))
		{
			adminPengeluaran.POST("", pengeluaranController.CreatePengeluaran)
			adminPengeluaran.PUT("/:id", pengeluaranController.UpdatePengeluaran)
			adminPengeluaran.DELETE("/:id", pengeluaranController.DeletePengeluaran)
		}
	}
}