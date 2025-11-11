

// routes/warga_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKategoriKegiatanRoutes(api *gin.RouterGroup, kategoriKegiatanController *controllers.KategoriKegiatanController, authMiddleware *middleware.AuthMiddleware) {
	kategori_kegiatan := api.Group("/kategori-kegiatan")
	{
		// Public routes (butuh auth)
		kategori_kegiatan.GET("", authMiddleware.RequireLevel(1, 2), kategoriKegiatanController.CreateKategoriKegiatan)
		kategori_kegiatan.GET("/:id", authMiddleware.RequireLevel(1, 2), kategoriKegiatanController.GetKategoriKegiatanByID)
		kategori_kegiatan.GET("/dropdown", authMiddleware.RequireLevel(1, 2), kategoriKegiatanController.GetKategoriKegiatanDropdown)
		
		// Admin only routes
		adminKategoriKegiatan := kategori_kegiatan.Group("")
		adminKategoriKegiatan.Use(authMiddleware.RequireLevel(1))
		{
			adminKategoriKegiatan.POST("", kategoriKegiatanController.CreateKategoriKegiatan)
			adminKategoriKegiatan.PUT("/:id", kategoriKegiatanController.UpdateKategoriKegiatan)
			adminKategoriKegiatan.DELETE("/:id", kategoriKegiatanController.DeleteKategoriKegiatan)
		}
	}
}