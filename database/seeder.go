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

	// Seed dalam urutan yang benar berdasarkan foreign key dependencies
	seedFunctions := []func() error{
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
	}

	for _, seedFunc := range seedFunctions {
		if err := seedFunc(); err != nil {
			return fmt.Errorf("seeding failed: %v", err)
		}
	}

	log.Println("Data seeding completed successfully")
	return nil
}

func seedLevels() error {
	levels := []models.Level{
		{LevelKode: "ADM", LevelNama: "Administrator"},
		{LevelKode: "SRT", LevelNama: "Sekretaris"},
		{LevelKode: "BND", LevelNama: "Bendahara"},
		{LevelKode: "KRT", LevelNama: "Ketua RT"},
		{LevelKode: "KRW", LevelNama: "Ketua RW"},
		{LevelKode: "WRG", LevelNama: "Warga"},
	}

	for i := range levels {
		levels[i].CreatedAt = time.Now()
		levels[i].UpdatedAt = time.Now()
	}

	if err := DB.Create(&levels).Error; err != nil {
		return fmt.Errorf("failed to seed levels: %v", err)
	}
	log.Println("Seeded levels")
	return nil
}

func seedUsers() error {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)

	users := []models.User{
		{
			Username: "admin",
			Password: string(hashedPassword),
			LevelID:  1,
		},
		{
			Username: "pengurus_rt",
			Password: string(hashedPassword),
			LevelID:  2,
		},
		{
			Username: "warga001",
			Password: string(hashedPassword),
			LevelID:  3,
		},
	}

	for i := range users {
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()
	}

	if err := DB.Create(&users).Error; err != nil {
		return fmt.Errorf("failed to seed users: %v", err)
	}
	log.Println("Seeded users")
	return nil
}

func seedAgama() error {
	agamaList := []string{"Islam", "Kristen", "Katolik", "Hindu", "Buddha", "Konghucu"}
	var agamas []models.Agama

	for _, nama := range agamaList {
		agamas = append(agamas, models.Agama{
			AgamaNama: nama,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})
	}

	if err := DB.Create(&agamas).Error; err != nil {
		return fmt.Errorf("failed to seed agama: %v", err)
	}
	log.Println("Seeded agama")
	return nil
}

func seedPekerjaan() error {
	pekerjaanList := []string{
		"PNS", "TNI", "Polri", "Karyawan Swasta", "Wiraswasta", "Petani", 
		"Nelayan", "Guru", "Dokter", "Perawat", "Pedagang", "Buruh", "Pensiunan",
	}
	var pekerjaans []models.Pekerjaan

	for _, nama := range pekerjaanList {
		pekerjaans = append(pekerjaans, models.Pekerjaan{
			PekerjaanNama: nama,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		})
	}

	if err := DB.Create(&pekerjaans).Error; err != nil {
		return fmt.Errorf("failed to seed pekerjaan: %v", err)
	}
	log.Println("Seeded pekerjaan")
	return nil
}

func seedKeluarga() error {
	var keluarga []models.Keluarga

	for i := 0; i < 20; i++ {
		status := "aktif"
		if i%10 == 0 {
			status = "nonaktif"
		}

		keluarga = append(keluarga, models.Keluarga{
			KeluargaNama:   faker.LastName() + " Family",
			KeluargaStatus: status,
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		})
	}

	if err := DB.Create(&keluarga).Error; err != nil {
		return fmt.Errorf("failed to seed keluarga: %v", err)
	}
	log.Println("Seeded keluarga")
	return nil
}

func seedWarga() error {
	var wargas []models.Warga
	var keluarga []models.Keluarga
	var agama []models.Agama
	var pekerjaan []models.Pekerjaan

	if err := DB.Find(&keluarga).Error; err != nil {
		return fmt.Errorf("failed to fetch keluarga: %v", err)
	}
	if err := DB.Find(&agama).Error; err != nil {
		return fmt.Errorf("failed to fetch agama: %v", err)
	}
	if err := DB.Find(&pekerjaan).Error; err != nil {
		return fmt.Errorf("failed to fetch pekerjaan: %v", err)
	}

	if len(keluarga) == 0 || len(agama) == 0 || len(pekerjaan) == 0 {
		return fmt.Errorf("required reference data not found")
	}

	kotaIndonesia := []string{
		"Jakarta", "Surabaya", "Bandung", "Medan", "Semarang", 
		"Makassar", "Palembang", "Denpasar", "Yogyakarta", "Malang",
	}

	for i := 0; i < 100; i++ {
		jenisKelamin := "L"
		if i%2 == 0 {
			jenisKelamin = "P"
		}

		statusAktif := "aktif"
		if i%20 == 0 {
			statusAktif = "nonaktif"
		}

		statusHidup := "hidup"
		if i%50 == 0 {
			statusHidup = "meninggal"
		}

		nik := "32"
		for j := 0; j < 14; j++ {
			nik += fmt.Sprintf("%d", rand.Intn(10))
		}

		warga := models.Warga{
			KeluargaID:        keluarga[rand.Intn(len(keluarga))].KeluargaID,
			WargaNama:         faker.FirstName() + " " + faker.LastName(),
			WargaNIK:          nik,
			WargaNoTlp:        "08" + fmt.Sprintf("%010d", rand.Intn(10000000000)),
			WargaTempatLahir:  kotaIndonesia[rand.Intn(len(kotaIndonesia))],
			WargaTanggalLahir: time.Now().AddDate(-rand.Intn(60)+18, -rand.Intn(12), -rand.Intn(30)),
			WargaJenisKelamin: jenisKelamin,
			WargaStatusAktif:  statusAktif,
			WargaStatusHidup:  statusHidup,
			AgamaID:           agama[rand.Intn(len(agama))].AgamaID,
			PekerjaanID:       pekerjaan[rand.Intn(len(pekerjaan))].PekerjaanID,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		}
		wargas = append(wargas, warga)
	}

	if err := DB.CreateInBatches(&wargas, 50).Error; err != nil {
		return fmt.Errorf("failed to seed warga: %v", err)
	}
	log.Println("Seeded warga")
	return nil
}

