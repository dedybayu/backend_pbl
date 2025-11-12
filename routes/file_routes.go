// // routes/file_routes.go
package routes

// import (
// 	"rt-management/controllers"
// 	"rt-management/middleware"

// 	"github.com/gin-gonic/gin"
// )

// func SetupFileRoutes(api *gin.RouterGroup, fileImageController *controllers.FileImageController, authMiddleware *middleware.AuthMiddleware) {
// 	// Public routes untuk akses gambar
// 	api.GET("/images/:path", fileImageController.GetGambarByPath)

// 	// Protected routes untuk upload dan management gambar
// 	file := api.Group("/files")
// 	{
// 		// Public routes (butuh auth)
// 		file.GET("/images/:type", authMiddleware.RequireLevel(1, 2), fileImageController.GetGambarByType)
// 		file.GET("/statistik", authMiddleware.RequireLevel(1, 2), fileImageController.GetStatistikGambar)
		
// 		// Admin only routes
// 		adminFile := file.Group("")
// 		adminFile.Use(authMiddleware.RequireLevel(1))
// 		{
// 			adminFile.POST("/upload", fileImageController.UploadGambar)
// 			adminFile.POST("/rename", fileImageController.RenameGambar) // Tambah endpoint rename
// 			adminFile.DELETE("/images/:path", fileImageController.DeleteGambar)
// 		}
// 	}
// }