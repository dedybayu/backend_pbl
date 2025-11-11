// routes/pemasukan_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKategoriPemasukanRoutes(api *gin.RouterGroup, kategoriPemasukanController *controllers.KategoriPemasukanController, authMiddleware *middleware.AuthMiddleware) {
	kategori := api.Group("/kategori-pemasukan")
	{
		// Public routes (butuh auth)
		kategori.GET("", authMiddleware.RequireLevel(1, 2), kategoriPemasukanController.GetAllKategoriPemasukan)
		kategori.GET("/dropdown", authMiddleware.RequireLevel(1, 2), kategoriPemasukanController.GetKategoriPemasukanDropdown)
		kategori.GET("/:id", authMiddleware.RequireLevel(1, 2), kategoriPemasukanController.GetKategoriPemasukanByID)
		
		// Admin only routes
		adminKategori := kategori.Group("")
		adminKategori.Use(authMiddleware.RequireLevel(1))
		{
			adminKategori.POST("", kategoriPemasukanController.CreateKategoriPemasukan)
			adminKategori.PUT("/:id", kategoriPemasukanController.UpdateKategoriPemasukan)
			adminKategori.DELETE("/:id", kategoriPemasukanController.DeleteKategoriPemasukan)
		}
	}
}