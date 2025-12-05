package database

import (
	"fmt"
	"log"
	"math/rand"
	"rt-management/models"
	"time"

	"github.com/bxcodec/faker/v3"
	"golang.org/x/crypto/bcrypt"
)

func SeedData() error {
	if DB == nil {
		return fmt.Errorf("database connection not initialized")
	}

	log.Println("Starting to seed data...")

	seeders := []func() error{
		seedLevels,
		seedUsers,
		seedAgama,
		seedPekerjaan,
		seedKeluarga,
		seedWarga,
		seedRumah,
		seedKategoriKegiatan,
		seedKegiatan,
		seedKategoriPengeluaran,
		seedPengeluaran,
		seedKategoriPemasukan,
		seedPemasukan,
		seedTagihanIuran,
		seedKategoriProduk,
		seedProduk,
		seedBroadcast,
	}

	for _, fn := range seeders {
		if err := fn(); err != nil {
			return err
		}
	}

	log.Println("Seeder completed!")
	return nil
}

/* --------------------- LEVEL ---------------------- */

func seedLevels() error {
	data := []models.Level{
		{LevelKode: "ADM", LevelNama: "Administrator"},
		{LevelKode: "SRT", LevelNama: "Sekretaris"},
		{LevelKode: "BND", LevelNama: "Bendahara"},
		{LevelKode: "KRT", LevelNama: "Ketua RT"},
		{LevelKode: "KRW", LevelNama: "Ketua RW"},
		{LevelKode: "WRG", LevelNama: "Warga"},
	}

	return DB.Create(&data).Error
}

/* --------------------- USERS ---------------------- */

func seedUsers() error {
	pass, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	data := []models.User{
		{Username: "admin", Password: string(pass), LevelID: 1},
		{Username: "sekretaris", Password: string(pass), LevelID: 2},
		{Username: "bendahara", Password: string(pass), LevelID: 3},
		{Username: "pengurus_rt", Password: string(pass), LevelID: 4},
		{Username: "pengurus_rw", Password: string(pass), LevelID: 5},
		{Username: "warga001", Password: string(pass), LevelID: 6},
	}

	return DB.Create(&data).Error
}

/* --------------------- AGAMA ---------------------- */

func seedAgama() error {
	names := []string{"Islam", "Kristen", "Katolik", "Hindu", "Buddha", "Konghucu"}
	var data []models.Agama
	for _, n := range names {
		data = append(data, models.Agama{AgamaNama: n})
	}
	return DB.Create(&data).Error
}

/* --------------------- PEKERJAAN ---------------------- */

func seedPekerjaan() error {
	names := []string{
		"PNS", "TNI", "Polri", "Karyawan Swasta", "Wiraswasta", "Petani",
		"Nelayan", "Guru", "Dokter", "Perawat", "Pedagang", "Buruh", "Pensiunan",
	}
	var data []models.Pekerjaan
	for _, n := range names {
		data = append(data, models.Pekerjaan{PekerjaanNama: n})
	}
	return DB.Create(&data).Error
}

/* --------------------- KELUARGA ---------------------- */

func seedKeluarga() error {
	var data []models.Keluarga
	for i := 0; i < 20; i++ {
		status := "aktif"
		if i%10 == 0 {
			status = "nonaktif"
		}
		data = append(data, models.Keluarga{
			KeluargaNama:   faker.LastName() + " Family",
			KeluargaStatus: status,
		})
	}
	return DB.Create(&data).Error
}

/* --------------------- WARGA ---------------------- */

func seedWarga() error {
	var keluarga []models.Keluarga
	var agama []models.Agama
	var pekerjaan []models.Pekerjaan

	DB.Find(&keluarga)
	DB.Find(&agama)
	DB.Find(&pekerjaan)

	if len(keluarga) == 0 || len(agama) == 0 || len(pekerjaan) == 0 {
		return fmt.Errorf("reference data missing")
	}

	cities := []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Semarang"}

	var data []models.Warga

	for i := 0; i < 100; i++ {
		nik := "32"
		for j := 0; j < 14; j++ {
			nik += fmt.Sprintf("%d", rand.Intn(10))
		}

		data = append(data, models.Warga{
			KeluargaID:        keluarga[rand.Intn(len(keluarga))].KeluargaID,
			WargaNama:         faker.Name(),
			WargaNIK:          nik,
			WargaNoTlp:        "08" + fmt.Sprintf("%010d", rand.Intn(1000000000)),
			WargaTempatLahir:  cities[rand.Intn(len(cities))],
			WargaTanggalLahir: time.Now().AddDate(-rand.Intn(40)-20, 0, 0),
			WargaJenisKelamin: []string{"L", "P"}[rand.Intn(2)],
			WargaStatusAktif:  []string{"aktif", "nonaktif"}[rand.Intn(2)],
			WargaStatusHidup:  []string{"hidup", "meninggal"}[rand.Intn(2)],
			AgamaID:           agama[rand.Intn(len(agama))].AgamaID,
			PekerjaanID:       pekerjaan[rand.Intn(len(pekerjaan))].PekerjaanID,
		})
	}

	return DB.CreateInBatches(&data, 50).Error
}

/* --------------------- RUMAH ---------------------- */

