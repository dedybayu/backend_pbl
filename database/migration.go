package database

import (
	"fmt"
	"log"
	"rt-management/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

func InitDB(config DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	DB = db
	log.Println("✅ Connected to database")
	return db, nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// ✅ URUTAN MIGRATION YANG BENAR:
	// 1. Tabel master/induk (tanpa foreign key)
	// 2. Tabel yang punya foreign key ke tabel #1
	// 3. Tabel yang punya foreign key ke tabel #2, dst

	tables := []interface{}{
		// ===== TABEL MASTER/INDUK (tanpa foreign key) =====
		&models.Level{},
		&models.Agama{},
		&models.Pekerjaan{},
		&models.Keluarga{},
		&models.KategoriKegiatan{},
		&models.KategoriPengeluaran{},
		&models.KategoriPemasukan{},
		&models.TagihanIuran{},
		&models.KategoriProduk{},
		&models.Rumah{}, // ✅ RUMAH DIPINDAHKAN KE SINI (tidak punya foreign key)

		// ===== TABEL YANG BUTUH TABEL MASTER =====
		&models.User{}, // butuh Level

		// ===== TABEL YANG BUTUH USER & MASTER LAIN =====
		&models.Warga{},          // butuh Keluarga, Agama, Pekerjaan, Rumah, User? (cek ini)
		&models.Produk{},         // butuh KategoriProduk, User
		&models.Kegiatan{},       // butuh KategoriKegiatan
		&models.Pengeluaran{},    // butuh KategoriPengeluaran
		&models.Pemasukan{},      // butuh KategoriPemasukan
		&models.Broadcast{},      // tidak punya foreign key
		&models.MutasiKeluarga{}, // butuh Keluarga
		&models.PesanWarga{},     // butuh User

		// ===== TABEL E-COMMERCE (butuh User & Produk) =====
		&models.Keranjang{},       // butuh User, Produk
		&models.Transaksi{},       // butuh User
		&models.DetailTransaksi{}, // butuh Transaksi, Produk
	}

	log.Println("🔄 Starting database migration...")

	for _, t := range tables {
		if err := DB.AutoMigrate(t); err != nil {
			log.Printf("❌ Migrate failed for %T: %v", t, err)
			// Jangan langsung return, lanjutkan untuk debug
		} else {
			log.Printf("✓ Migrated: %T", t)
		}
	}

	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("✅ Migration complete")
	return nil
}

func DropTables() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	DB.Exec("SET FOREIGN_KEY_CHECKS = 0")

	// ✅ URUTAN DROP YANG BENAR (berlawanan dengan migration):
	// Child/tabel dependen dihapus dulu, parent belakangan

	tables := []interface{}{
		// ===== TABEL CHILD/DEPENDEN (dihapus dulu) =====
		&models.DetailTransaksi{}, // butuh Transaksi, Produk
		&models.Transaksi{},       // butuh User
		&models.Keranjang{},       // butuh User, Produk

		// ===== TABEL CHILD LAINNYA =====
		&models.Produk{},         // butuh KategoriProduk, User
		&models.Pemasukan{},      // butuh KategoriPemasukan
		&models.Pengeluaran{},    // butuh KategoriPengeluaran
		&models.MutasiKeluarga{}, // butuh Keluarga
		&models.Broadcast{},
		&models.Kegiatan{},   // butuh KategoriKegiatan
		&models.Warga{},      // butuh Keluarga, Agama, Pekerjaan, Rumah
		&models.PesanWarga{}, // butuh User

		// ===== TABEL DEPENDEN =====
		&models.User{},  // butuh Level
		&models.Rumah{}, // tidak punya foreign key

		// ===== TABEL MASTER/PARENT (dihapus belakangan) =====
		&models.KategoriProduk{},
		&models.TagihanIuran{},
		&models.KategoriPemasukan{},
		&models.KategoriPengeluaran{},
		&models.KategoriKegiatan{},
		&models.Pekerjaan{},
		&models.Agama{},
		&models.Keluarga{},
		&models.Level{},
	}

	log.Println("🗑 Starting to drop tables...")

	for _, t := range tables {
		// Cek apakah tabel exist sebelum drop
		if DB.Migrator().HasTable(t) {
			if err := DB.Migrator().DropTable(t); err != nil {
				log.Printf("⚠️ Failed to drop %T: %v", t, err)
			} else {
				log.Printf("✓ Dropped: %T", t)
			}
		} else {
			log.Printf("ℹ️ Table %T doesn't exist", t)
		}
	}

	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("✅ All tables dropped")
	return nil
}

func CleanMigrate() error {
	log.Println("🔄 Clean migration started...")

	if err := DropTables(); err != nil {
		return err
	}

	if err := Migrate(); err != nil {
		return err
	}

	log.Println("✨ Clean migration done")
	return nil
}
