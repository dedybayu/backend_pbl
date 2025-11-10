package models

import (
	"time"
)

type Level struct {
	LevelID   uint      `gorm:"primaryKey;autoIncrement" json:"level_id"`
	LevelKode string    `gorm:"unique;not null" json:"level_kode"`
	LevelNama string    `gorm:"not null" json:"level_nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Users     []User    `gorm:"foreignKey:LevelID" json:"users"` // Relasi ke users
}

type User struct {
	UserID   uint      `gorm:"primaryKey;autoIncrement" json:"user_id"`
	Username string    `gorm:"unique;not null" json:"username"`
	Password string    `gorm:"not null" json:"password"`
	LevelID  uint      `gorm:"not null" json:"level_id"`
	Level    Level     `gorm:"foreignKey:LevelID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"level"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Keluarga struct {
	KeluargaID     uint      `gorm:"primaryKey;autoIncrement" json:"keluarga_id"`
	KeluargaNama   string    `gorm:"not null" json:"keluarga_nama"`
	KeluargaStatus string    `gorm:"type:enum('aktif','nonaktif');default:'aktif'" json:"keluarga_status"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Wargas         []Warga   `gorm:"foreignKey:KeluargaID" json:"wargas"`
}

type Agama struct {
	AgamaID   uint      `gorm:"primaryKey;autoIncrement" json:"agama_id"`
	AgamaNama string    `gorm:"not null" json:"agama_nama"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Wargas    []Warga   `gorm:"foreignKey:AgamaID" json:"wargas"`
}

