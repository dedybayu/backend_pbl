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

	tables := []interface{}{
		&models.Level{},
		&models.Agama{},
		&models.Pekerjaan{},
		&models.Keluarga{},
		&models.KategoriKegiatan{},
		&models.KategoriPengeluaran{},
		&models.KategoriPemasukan{},
		&models.TagihanIuran{},
		&models.KategoriProduk{},
		&models.User{},
		&models.Warga{},
		&models.Rumah{},
		&models.Kegiatan{},
		&models.Broadcast{},
		&models.MutasiKeluarga{},
		&models.Pengeluaran{},
		&models.Pemasukan{},
		&models.Produk{},
	}

	for _, t := range tables {
		if err := DB.AutoMigrate(t); err != nil {
			return fmt.Errorf("migrate failed %T: %v", t, err)
		}
		log.Printf("✓ Migrated: %T", t)
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

	for _, t := range tables {
		_ = DB.Migrator().DropTable(t)
		log.Printf("✓ Dropped: %T", t)
	}

	DB.Exec("SET FOREIGN_KEY_CHECKS = 1")

	log.Println("🗑 All tables dropped")
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
