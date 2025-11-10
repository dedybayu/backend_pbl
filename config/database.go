// config/database.go
package config

import (
	"log"
	"rt-management/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func InitDB() (*gorm.DB, error) {
	dsn := "dedybayu:dbsn@tcp(127.0.0.1:3306)/rt_management?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Cek koneksi database
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		return nil, err
	}

	log.Println("✅ Connected to database successfully!")

	// Hanya auto migrate untuk development jika diperlukan
	// Untuk production, sebaiknya tidak menggunakan AutoMigrate
	// jika struktur tabel sudah fix
	if shouldAutoMigrate() {
		err = db.AutoMigrate(
			&models.Level{},
			&models.User{},
			&models.Keluarga{},
			&models.Agama{},
			&models.Pekerjaan{},
			&models.Warga{},
			&models.Rumah{},
			&models.KategoriKegiatan{},
			&models.Kegiatan{},
			&models.Broadcast{},
			&models.MutasiKeluarga{},
			&models.KategoriPengeluaran{},
			&models.Pengeluaran{},
			&models.KategoriPemasukan{},
			&models.Pemasukan{},
			&models.TagihanIuran{},
			&models.KategoriProduk{},
			&models.Produk{},
		)

		if err != nil {
			log.Printf("Warning: AutoMigrate error: %v", err)
		} else {
			log.Println("✅ Database tables migrated successfully!")
		}
	}

	return db, nil
}

func shouldAutoMigrate() bool {
	// Tambahkan logic sesuai kebutuhan
	// Return false untuk production, true untuk development
	return false // Ubah ke true jika ingin auto migrate
}