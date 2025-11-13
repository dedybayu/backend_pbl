// main.go
package main

import (
	"log"
	"os"
	"rt-management/config"
	"rt-management/controllers"
	"rt-management/middleware"
	"rt-management/routes"
	"rt-management/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Initialize database
	db, err := config.InitDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rt-management-secret-key-2024" // For development only
		log.Println("Warning: JWT_SECRET not set, using fallback key")
	}

	// Initialize JWT utils
	jwtUtils := utils.NewJWTUtils(jwtSecret)

	// Initialize controllers
	authController := controllers.NewAuthController(db, jwtUtils)
	userController := controllers.NewUserController(db)
	levelController := controllers.NewLevelController(db)
	keluargaController := controllers.NewKeluargaController(db)
	wargaController := controllers.NewWargaController(db)
	rumahController := controllers.NewRumahController(db)
	kegiatanController := controllers.NewKegiatanController(db)
	mutasiKeluargaController := controllers.NewMutasiKeluargaController(db)
	broadcastController := controllers.NewBroadcastController(db)
	kategoriPengeluaranController := controllers.NewKategoriPengeluaranController(db)
	pengeluaranController := controllers.NewPengeluaranController(db)
	kategoriPemasukanController := controllers.NewKategoriPemasukanController(db)
	pemasukanController := controllers.NewPemasukanController(db)
	tagihanIuranController := controllers.NewTagihanIuranController(db)
	kategoriProdukController := controllers.NewKategoriProdukController(db)
	ProfileController := controllers.NewProfileController(db)
	// fileImageController := controllers.NewFileImageController()
	produkController := controllers.NewProdukController(db)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtils)

	// Setup router
	router := gin.Default()

	// ✅ ADD SANITIZATION MIDDLEWARE FIRST (Security Enhancement)
	router.Use(middleware.SanitizationMiddleware())

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		
		c.Next()
	})

	// Setup routes configuration
	routeConfig := &routes.RouteConfig{
		AuthController:     authController,
		UserController:     userController,
		LevelController:    levelController,
		KeluargaController: keluargaController,
		WargaController:    wargaController,
		RumahController:    rumahController,
		KegiatanController: kegiatanController,
		BroadcastController: broadcastController,
		MutasiKeluargaController: mutasiKeluargaController,
		KategoriPengeluaranController: kategoriPengeluaranController,
		PengeluaranController: pengeluaranController,
		KategoriPemasukanController: kategoriPemasukanController,
		PemasukanController: pemasukanController,
		TagihanIuranController: tagihanIuranController,
		KategoriProdukController: kategoriProdukController,
		ProdukController: produkController,
		ProfileController: ProfileController,
		// FileImageController: fileImageController,
		AuthMiddleware:     authMiddleware,
	}

	// Setup all routes
	routes.SetupRoutes(router, routeConfig)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Printf("🚀 Server starting on :%s", port)
	log.Printf("🔒 Security features: SQL Injection Protection, Input Sanitization, JWT Authentication")
	
	if err := router.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}