func seedRumah() error {
	rumahs := []models.Rumah{}

	for i := 1; i <= 5; i++ {
		rumahs = append(rumahs, models.Rumah{
			RumahAlamat: fmt.Sprintf("Jl. Contoh No.%d", i),
			RumahStatus: "ditempati",
			WargaID:     uint(i),
		})
	}

	return DB.Create(&rumahs).Error
}




/* --------------------- MASTER DATA LAIN ---------------------- */

func seedKategoriKegiatan() error {
	names := []string{"Gotong Royong", "Rapat RT", "Hari Besar", "Olahraga", "Kesehatan"}
	var data []models.KategoriKegiatan
	for _, n := range names {
		data = append(data, models.KategoriKegiatan{KategoriKegiatanNama: n})
	}
	return DB.Create(&data).Error
}

func seedKegiatan() error {
	var kategori []models.KategoriKegiatan
	DB.Find(&kategori)

	places := []string{"Balai RW", "Lapangan", "Masjid", "Sekolah"}

	var data []models.Kegiatan
	for i := 0; i < 30; i++ {
		data = append(data, models.Kegiatan{
			KegiatanNama:       faker.Sentence(),
			KategoriKegiatanID: kategori[rand.Intn(len(kategori))].KategoriKegiatanID,
			KegiatanTanggal:    time.Now().AddDate(0, 0, rand.Intn(30)),
			KegiatanLokasi:     places[rand.Intn(len(places))],
			KegiatanPJ:         faker.Name(),
			KegiatanDeskripsi:  faker.Paragraph(),
		})
	}
	return DB.Create(&data).Error
}

func seedKategoriPengeluaran() error {
	names := []string{"Listrik", "Air", "Kebersihan", "Kegiatan RT", "Administrasi"}
	var data []models.KategoriPengeluaran
	for _, n := range names {
		data = append(data, models.KategoriPengeluaran{KategoriPengeluaranNama: n})
	}
	return DB.Create(&data).Error
}

func seedPengeluaran() error {
	var kategori []models.KategoriPengeluaran
	DB.Find(&kategori)

	var data []models.Pengeluaran
	for i := 0; i < 50; i++ {
		data = append(data, models.Pengeluaran{
			KategoriPengeluaranID: kategori[rand.Intn(len(kategori))].KategoriPengeluaranID,
			PengeluaranNama:       faker.Word(),
			PengeluaranTanggal:    time.Now().AddDate(0, 0, -rand.Intn(100)),
			PengeluaranNominal:    float64(rand.Intn(800000) + 100000),
			PengeluaranBukti:      faker.Word() + ".jpg",
		})
	}
	return DB.Create(&data).Error
}

func seedKategoriPemasukan() error {
	names := []string{"Iuran Warga", "Sumbangan", "Dana Desa", "Lain-lain"}
	var data []models.KategoriPemasukan
	for _, n := range names {
		data = append(data, models.KategoriPemasukan{KategoriPemasukanNama: n})
	}
	return DB.Create(&data).Error
}

func seedPemasukan() error {
	var kategori []models.KategoriPemasukan
	DB.Find(&kategori)

	var data []models.Pemasukan
	for i := 0; i < 40; i++ {
		data = append(data, models.Pemasukan{
			KategoriPemasukanID: kategori[rand.Intn(len(kategori))].KategoriPemasukanID,
			PemasukanNama:       faker.Word(),
			PemasukanTanggal:    time.Now().AddDate(0, 0, -rand.Intn(60)),
			PemasukanNominal:    float64(rand.Intn(1500000) + 500000),
		})
	}
	return DB.Create(&data).Error
}

func seedTagihanIuran() error {
	names := []string{"Iuran Kebersihan", "Iuran Keamanan", "Iuran Kegiatan", "Iuran Sampah"}
	var data []models.TagihanIuran
	for _, n := range names {
		data = append(data, models.TagihanIuran{TagihanIuran: n})
	}
	return DB.Create(&data).Error
}

/* --------------------- PRODUK (E-COMMERCE) ---------------------- */

func seedKategoriProduk() error {
	names := []string{"Makanan", "Minuman", "Peralatan Rumah Tangga", "Elektronik"}
	var data []models.KategoriProduk
	for _, n := range names {
		data = append(data, models.KategoriProduk{KategoriProdukNama: n})
	}
	return DB.Create(&data).Error
}

func seedProduk() error {
	var kategori []models.KategoriProduk
	DB.Find(&kategori)

	var data []models.Produk
	for i := 0; i < 20; i++ {
		data = append(data, models.Produk{
			ProdukNama:       faker.Word(),
			ProdukDeskripsi:  faker.Sentence(),
			ProdukStok:       rand.Intn(100) + 1,
			ProdukHarga:      float64(rand.Intn(100000) + 10000),
			ProdukFoto:       faker.Word() + ".jpg",
			KategoriProdukID: kategori[rand.Intn(len(kategori))].KategoriProdukID,
		})
	}
	return DB.Create(&data).Error
}

func seedBroadcast() error {
	// buat 5 data broadcast
	for i := 0; i < 5; i++ {
		broadcast := models.Broadcast{
			BroadcastNama:      faker.Sentence(),   // judul/nama broadcast
			BroadcastDeskripsi: faker.Paragraph(), // deskripsi
			BroadcastFoto:      faker.Word() + ".jpg",       // contoh URL foto palsu
			BroadcastDokumen:   faker.Word() + ".pdf",       // contoh URL dokumen palsu
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := DB.Create(&broadcast).Error; err != nil {
			return err
		}
	}

	return nil
}