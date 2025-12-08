package main

import (
	"flag"
	"log"
	"os"
	"rt-management/controllers"
	"rt-management/database"
	"rt-management/middleware"
	"rt-management/routes"
	"rt-management/utils"

	"github.com/gin-gonic/gin"

	swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "rt-management/docs" // ← harus sesuai module path projek
)

func main() {

	// FLAGS
	migrate := flag.Bool("migrate", false, "Run database migration only")
	seed := flag.Bool("seed", false, "Run database seed only")
	migrateSeed := flag.Bool("migrate-seed", false, "Run migration and then seed")
	flag.Parse()

	// CONNECT DB
	dbConfig := database.DatabaseConfig{
		Host:     "localhost",
		Port:     "3306",
		User:     "dedybayu",
		Password: "dbsn",
		DBName:   "rt_management",
	}

	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Fatal("❌ DB connection error:", err)
	}

	// RUN MIGRATION ONLY
	if *migrate {
		log.Println("🔄 Running migration...")
		if err := database.CleanMigrate(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// RUN SEED ONLY
	if *seed {
		log.Println("🌱 Running seeder...")
		if err := database.SeedData(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// RUN MIGRATE + SEED
	if *migrateSeed {
		log.Println("🔄 Running migration and seeding...")
		if err := database.CleanMigrate(); err != nil {
			log.Fatal(err)
		}
		if err := database.SeedData(); err != nil {
			log.Fatal(err)
		}
		return
	}

	// =============== NORMAL MODE (START API SERVER) ===============

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "rt-management-secret-key"
	}

	jwtUtils := utils.NewJWTUtils(jwtSecret)

	// CONTROLLERS
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

	// MIDDLEWARE
	authMiddleware := middleware.NewAuthMiddleware(jwtUtils)

	// ROUTER
	r := gin.Default()
	r.Use(middleware.SanitizationMiddleware())

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	routeConfig := &routes.RouteConfig{
		AuthController:                authController,
		UserController:                userController,
		LevelController:               levelController,
		KeluargaController:            keluargaController,
		WargaController:               wargaController,
		RumahController:               rumahController,
		KegiatanController:            kegiatanController,
		BroadcastController:           broadcastController,
		MutasiKeluargaController:      mutasiKeluargaController,
		KategoriPengeluaranController: kategoriPengeluaranController,
		PengeluaranController:         pengeluaranController,
		KategoriPemasukanController:   kategoriPemasukanController,
		PemasukanController:           pemasukanController,
		TagihanIuranController:        tagihanIuranController,
		KategoriProdukController:      kategoriProdukController,
		ProdukController:              produkController,
		ProfileController:             profileController,
		AuthMiddleware:                authMiddleware,
	}

	routes.SetupRoutes(r, routeConfig)
	// Swagger UI: http://localhost:8080/swagger/index.html
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))


	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server running at :%s", port)
	r.Run(":" + port)
}
