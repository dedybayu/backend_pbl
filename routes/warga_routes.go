// routes/warga_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupWargaRoutes(api *gin.RouterGroup, wargaController *controllers.WargaController, authMiddleware *middleware.AuthMiddleware) {
	warga := api.Group("/warga")
	{
		// Public routes (butuh auth)
		warga.GET("", authMiddleware.RequireLevel(1, 2), wargaController.GetAllWarga)
		warga.GET("/total", authMiddleware.RequireLevel(1, 2), wargaController.GetTotalWarga)
		warga.GET("/stats", authMiddleware.RequireLevel(1, 2), wargaController.GetWargaStats)
		warga.GET("/search", authMiddleware.RequireLevel(1, 2), wargaController.SearchWarga)
		warga.GET("/keluarga/:keluarga_id", authMiddleware.RequireLevel(1, 2), wargaController.GetWargaByKeluarga)
		warga.GET("/:id", authMiddleware.RequireLevel(1, 2), wargaController.GetWargaByID)
		
		// Admin only routes
		adminWarga := warga.Group("")
		adminWarga.Use(authMiddleware.RequireLevel(1))
		{
			adminWarga.POST("", wargaController.CreateWarga)
			adminWarga.PUT("/:id", wargaController.UpdateWarga)
			adminWarga.DELETE("/:id", wargaController.DeleteWarga)
		}
	}
}