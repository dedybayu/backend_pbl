// routes/produk_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKategoriProdukRoutes(api *gin.RouterGroup, kategoriProdukController *controllers.KategoriProdukController, authMiddleware *middleware.AuthMiddleware) {
	kategori := api.Group("/kategori-produk")
	{
		// Public routes (butuh auth)
		kategori.GET("", authMiddleware.RequireLevel(1, 2), kategoriProdukController.GetAllKategoriProduk)
		kategori.GET("/dropdown", authMiddleware.RequireLevel(1, 2), kategoriProdukController.GetKategoriProdukDropdown)
		kategori.GET("/statistik", authMiddleware.RequireLevel(1, 2), kategoriProdukController.GetStatistikKategoriProduk)
		kategori.GET("/:id", authMiddleware.RequireLevel(1, 2), kategoriProdukController.GetKategoriProdukByID)
		
		// Admin only routes
		adminKategori := kategori.Group("")
		adminKategori.Use(authMiddleware.RequireLevel(1))
		{
			adminKategori.POST("", kategoriProdukController.CreateKategoriProduk)
			adminKategori.PUT("/:id", kategoriProdukController.UpdateKategoriProduk)
			adminKategori.DELETE("/:id", kategoriProdukController.DeleteKategoriProduk)
		}
	}
}