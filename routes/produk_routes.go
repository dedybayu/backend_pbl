// routes/produk_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupProdukRoutes(api *gin.RouterGroup, produkController *controllers.ProdukController, authMiddleware *middleware.AuthMiddleware) {
	produk := api.Group("/produk")
	{
		// Public routes (butuh auth)
		produk.GET("", authMiddleware.RequireLevel(1, 2), produkController.GetAllProduk)
		produk.GET("/terbaru", authMiddleware.RequireLevel(1, 2), produkController.GetProdukTerbaru)
		produk.GET("/stok-menipis", authMiddleware.RequireLevel(1, 2), produkController.GetProdukStokMenipis)
		produk.GET("/statistik", authMiddleware.RequireLevel(1, 2), produkController.GetStatistikProduk)
		produk.GET("/:id", authMiddleware.RequireLevel(1, 2), produkController.GetProdukByID)
		
		// Admin only routes
		adminProduk := produk.Group("")
		adminProduk.Use(authMiddleware.RequireLevel(1))
		{
			adminProduk.POST("", produkController.CreateProduk)
			adminProduk.PUT("/:id", produkController.UpdateProduk)
			adminProduk.PATCH("/:id/stok", produkController.UpdateStokProduk)
			adminProduk.DELETE("/:id", produkController.DeleteProduk)
		}
	}
}