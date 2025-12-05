package models

import (
	"time"
)

/* ============================
   LEVEL & USER
============================ */

type Level struct {
	LevelID   uint      `gorm:"primaryKey;autoIncrement" json:"level_id"`
	LevelKode string    `gorm:"unique;not null;size:10" json:"level_kode"`
	LevelNama string    `gorm:"not null;size:50" json:"level_nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Users []User `gorm:"foreignKey:LevelID"`
}

type User struct {
	UserID      uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username    string    `gorm:"unique;not null;size:50" json:"username"`
	Password    string    `gorm:"not null;size:255" json:"password"`
	LevelID     uint      `gorm:"not null" json:"level_id"`
	Level       Level     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"level"`
	FotoProfile string    `gorm:"size:255" json:"foto_profile"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

/* ============================
   KELUARGA
============================ */

type Keluarga struct {
	KeluargaID     uint      `gorm:"primaryKey;autoIncrement" json:"keluarga_id"`
	KeluargaNama   string    `gorm:"not null;size:100" json:"keluarga_nama"`
	KeluargaStatus string    `gorm:"type:enum('aktif','nonaktif');default:'aktif'" json:"keluarga_status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	Wargas          []Warga          `gorm:"foreignKey:KeluargaID"`
	MutasiKeluargas []MutasiKeluarga `gorm:"foreignKey:KeluargaID"`
}

/* ============================
   AGAMA & PEKERJAAN
============================ */

