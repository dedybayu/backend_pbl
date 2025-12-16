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
		tagihan.GET("", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetAll)
		tagihan.GET("/:id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetByID)
		tagihan.GET("/warga/:warga_id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.GetByWargaID)
		tagihan.POST("/:id/bayar", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), tagihanIuranController.BayarIuran)

		// Routes file bukti (public access untuk melihat bukti)
		tagihan.GET("/files/:filename", tagihanIuranController.GetFileBukti)

		// Routes khusus admin (level 1) dan ketua RT (level 3)
		adminTagihan := tagihan.Group("")
		adminTagihan.Use(authMiddleware.RequireLevel(1, 3))
		{
			adminTagihan.POST("", tagihanIuranController.CreateTagihan)
			adminTagihan.PUT("/:id", tagihanIuranController.UpdateTagihan)
			adminTagihan.DELETE("/:id", tagihanIuranController.DeleteTagihan)
			adminTagihan.PUT("/:id/verifikasi", tagihanIuranController.UpdateStatusVerifikasi)
		}
	}
}
