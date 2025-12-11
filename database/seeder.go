package database

import (
	"fmt"
	"log"
	"math/rand"
	"rt-management/models"
	"strconv"
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
		seedRumah,
		seedWarga,
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

	users := []models.User{
		{
			Username:  "admin",
			Password:  string(pass),
			UserNama:  "Administrator",
			UserAlamat: "Jalan Suhat No. 1",
			UserNoTelp: "081111111111",
			UserEmail: "admin@example.com",
			LevelID:   1,
		},
		{
			Username:  "sekretaris",
			Password:  string(pass),
			UserNama:  "Sekretaris Desa",
			UserAlamat: "Jalan Mawar No. 2",
			UserNoTelp: "082222222222",
			UserEmail: "sekretaris@example.com",
			LevelID:   2,
		},
		{
			Username:  "bendahara",
			Password:  string(pass),
			UserNama:  "Bendahara Desa",
			UserAlamat: "Jalan Melati No. 3",
			UserNoTelp: "083333333333",
			UserEmail: "bendahara@example.com",
			LevelID:   3,
		},
		{
			Username:  "pengurus_rt",
			Password:  string(pass),
			UserNama:  "Pengurus RT",
			UserAlamat: "Jalan Kenanga No. 4",
			UserNoTelp: "084444444444",
			UserEmail: "pengurus_rt@example.com",
			LevelID:   4,
		},
		{
			Username:  "pengurus_rw",
			Password:  string(pass),
			UserNama:  "Pengurus RW",
			UserAlamat: "Jalan Anggrek No. 5",
			UserNoTelp: "085555555555",
			UserEmail: "pengurus_rw@example.com",
			LevelID:   5,
		},
		{
			Username:  "warga001",
			Password:  string(pass),
			UserNama:  "Warga Desa 001",
			UserAlamat: "Jalan Cempaka No. 6",
			UserNoTelp: "086666666666",
			UserEmail: "warga001@example.com",
			LevelID:   6,
		},
	}

	return DB.Create(&users).Error
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

func seedRumah() error {
	rumahs := []models.Rumah{}

	for i := 1; i <= 5; i++ {
		rumahs = append(rumahs, models.Rumah{
			RumahAlamat: fmt.Sprintf("Jl. Contoh No.%d", i),
			RumahStatus: "tersedia", // default
		})
	}

	return DB.Create(&rumahs).Error
}

func seedWarga() error {
	var keluarga []models.Keluarga
	var agama []models.Agama
	var pekerjaan []models.Pekerjaan
	var rumah []models.Rumah

	DB.Find(&keluarga)
	DB.Find(&agama)
	DB.Find(&pekerjaan)
	DB.Find(&rumah)

	if len(keluarga) == 0 || len(agama) == 0 || len(pekerjaan) == 0 {
		return fmt.Errorf("reference data missing")
	}

	cities := []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Semarang", "Yogyakarta", "Makassar"}
	var data []models.Warga

	for i := 0; i < 100; i++ {

		// generate nik
		nik := "32"
		for j := 0; j < 14; j++ {
			nik += fmt.Sprintf("%d", rand.Intn(10))
		}

		// generate phone
		phone := "08"
		for j := 0; j < 10; j++ {
			phone += fmt.Sprintf("%d", rand.Intn(10))
		}

		// === BUAT USER OTOMATIS ===
		hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

		user := models.User{
			Username:   "warga000" + strconv.Itoa(i+1),
			Password:   string(hash),
			UserNama:   faker.Name(),
			UserAlamat: faker.Sentence(),
			UserNoTelp: phone,
			UserEmail:  faker.Email(),
			LevelID:    6, // misal level warga
		}

		if err := DB.Create(&user).Error; err != nil {
			return fmt.Errorf("gagal membuat user: %v", err)
		}

		// === BUAT DATA WARGA ===
		warga := models.Warga{
			UserID:            user.UserID, // ✔ uint langsung, BUKAN &user.UserID
			KeluargaID:        keluarga[rand.Intn(len(keluarga))].KeluargaID,
			WargaNama:         user.UserNama,
			WargaNIK:          nik,
			WargaNoTlp:        phone,
			WargaTempatLahir:  cities[rand.Intn(len(cities))],
			WargaTanggalLahir: time.Now().AddDate(-rand.Intn(40)-20, 0, 0),
			WargaJenisKelamin: []string{"L", "P"}[rand.Intn(2)],
			WargaStatusAktif:  []string{"aktif", "nonaktif"}[rand.Intn(2)],
			WargaStatusHidup:  []string{"hidup", "meninggal"}[rand.Intn(2)],
			AgamaID:           agama[rand.Intn(len(agama))].AgamaID,
			PekerjaanID:       pekerjaan[rand.Intn(len(pekerjaan))].PekerjaanID,
			RumahID:           rumah[rand.Intn(len(rumah))].RumahID,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}

		// assign rumah hanya warga hidup & aktif
		if len(rumah) > 0 && warga.WargaStatusAktif == "aktif" && warga.WargaStatusHidup == "hidup" {
			if rand.Intn(100) < 80 {
				warga.RumahID = rumah[rand.Intn(len(rumah))].RumahID
			}
		}

		data = append(data, warga)
	}

	// batch insert warga
	if err := DB.CreateInBatches(&data, 50).Error; err != nil {
		return err
	}

	// update status rumah menjadi ditempati
	for _, w := range data {
		if w.RumahID != 0 {
			DB.Model(&models.Rumah{}).
				Where("rumah_id = ?", w.RumahID).
				Update("rumah_status", "ditempati")
		}
	}

	return nil
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
		data = append(data, models.TagihanIuran{
			TagihanIuran: n,
			WargaID:       1, // default admin
		})
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
	var user []models.User
	DB.Find(&kategori)
	DB.Find(&user)

	var data []models.Produk
	for i := 0; i < 20; i++ {
		data = append(data, models.Produk{
			ProdukNama:       faker.Word(),
			ProdukDeskripsi:  faker.Sentence(),
			ProdukStok:       rand.Intn(100) + 1,
			ProdukHarga:      float64(rand.Intn(100000) + 10000),
			ProdukFoto:       faker.Word() + ".jpg",
			KategoriProdukID: kategori[rand.Intn(len(kategori))].KategoriProdukID,
			UserID:           user[rand.Intn(len(user))].UserID,
		})
	}
	return DB.Create(&data).Error
}

func seedBroadcast() error {
	// buat 5 data broadcast
	for i := 0; i < 5; i++ {
		broadcast := models.Broadcast{
			BroadcastNama:      faker.Sentence(),      // judul/nama broadcast
			BroadcastDeskripsi: faker.Paragraph(),     // deskripsi
			BroadcastFoto:      faker.Word() + ".jpg", // contoh URL foto palsu
			BroadcastDokumen:   faker.Word() + ".pdf", // contoh URL dokumen palsu
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}

		if err := DB.Create(&broadcast).Error; err != nil {
			return err
		}
	}

	return nil
}
