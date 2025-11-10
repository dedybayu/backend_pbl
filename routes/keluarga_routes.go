// routes/keluarga_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

// routes/keluarga_routes.go (update)
func SetupKeluargaRoutes(api *gin.RouterGroup, keluargaController *controllers.KeluargaController, authMiddleware *middleware.AuthMiddleware) {
	keluarga := api.Group("/keluarga")
	{
		// Public routes (butuh auth)
		keluarga.GET("", authMiddleware.RequireLevel(1, 2), keluargaController.GetAllKeluarga)
		keluarga.GET("/aktif", authMiddleware.RequireLevel(1, 2), keluargaController.GetKeluargaAktif) // ✅ NEW
		keluarga.GET("/stats", authMiddleware.RequireLevel(1, 2), keluargaController.GetKeluargaStats)
		keluarga.GET("/search", authMiddleware.RequireLevel(1, 2), keluargaController.SearchKeluarga)
		keluarga.GET("/:id", authMiddleware.RequireLevel(1, 2), keluargaController.GetKeluargaByID)
		keluarga.GET("/:id/details", authMiddleware.RequireLevel(1, 2), keluargaController.GetKeluargaWithDetails)
		
		// Admin only routes
		adminKeluarga := keluarga.Group("")
		adminKeluarga.Use(authMiddleware.RequireLevel(1))
		{
			adminKeluarga.POST("", keluargaController.CreateKeluarga)
			adminKeluarga.PUT("/:id", keluargaController.UpdateKeluarga)
			adminKeluarga.DELETE("/:id", keluargaController.DeleteKeluarga)
		}
	}
}