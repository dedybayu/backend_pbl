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

func InitDB(config DatabaseConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.User, config.Password, config.Host, config.Port, config.DBName)

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true, // Nonaktifkan sementara constraint saat migrasi
	})
	if err != nil {
		return err
	}

	DB = db
	log.Println("Connected to database successfully")
	return nil
}

func Migrate() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Nonaktifkan foreign key checks sementara
	if err := DB.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		return err
	}

	// Migrasi tabel master terlebih dahulu (tanpa foreign key)
	masterTables := []interface{}{
		&models.Level{},
		&models.Agama{},
		&models.Pekerjaan{},
		&models.Keluarga{},
		&models.KategoriKegiatan{},
		&models.KategoriPengeluaran{},
		&models.KategoriPemasukan{},
		&models.TagihanIuran{},
		&models.KategoriProduk{},
	}

	for _, table := range masterTables {
		if err := DB.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to migrate table %T: %v", table, err)
		}
		log.Printf("Migrated master table: %T", table)
	}

	// Migrasi tabel dengan foreign key
	relationTables := []interface{}{
		&models.User{},        // butuh Level
		&models.Warga{},       // butuh Keluarga, Agama, Pekerjaan
		&models.Rumah{},       // butuh Warga
		&models.Kegiatan{},    // butuh KategoriKegiatan
		&models.Broadcast{},   // tidak ada foreign key
		&models.MutasiKeluarga{}, // butuh Keluarga
		&models.Pengeluaran{}, // butuh KategoriPengeluaran
		&models.Pemasukan{},   // butuh KategoriPemasukan
		&models.Produk{},      // butuh KategoriProduk
	}

	for _, table := range relationTables {
		if err := DB.AutoMigrate(table); err != nil {
			return fmt.Errorf("failed to migrate table %T: %v", table, err)
		}
		log.Printf("Migrated relation table: %T", table)
	}

	// Aktifkan kembali foreign key checks
	if err := DB.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
		return err
	}

	log.Println("Database migration completed successfully")
	return nil
}

func DropTables() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	// Nonaktifkan foreign key checks
	if err := DB.Exec("SET FOREIGN_KEY_CHECKS=0").Error; err != nil {
		return err
	}

	// Daftar tabel untuk di-drop (urutan terbalik dari migration)
	tables := []interface{}{
		&models.Produk{},
		&models.Pemasukan{},
		&models.Pengeluaran{},
		&models.MutasiKeluarga{},
		&models.Broadcast{},
		&models.Kegiatan{},
		&models.Rumah{},
		&models.Warga{},
		&models.User{},
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

	for _, table := range tables {
		if err := DB.Migrator().DropTable(table); err != nil {
			return fmt.Errorf("failed to drop table %T: %v", table, err)
		}
		log.Printf("Dropped table: %T", table)
	}

	// Aktifkan kembali foreign key checks
	if err := DB.Exec("SET FOREIGN_KEY_CHECKS=1").Error; err != nil {
		return err
	}

	log.Println("All tables dropped successfully")
	return nil
}