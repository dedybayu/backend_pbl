// routes/pengeluaran_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupKategoriPengeluaranRoutes(api *gin.RouterGroup, kategoriPengeluaranController *controllers.KategoriPengeluaranController, authMiddleware *middleware.AuthMiddleware) {
	kategori := api.Group("/kategori-pengeluaran")
	{
		// Public routes (butuh auth)
		kategori.GET("", authMiddleware.RequireLevel(1, 2), kategoriPengeluaranController.GetAllKategoriPengeluaran)
		kategori.GET("/dropdown", authMiddleware.RequireLevel(1, 2), kategoriPengeluaranController.GetKategoriPengeluaranDropdown)
		kategori.GET("/:id", authMiddleware.RequireLevel(1, 2), kategoriPengeluaranController.GetKategoriPengeluaranByID)
		
		// Admin only routes
		adminKategori := kategori.Group("")
		adminKategori.Use(authMiddleware.RequireLevel(1))
		{
			adminKategori.POST("", kategoriPengeluaranController.CreateKategoriPengeluaran)
			adminKategori.PUT("/:id", kategoriPengeluaranController.UpdateKategoriPengeluaran)
			adminKategori.DELETE("/:id", kategoriPengeluaranController.DeleteKategoriPengeluaran)
		}
	}
}