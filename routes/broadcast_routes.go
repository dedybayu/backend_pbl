// routes/broadcast_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupBroadcastRoutes(api *gin.RouterGroup, broadcastController *controllers.BroadcastController, authMiddleware *middleware.AuthMiddleware) {
	broadcast := api.Group("/broadcast")
	{
		// Public routes (butuh auth)
		broadcast.GET("", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetAllBroadcast)
		broadcast.GET("/terbaru", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetBroadcastTerbaru)
		broadcast.GET("/statistik", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetStatistikBroadcast)
		broadcast.GET("/search", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.SearchBroadcast)
		broadcast.GET("/:id", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetBroadcastByID)
		broadcast.GET("/dokumen/:filename", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetBroadcastDokumen)
		broadcast.GET("/image/:filename", authMiddleware.RequireLevel(1, 2, 3, 4, 5, 6), broadcastController.GetBroadcastFoto)
		
		// Admin only routes
		adminBroadcast := broadcast.Group("")
		adminBroadcast.Use(authMiddleware.RequireLevel(1, 2, 3, 4, 5))
		{
			adminBroadcast.POST("", broadcastController.CreateBroadcast)
			adminBroadcast.PUT("/:id", broadcastController.UpdateBroadcast)
			adminBroadcast.DELETE("/:id", broadcastController.DeleteBroadcast)
		}
	}
}