type Pekerjaan struct {
	PekerjaanID   uint      `gorm:"primaryKey;autoIncrement" json:"pekerjaan_id"`
	PekerjaanNama string    `gorm:"not null" json:"pekerjaan_nama"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Wargas        []Warga   `gorm:"foreignKey:PekerjaanID" json:"wargas"`
}

type Warga struct {
	WargaID           uint      `gorm:"primaryKey;autoIncrement" json:"warga_id"`
	KeluargaID        uint      `gorm:"not null" json:"keluarga_id"`
	WargaNama         string    `gorm:"not null" json:"warga_nama"`
	WargaNIK          string    `gorm:"unique;not null" json:"warga_nik"`
	WargaNoTlp        string    `json:"warga_no_tlp"`
	WargaTempatLahir  string    `json:"warga_tempat_lahir"`
	WargaTanggalLahir time.Time `json:"warga_tanggal_lahir"`
	WargaJenisKelamin string    `gorm:"type:enum('L','P')" json:"warga_jenis_kelamin"`
	WargaStatusAktif  string    `gorm:"type:enum('aktif','nonaktif');default:'aktif'" json:"warga_status_aktif"`
	WargaStatusHidup  string    `gorm:"type:enum('hidup','meninggal');default:'hidup'" json:"warga_status_hidup"`
	AgamaID           uint      `json:"agama_id"`
	PekerjaanID       uint      `json:"pekerjaan_id"`
	Keluarga          Keluarga  `gorm:"foreignKey:KeluargaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"keluarga"`
	Agama             Agama     `gorm:"foreignKey:AgamaID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"agama"`
	Pekerjaan         Pekerjaan `gorm:"foreignKey:PekerjaanID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"pekerjaan"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Rumahs            []Rumah   `gorm:"foreignKey:WargaID" json:"rumahs"`
}

type Rumah struct {
	RumahID     uint      `gorm:"primaryKey;autoIncrement" json:"rumah_id"`
	RumahAlamat string    `gorm:"not null" json:"rumah_alamat"`
	RumahStatus string    `gorm:"type:enum('tersedia','ditempati');default:'tersedia'" json:"rumah_status"`
	WargaID     uint      `json:"warga_id"`
	Warga       Warga     `gorm:"foreignKey:WargaID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL;" json:"warga"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KategoriKegiatan struct {
	KategoriKegiatanID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_kegiatan_id"`
	KategoriKegiatanNama string    `gorm:"not null" json:"kategori_kegiatan_nama"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
	Kegiatans            []Kegiatan `gorm:"foreignKey:KategoriKegiatanID" json:"kegiatans"`
}

type Kegiatan struct {
	KegiatanID          uint              `gorm:"primaryKey;autoIncrement" json:"kegiatan_id"`
	KegiatanNama        string            `gorm:"not null" json:"kegiatan_nama"`
	KategoriKegiatanID  uint              `gorm:"not null" json:"kategori_kegiatan_id"`
	KegiatanTanggal     time.Time         `json:"kegiatan_tanggal"`
	KegiatanLokasi      string            `json:"kegiatan_lokasi"`
	KegiatanPJ          string            `json:"kegiatan_pj"`
	KegiatanDeskripsi   string            `json:"kegiatan_deskripsi"`
	KategoriKegiatan    KategoriKegiatan  `gorm:"foreignKey:KategoriKegiatanID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_kegiatan"`
	CreatedAt           time.Time         `json:"created_at"`
	UpdatedAt           time.Time         `json:"updated_at"`
}

type Broadcast struct {
	BroadcastID      uint      `gorm:"primaryKey;autoIncrement" json:"broadcast_id"`
	BroadcastNama    string    `gorm:"not null" json:"broadcast_nama"`
	BroadcastDeskripsi string  `json:"broadcast_deskripsi"`
	BroadcastFoto    string    `json:"broadcast_foto"`
	BroadcastDokumen string    `json:"broadcast_dokumen"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type MutasiKeluarga struct {
	MutasiKeluargaID    uint      `gorm:"primaryKey;autoIncrement" json:"mutasi_keluarga_id"`
	KeluargaID          uint      `gorm:"not null" json:"keluarga_id"`
	MutasiKeluargaJenis string    `gorm:"not null" json:"mutasi_keluarga_jenis"`
	MutasiKeluargaAlasan string   `json:"mutasi_keluarga_alasan"`
	MutasiKeluargaTanggal time.Time `json:"mutasi_keluarga_tanggal"`
	Keluarga            Keluarga  `gorm:"foreignKey:KeluargaID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"keluarga"`
	CreatedAt           time.Time `json:"created_at"`
	UpdatedAt           time.Time `json:"updated_at"`
}

type KategoriPengeluaran struct {
	KategoriPengeluaranID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_pengeluaran_id"`
	KategoriPengeluaranNama string    `gorm:"not null" json:"kategori_pengeluaran_nama"`
	CreatedAt               time.Time `json:"created_at"`
	UpdatedAt               time.Time `json:"updated_at"`
	Pengeluarans            []Pengeluaran `gorm:"foreignKey:KategoriPengeluaranID" json:"pengeluarans"`
}

type Pengeluaran struct {
	PengeluaranID          uint                 `gorm:"primaryKey;autoIncrement" json:"pengeluaran_id"`
	KategoriPengeluaranID  uint                 `gorm:"not null" json:"kategori_pengeluaran_id"`
	PengeluaranNama        string               `gorm:"not null" json:"pengeluaran_nama"`
	PengeluaranTanggal     time.Time            `json:"pengeluaran_tanggal"`
	PengeluaranNominal     float64              `gorm:"not null" json:"pengeluaran_nominal"`
	PengeluaranBukti       string               `json:"pengeluaran_bukti"`
	KategoriPengeluaran    KategoriPengeluaran  `gorm:"foreignKey:KategoriPengeluaranID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_pengeluaran"`
	CreatedAt              time.Time            `json:"created_at"`
	UpdatedAt              time.Time            `json:"updated_at"`
}

type KategoriPemasukan struct {
	KategoriPemasukanID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_pemasukan_id"`
	KategoriPemasukanNama string    `gorm:"not null" json:"kategori_pemasukan_nama"`
	CreatedAt             time.Time `json:"created_at"`
	UpdatedAt             time.Time `json:"updated_at"`
	Pemasukans            []Pemasukan `gorm:"foreignKey:KategoriPemasukanID" json:"pemasukans"`
}

type Pemasukan struct {
	PemasukanID          uint               `gorm:"primaryKey;autoIncrement" json:"pemasukan_id"`
	KategoriPemasukanID  uint               `gorm:"not null" json:"kategori_pemasukan_id"`
	PemasukanNama        string             `gorm:"not null" json:"pemasukan_nama"`
	PemasukanTanggal     time.Time          `json:"pemasukan_tanggal"`
	PemasukanNominal     float64            `gorm:"not null" json:"pemasukan_nominal"`
	PemasukanBukti       string             `json:"pemasukan_bukti"`
	KategoriPemasukan    KategoriPemasukan  `gorm:"foreignKey:KategoriPemasukanID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_pemasukan"`
	CreatedAt            time.Time          `json:"created_at"`
	UpdatedAt            time.Time          `json:"updated_at"`
}

type TagihanIuran struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	TagihanIuran string   `gorm:"not null" json:"tagihan_iuran"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type KategoriProduk struct {
	KategoriProdukID   uint      `gorm:"primaryKey;autoIncrement" json:"kategori_produk_id"`
	KategoriProdukNama string    `gorm:"not null" json:"kategori_produk_nama"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
	Produks            []Produk  `gorm:"foreignKey:KategoriProdukID" json:"produks"`
}

type Produk struct {
	ProdukID         uint           `gorm:"primaryKey;autoIncrement" json:"produk_id"`
	ProdukNama       string         `gorm:"not null" json:"produk_nama"`
	ProdukDeskripsi  string         `json:"produk_deskripsi"`
	ProdukStok       int            `gorm:"not null" json:"produk_stok"`
	ProdukHarga      float64        `gorm:"not null" json:"produk_harga"`
	ProdukFoto       string         `gorm:"not null" json:"produk_foto"`
	KategoriProdukID uint           `gorm:"not null" json:"kategori_produk_id"`
	KategoriProduk   KategoriProduk `gorm:"foreignKey:KategoriProdukID;constraint:OnUpdate:CASCADE,OnDelete:RESTRICT;" json:"kategori_produk"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}