package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKegiatanRoutes(api *gin.RouterGroup, kegiatanController *controllers.KegiatanController, authMiddleware *middleware.AuthMiddleware) {
	kegiatan := api.Group("/kegiatan")
	{
		// Public routes (butuh auth)
		kegiatan.GET("", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), kegiatanController.GetAllKegiatan)
		kegiatan.GET("/mendatang", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), kegiatanController.GetKegiatanMendatang)
		kegiatan.GET("/statistik", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), kegiatanController.GetStatistikKegiatan)
		kegiatan.GET("/search", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), kegiatanController.SearchKegiatan)
		kegiatan.GET("/:id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), kegiatanController.GetKegiatanByID)
		
		// Admin only routes
		adminKategoriKegiatan := kegiatan.Group("")
		adminKategoriKegiatan.Use(authMiddleware.RequireLevel(1, 2, 3, 4, 5))
		{
			adminKategoriKegiatan.POST("", kegiatanController.CreateKegiatan)
			adminKategoriKegiatan.PUT("/:id", kegiatanController.UpdateKegiatan)
			adminKategoriKegiatan.DELETE("/:id", kegiatanController.DeleteKegiatan)
			// adminKategoriKegiatan.GET("/:tahun/:bulan", kegiatanController.GetKegiatanByBulanTahun)
		}
	}
}