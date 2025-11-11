// routes/mutasi_keluarga_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupMutasiKeluargaRoutes(api *gin.RouterGroup, mutasiKeluargaController *controllers.MutasiKeluargaController, authMiddleware *middleware.AuthMiddleware) {
	mutasi := api.Group("/mutasi-keluarga")
	{
		// Public routes (butuh auth)
		mutasi.GET("", authMiddleware.RequireLevel(1, 2), mutasiKeluargaController.GetAllMutasiKeluarga)
		mutasi.GET("/statistik", authMiddleware.RequireLevel(1, 2), mutasiKeluargaController.GetStatistikMutasiKeluarga)
		mutasi.GET("/laporan", authMiddleware.RequireLevel(1, 2), mutasiKeluargaController.GetLaporanMutasiBulanan)
		mutasi.GET("/:id", authMiddleware.RequireLevel(1, 2), mutasiKeluargaController.GetMutasiKeluargaByID)
		mutasi.GET("/keluarga/:keluarga_id", authMiddleware.RequireLevel(1, 2), mutasiKeluargaController.GetMutasiByKeluargaID)
		
		// Admin only routes
		adminMutasi := mutasi.Group("")
		adminMutasi.Use(authMiddleware.RequireLevel(1))
		{
			adminMutasi.POST("", mutasiKeluargaController.CreateMutasiKeluarga)
			adminMutasi.PUT("/:id", mutasiKeluargaController.UpdateMutasiKeluarga)
			adminMutasi.DELETE("/:id", mutasiKeluargaController.DeleteMutasiKeluarga)
		}
	}
}