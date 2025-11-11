// package routes

// import (
// 	"rt-management/controllers"

// 	"github.com/gin-gonic/gin"
// 	"gorm.io/gorm"
// )

// func SetupRumahRoutes(router *gin.Engine, db *gorm.DB) {
// 	rumahController := controllers.NewRumahController(db)

// 	rumahRoutes := router.Group("/api/rumah")
// 	{
// 		rumahRoutes.POST("/", rumahController.CreateRumah)
// 		rumahRoutes.GET("/", rumahController.GetAllRumah)
// 		rumahRoutes.GET("/:id", rumahController.GetRumahByID)
// 		rumahRoutes.PUT("/:id", rumahController.UpdateRumah)
// 		rumahRoutes.DELETE("/:id", rumahController.DeleteRumah)
// 		rumahRoutes.GET("/status/:status", rumahController.GetRumahByStatus)
// 	}
// }


// routes/warga_routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRumahRoutes(api *gin.RouterGroup, rumahController *controllers.RumahController, authMiddleware *middleware.AuthMiddleware) {
	warga := api.Group("/rumah")
	{
		// Public routes (butuh auth)
		warga.GET("", authMiddleware.RequireLevel(1, 2), rumahController.GetAllRumah)
		warga.GET("/:id", authMiddleware.RequireLevel(1, 2), rumahController.GetRumahByID)
		
		// Admin only routes
		adminWarga := warga.Group("")
		adminWarga.Use(authMiddleware.RequireLevel(1))
		{
			adminWarga.POST("", rumahController.CreateRumah)
			adminWarga.PUT("/:id", rumahController.UpdateRumah)
			adminWarga.DELETE("/:id", rumahController.DeleteRumah)
		}
	}
}