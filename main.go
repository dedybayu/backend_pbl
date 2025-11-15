package main

import (
	"flag"
	"log"
	"os"
	"rt-management/config"
	"rt-management/controllers"
	"rt-management/database"
	"rt-management/middleware"
	"rt-management/routes"
	"rt-management/utils"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// ----------------------------------------------------------------------
	// 1. FLAG OPSIONAL UNTUK MIGRATE & SEED
	// ----------------------------------------------------------------------
	migrate := flag.Bool("migrate", false, "Run database migration")
	seed := flag.Bool("seed", false, "Run database seeder")
	flag.Parse()

	// ----------------------------------------------------------------------
	// 2. INIT DATABASE
	// ----------------------------------------------------------------------
	dbConfig := database.DatabaseConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "dedybayu",
		Password: "dbsn",
		DBName:   "rt_management",
	}

	if err := database.InitDB(dbConfig); err != nil {
		log.Fatal("❌ Failed to connect DB:", err)
	}

	// ----------------------------------------------------------------------
	// 3. JALANKAN MIGRATION JIKA DIMINTA
	// ----------------------------------------------------------------------
	if *migrate {
		log.Println("🔄 Running MIGRATION...")
		if err := database.DropTables(); err != nil {
			log.Fatal("❌ Failed DropTables:", err)
		}
		if err := database.Migrate(); err != nil {
			log.Fatal("❌ Failed Migrate:", err)
		}
		log.Println("✅ Migration complete")
	}

	// ----------------------------------------------------------------------
	// 4. JALANKAN SEEDER JIKA DIMINTA
	// ----------------------------------------------------------------------
	if *seed {
		log.Println("🌱 Running SEEDER...")
		if err := database.SeedData(); err != nil {
			log.Fatal("❌ Failed Seeder:", err)
		}
		log.Println("✅ Seeding complete")
	}

	// ----------------------------------------------------------------------
	// 5. JIKA HANYA MIGRATE/SEED, STOP. JANGAN JALANIN SERVER
	// ----------------------------------------------------------------------
	if *migrate || *seed {
		return
	}

	// ----------------------------------------------------------------------
	// 6. LOAD ENV & MULAI SERVER SEPERTI BIASA
	// ----------------------------------------------------------------------
	if err := godotenv.Load(); err != nil {
		log.Println("⚠️ No .env file found")
	}

	db, err := config.InitDB()
	if err != nil {
		log.Fatal("❌ Failed to connect DB:", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rt-management-secret-key-2024"
	}

	jwtUtils := utils.NewJWTUtils(jwtSecret)

	// Controllers
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
	produkController := controllers.NewProdukController(db)
	profileController := controllers.NewProfileController(db)

	// Middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtUtils)

	// Router
	router := gin.Default()
	router.Use(middleware.SanitizationMiddleware())

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

	// Router config
	routeConfig := &routes.RouteConfig{
		AuthController:               authController,
		UserController:               userController,
		LevelController:              levelController,
		KeluargaController:           keluargaController,
		WargaController:              wargaController,
		RumahController:              rumahController,
		KegiatanController:           kegiatanController,
		BroadcastController:          broadcastController,
		MutasiKeluargaController:     mutasiKeluargaController,
		KategoriPengeluaranController: kategoriPengeluaranController,
		PengeluaranController:        pengeluaranController,
		KategoriPemasukanController:  kategoriPemasukanController,
		PemasukanController:          pemasukanController,
		TagihanIuranController:       tagihanIuranController,
		KategoriProdukController:     kategoriProdukController,
		ProdukController:             produkController,
		ProfileController:            profileController,
		AuthMiddleware:               authMiddleware,
	}

	routes.SetupRoutes(router, routeConfig)

	// Start Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running at :%s", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Server error:", err)
	}
}