func seedRumah() error {
	var rumahs []models.Rumah
	var warga []models.Warga

	if err := DB.Find(&warga).Error; err != nil {
		return fmt.Errorf("failed to fetch warga: %v", err)
	}

	if len(warga) == 0 {
		return fmt.Errorf("no warga data found")
	}

	jalanList := []string{
		"Jl. Merdeka", "Jl. Sudirman", "Jl. Thamrin", "Jl. Gatot Subroto", "Jl. Asia Afrika",
	}
	kotaList := []string{"Jakarta", "Bandung", "Surabaya", "Medan", "Semarang"}

	for i := 0; i < 50; i++ {
		status := "tersedia"
		wargaID := uint(0)
		
		if i < 40 {
			status = "ditempati"
			wargaID = warga[rand.Intn(len(warga))].WargaID
		}

		alamat := fmt.Sprintf("%s No. %d, %s", 
			jalanList[rand.Intn(len(jalanList))], 
			rand.Intn(100)+1, 
			kotaList[rand.Intn(len(kotaList))])

		rumah := models.Rumah{
			RumahAlamat: alamat,
			RumahStatus: status,
			WargaID:     wargaID,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}
		rumahs = append(rumahs, rumah)
	}

	if err := DB.Create(&rumahs).Error; err != nil {
		return fmt.Errorf("failed to seed rumah: %v", err)
	}
	log.Println("Seeded rumah")
	return nil
}

func seedKategoriKegiatan() error {
	kategoriList := []string{"Gotong Royong", "Rapat RT", "Peringatan Hari Besar", "Olahraga", "Kesehatan"}
	var kategoris []models.KategoriKegiatan

	for _, nama := range kategoriList {
		kategoris = append(kategoris, models.KategoriKegiatan{
			KategoriKegiatanNama: nama,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		})
	}

	if err := DB.Create(&kategoris).Error; err != nil {
		return fmt.Errorf("failed to seed kategori kegiatan: %v", err)
	}
	log.Println("Seeded kategori kegiatan")
	return nil
}

func seedKegiatan() error {
	var kegiatans []models.Kegiatan
	var kategori []models.KategoriKegiatan

	if err := DB.Find(&kategori).Error; err != nil {
		return fmt.Errorf("failed to fetch kategori kegiatan: %v", err)
	}

	if len(kategori) == 0 {
		return fmt.Errorf("no kategori kegiatan data found")
	}

	lokasiList := []string{"Balai RW", "Lapangan", "Masjid", "Sekolah", "Puskesmas"}

	for i := 0; i < 30; i++ {
		kegiatan := models.Kegiatan{
			KegiatanNama:       faker.Sentence(),
			KategoriKegiatanID: kategori[rand.Intn(len(kategori))].KategoriKegiatanID,
			KegiatanTanggal:    time.Now().AddDate(0, 0, rand.Intn(60)-30),
			KegiatanLokasi:     lokasiList[rand.Intn(len(lokasiList))],
			KegiatanPJ:         faker.FirstName() + " " + faker.LastName(),
			KegiatanDeskripsi:  faker.Paragraph(),
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		}
		kegiatans = append(kegiatans, kegiatan)
	}

	if err := DB.Create(&kegiatans).Error; err != nil {
		return fmt.Errorf("failed to seed kegiatan: %v", err)
	}
	log.Println("Seeded kegiatan")
	return nil
}

func seedKategoriPengeluaran() error {
	kategoriList := []string{"Listrik", "Air", "Kebersihan", "Kegiatan RT", "Administrasi"}
	var kategoris []models.KategoriPengeluaran

	for _, nama := range kategoriList {
		kategoris = append(kategoris, models.KategoriPengeluaran{
			KategoriPengeluaranNama: nama,
			CreatedAt:               time.Now(),
			UpdatedAt:               time.Now(),
		})
	}

	if err := DB.Create(&kategoris).Error; err != nil {
		return fmt.Errorf("failed to seed kategori pengeluaran: %v", err)
	}
	log.Println("Seeded kategori pengeluaran")
	return nil
}

