// routes/routes.go
package routes

import (
	"rt-management/controllers"
	"rt-management/middleware"

	"github.com/gin-gonic/gin"
)

type RouteConfig struct {
	AuthController     *controllers.AuthController
	UserController     *controllers.UserController
	LevelController    *controllers.LevelController
	KeluargaController *controllers.KeluargaController
	WargaController    *controllers.WargaController
	RumahController    *controllers.RumahController
	KegiatanController *controllers.KegiatanController
	AuthMiddleware     *middleware.AuthMiddleware
}

func SetupRoutes(router *gin.Engine, config *RouteConfig) {
	// Health check route
	router.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "OK",
			"message": "RT Management API is running",
			"security": "SQL Injection Protection Enabled",
		})
	})

	// API documentation route
	router.GET("/api/docs", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "RT Management API Documentation",
			"endpoints": map[string]string{
				"auth":     "/auth/login, /auth/logout, /auth/profile",
				"users":    "/api/users, /api/users/:id",
				"levels":   "/api/levels, /api/levels/:id",
				"families": "/api/keluarga, /api/keluarga/:id",
			},
			"security": "JWT Authentication & SQL Injection Protection Active",
		})
	})

	// Setup auth routes
	SetupAuthRoutes(router, config.AuthController, config.AuthMiddleware)

	// Protected API routes
	api := router.Group("/api")
	api.Use(config.AuthMiddleware.Auth()) // Semua endpoint di /api butuh auth
	{
		// Setup level routes
		SetupLevelRoutes(api, config.LevelController, config.AuthMiddleware)

		// Setup user routes
		SetupUserRoutes(api, config.UserController, config.AuthMiddleware)

		// Setup keluarga routes
		SetupKeluargaRoutes(api, config.KeluargaController, config.AuthMiddleware)

		// Setup warga routes
		SetupWargaRoutes(api, config.WargaController, config.AuthMiddleware)

		// Setup rumah routes
		SetupRumahRoutes(api, config.RumahController, config.AuthMiddleware)

		// Setup kegiatan routes
		SetupKegiatanRoutes(api, config.KegiatanController, config.AuthMiddleware)
	}
}