type Agama struct {
	AgamaID   uint      `gorm:"primaryKey;autoIncrement" json:"agama_id"`
	AgamaNama string    `gorm:"not null;size:50" json:"agama_nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Pekerjaan struct {
	PekerjaanID   uint      `gorm:"primaryKey;autoIncrement" json:"pekerjaan_id"`
	PekerjaanNama string    `gorm:"not null;size:100" json:"pekerjaan_nama"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

/* ============================
   WARGA
============================ */

type Warga struct {
	WargaID           uint      `gorm:"primaryKey;autoIncrement" json:"warga_id"`
	KeluargaID        uint      `gorm:"not null" json:"keluarga_id"`
	WargaNama         string    `gorm:"not null;size:100" json:"warga_nama"`
	WargaNIK          string    `gorm:"unique;not null;size:16" json:"warga_nik"`
	WargaNoTlp        string    `gorm:"size:15" json:"warga_no_tlp"`
	WargaTempatLahir  string    `gorm:"size:50" json:"warga_tempat_lahir"`
	WargaTanggalLahir time.Time `json:"warga_tanggal_lahir"`
	WargaJenisKelamin string    `gorm:"type:enum('L','P');default:'L'" json:"warga_jenis_kelamin"`
	WargaStatusAktif  string    `gorm:"type:enum('aktif','nonaktif');default:'aktif'" json:"warga_status_aktif"`
	WargaStatusHidup  string    `gorm:"type:enum('hidup','meninggal');default:'hidup'" json:"warga_status_hidup"`

	AgamaID     uint     `json:"agama_id"`
	PekerjaanID uint     `json:"pekerjaan_id"`

	Keluarga  Keluarga  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"keluarga"`
	Agama     *Agama     `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"agama"`
	Pekerjaan *Pekerjaan `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"pekerjaan"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Rumahs []Rumah `gorm:"foreignKey:WargaID"`
}

/* ============================
   RUMAH
============================ */

type Rumah struct {
	RumahID     uint      `gorm:"primaryKey;autoIncrement" json:"rumah_id"`
	RumahAlamat string    `gorm:"not null" json:"rumah_alamat"`
	RumahStatus string    `gorm:"type:enum('tersedia','ditempati');default:'tersedia'" json:"rumah_status"`

	WargaID uint  `json:"warga_id"`
	Warga   *Warga `gorm:"constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"warga"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

/* ============================
   KEGIATAN
============================ */

type KategoriKegiatan struct {
    KategoriKegiatanID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_kegiatan_id"`
    KategoriKegiatanNama string    `gorm:"not null;size:100" json:"kategori_kegiatan_nama"`
    CreatedAt            time.Time `json:"created_at"`
    UpdatedAt            time.Time `json:"updated_at"`

    // Relasi: 1 kategori → banyak kegiatan
    Kegiatans []Kegiatan `gorm:"foreignKey:KategoriKegiatanID"`
}


type Kegiatan struct {
    KegiatanID         uint      `gorm:"primaryKey;autoIncrement" json:"kegiatan_id"`
    KegiatanNama       string    `gorm:"not null;size:100" json:"kegiatan_nama"`

    // Foreign Key yang benar
    KategoriKegiatanID uint      `gorm:"not null" json:"kategori_kegiatan_id"`

    KegiatanTanggal    time.Time `json:"kegiatan_tanggal"`
    KegiatanLokasi     string    `json:"kegiatan_lokasi"`
    KegiatanPJ         string    `gorm:"size:100" json:"kegiatan_pj"`
    KegiatanDeskripsi  string    `gorm:"type:text" json:"kegiatan_deskripsi"`

    // Relasi ke parent kategori
    KategoriKegiatan   KategoriKegiatan `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_kegiatan"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}


/* ============================
   BROADCAST
============================ */

type Broadcast struct {
	BroadcastID        uint      `gorm:"primaryKey;autoIncrement" json:"broadcast_id"`
	BroadcastNama      string    `gorm:"not null;size:100" json:"broadcast_nama"`
	BroadcastDeskripsi string    `gorm:"type:text" json:"broadcast_deskripsi"`
	BroadcastFoto      string    `gorm:"size:255" json:"broadcast_foto"`
	BroadcastDokumen   string    `gorm:"size:255" json:"broadcast_dokumen"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

/* ============================
   MUTASI KELUARGA
============================ */


type MutasiKeluarga struct {
	MutasiKeluargaID      uint      `gorm:"primaryKey;autoIncrement" json:"mutasi_keluarga_id"`
	KeluargaID            uint      `gorm:"not null" json:"keluarga_id"`
	MutasiKeluargaJenis   string    `gorm:"not null;size:50" json:"mutasi_keluarga_jenis"`
	MutasiKeluargaAlasan  string    `gorm:"type:text" json:"mutasi_keluarga_alasan"`
	MutasiKeluargaTanggal time.Time `json:"mutasi_keluarga_tanggal"`

	Keluarga Keluarga `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"keluarga"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

/* ============================
   KEUANGAN (PENGELUARAN)
============================ */

type KategoriPengeluaran struct {
    KategoriPengeluaranID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_pengeluaran_id"`
    KategoriPengeluaranNama string    `gorm:"not null;size:100" json:"kategori_pengeluaran_nama"`
    CreatedAt               time.Time `json:"created_at"`
    UpdatedAt               time.Time `json:"updated_at"`

    // Relasi 1 kategori -> banyak pengeluaran
    Pengeluarans []Pengeluaran `gorm:"foreignKey:KategoriPengeluaranID"`
}


type Pengeluaran struct {
    PengeluaranID         uint      `gorm:"primaryKey;autoIncrement" json:"pengeluaran_id"`
    KategoriPengeluaranID uint      `gorm:"not null" json:"kategori_pengeluaran_id"`

    PengeluaranNama       string    `gorm:"not null;size:100" json:"pengeluaran_nama"`
    PengeluaranTanggal    time.Time `json:"pengeluaran_tanggal"`
    PengeluaranNominal    float64   `gorm:"not null;type:decimal(15,2)" json:"pengeluaran_nominal"`
    PengeluaranBukti      string    `gorm:"size:255" json:"pengeluaran_bukti"`

    KategoriPengeluaran   KategoriPengeluaran `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_pengeluaran"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}


/* ============================
   KEUANGAN (PEMASUKAN)
============================ */

type KategoriPemasukan struct {
    KategoriPemasukanID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_pemasukan_id"`
    KategoriPemasukanNama string    `gorm:"not null;size:100" json:"kategori_pemasukan_nama"`
    CreatedAt             time.Time `json:"created_at"`
    UpdatedAt             time.Time `json:"updated_at"`

    // 1 kategori punya banyak pemasukan
    Pemasukans []Pemasukan `gorm:"foreignKey:KategoriPemasukanID"`
}


type Pemasukan struct {
    PemasukanID         uint      `gorm:"primaryKey;autoIncrement" json:"pemasukan_id"`
    KategoriPemasukanID uint      `gorm:"not null" json:"kategori_pemasukan_id"`
    PemasukanNama       string    `gorm:"not null;size:100" json:"pemasukan_nama"`
    PemasukanTanggal    time.Time `json:"pemasukan_tanggal"`
    PemasukanNominal    float64   `gorm:"not null;type:decimal(15,2)" json:"pemasukan_nominal"`
    PemasukanBukti      string    `gorm:"size:255" json:"pemasukan_bukti"`

    // relasi many-to-one ke kategori
    KategoriPemasukan   KategoriPemasukan `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_pemasukan"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}


/* ============================
   TAGIHAN IURAN 
============================ */

//TODO konek sama warga

type TagihanIuran struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TagihanIuran string    `gorm:"not null;size:100" json:"tagihan_iuran"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

/* ============================
   PRODUK (ECOMMERCE)
============================ */

type KategoriProduk struct {
    KategoriProdukID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_produk_id"`
    KategoriProdukNama string    `gorm:"not null;size:100" json:"kategori_produk_nama"`
    CreatedAt          time.Time `json:"created_at"`
    UpdatedAt          time.Time `json:"updated_at"`

    // 1 kategori punya banyak produk
    Produks []Produk `gorm:"foreignKey:KategoriProdukID"`
}


type Produk struct {
    ProdukID         uint      `gorm:"primaryKey;autoIncrement" json:"produk_id"`
    ProdukNama       string    `gorm:"not null;size:100" json:"produk_nama"`
    ProdukDeskripsi  string    `gorm:"type:text" json:"produk_deskripsi"`
    ProdukStok       int       `gorm:"not null" json:"produk_stok"`
    ProdukHarga      float64   `gorm:"not null;type:decimal(15,2)" json:"produk_harga"`
    ProdukFoto       string    `gorm:"not null;size:255" json:"produk_foto"`
    KategoriProdukID uint      `gorm:"not null" json:"kategori_produk_id"`

    // relasi many-to-one (produk → kategori)
    KategoriProduk KategoriProduk `gorm:"constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_produk"`

    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}