func seedPengeluaran() error {
	var pengeluarans []models.Pengeluaran
	var kategori []models.KategoriPengeluaran

	if err := DB.Find(&kategori).Error; err != nil {
		return fmt.Errorf("failed to fetch kategori pengeluaran: %v", err)
	}

	if len(kategori) == 0 {
		return fmt.Errorf("no kategori pengeluaran data found")
	}

	for i := 0; i < 50; i++ {
		pengeluaran := models.Pengeluaran{
			KategoriPengeluaranID: kategori[rand.Intn(len(kategori))].KategoriPengeluaranID,
			PengeluaranNama:       faker.Word(),
			PengeluaranTanggal:    time.Now().AddDate(0, 0, -rand.Intn(90)),
			PengeluaranNominal:    float64(rand.Intn(1000000) + 50000),
			PengeluaranBukti:      "bukti_" + faker.Word() + ".jpg",
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		}
		pengeluarans = append(pengeluarans, pengeluaran)
	}

	if err := DB.Create(&pengeluarans).Error; err != nil {
		return fmt.Errorf("failed to seed pengeluaran: %v", err)
	}
	log.Println("Seeded pengeluaran")
	return nil
}

func seedKategoriPemasukan() error {
	kategoriList := []string{"Iuran Warga", "Sumbangan", "Dana Desa", "Lain-lain"}
	var kategoris []models.KategoriPemasukan

	for _, nama := range kategoriList {
		kategoris = append(kategoris, models.KategoriPemasukan{
			KategoriPemasukanNama: nama,
			CreatedAt:             time.Now(),
			UpdatedAt:             time.Now(),
		})
	}

	if err := DB.Create(&kategoris).Error; err != nil {
		return fmt.Errorf("failed to seed kategori pemasukan: %v", err)
	}
	log.Println("Seeded kategori pemasukan")
	return nil
}

func seedPemasukan() error {
	var pemasukans []models.Pemasukan
	var kategori []models.KategoriPemasukan

	if err := DB.Find(&kategori).Error; err != nil {
		return fmt.Errorf("failed to fetch kategori pemasukan: %v", err)
	}

	if len(kategori) == 0 {
		return fmt.Errorf("no kategori pemasukan data found")
	}

	for i := 0; i < 40; i++ {
		pemasukan := models.Pemasukan{
			KategoriPemasukanID: kategori[rand.Intn(len(kategori))].KategoriPemasukanID,
			PemasukanNama:       faker.Word(),
			PemasukanTanggal:    time.Now().AddDate(0, 0, -rand.Intn(90)),
			PemasukanNominal:    float64(rand.Intn(2000000) + 100000),
			PemasukanBukti:      "bukti_" + faker.Word() + ".jpg",
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		pemasukans = append(pemasukans, pemasukan)
	}

	if err := DB.Create(&pemasukans).Error; err != nil {
		return fmt.Errorf("failed to seed pemasukan: %v", err)
	}
	log.Println("Seeded pemasukan")
	return nil
}

func seedTagihanIuran() error {
	tagihanList := []string{"Iuran Kebersihan", "Iuran Keamanan", "Iuran Kegiatan", "Iuran Sampah"}
	var tagihans []models.TagihanIuran

	for _, nama := range tagihanList {
		tagihans = append(tagihans, models.TagihanIuran{
			TagihanIuran: nama,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		})
	}

	if err := DB.Create(&tagihans).Error; err != nil {
		return fmt.Errorf("failed to seed tagihan iuran: %v", err)
	}
	log.Println("Seeded tagihan iuran")
	return nil
}

func seedKategoriProduk() error {
	kategoriList := []string{"Makanan", "Minuman", "Peralatan Rumah Tangga", "Elektronik"}
	var kategoris []models.KategoriProduk

	for _, nama := range kategoriList {
		kategoris = append(kategoris, models.KategoriProduk{
			KategoriProdukNama: nama,
			CreatedAt:          time.Now(),
			UpdatedAt:          time.Now(),
		})
	}

	if err := DB.Create(&kategoris).Error; err != nil {
		return fmt.Errorf("failed to seed kategori produk: %v", err)
	}
	log.Println("Seeded kategori produk")
	return nil
}

func seedProduk() error {
	var produks []models.Produk
	var kategori []models.KategoriProduk

	if err := DB.Find(&kategori).Error; err != nil {
		return fmt.Errorf("failed to fetch kategori produk: %v", err)
	}

	if len(kategori) == 0 {
		return fmt.Errorf("no kategori produk data found")
	}

	for i := 0; i < 20; i++ {
		produk := models.Produk{
			ProdukNama:       faker.Word(),
			ProdukDeskripsi:  faker.Sentence(),
			ProdukStok:       rand.Intn(100) + 1,
			ProdukHarga:      float64(rand.Intn(100000) + 5000),
			ProdukFoto:       "produk_" + faker.Word() + ".jpg",
			KategoriProdukID: kategori[rand.Intn(len(kategori))].KategoriProdukID,
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		}
		produks = append(produks, produk)
	}

	if err := DB.Create(&produks).Error; err != nil {
		return fmt.Errorf("failed to seed produk: %v", err)
	}
	log.Println("Seeded produk")
	return nil
}