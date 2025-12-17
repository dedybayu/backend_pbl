package controllers

import (
	"net/http"

	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardController struct {
	db *gorm.DB
}

func NewDashboardController(db *gorm.DB) *DashboardController {
	return &DashboardController{db: db}
}

// DashboardStatistik godoc
// @Summary      Statistik Dashboard
// @Description  Mengambil total warga, keuangan, total produk, total kegiatan
// @Tags         Dashboard
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/dashboard/statistik [get]
// @Security     ApiKeyAuth
func (dc *DashboardController) DashboardStatistik(c *gin.Context) {
	var (
		totalWarga      int64
		totalProduk     int64
		totalKegiatan   int64
		totalPemasukan  float64
		totalPengeluaran float64
	)

	// Total warga (hanya yang aktif & hidup — opsional, bisa dihapus kalau mau semua)
	dc.db.Model(&models.Warga{}).
		Where("warga_status_hidup = ?", "hidup").
		Count(&totalWarga)

	// Total produk
	dc.db.Model(&models.Produk{}).
		Count(&totalProduk)

	// Total kegiatan
	dc.db.Model(&models.Kegiatan{}).
		Count(&totalKegiatan)

	// Total pemasukan
	dc.db.Model(&models.Pemasukan{}).
		Select("COALESCE(SUM(pemasukan_nominal), 0)").
		Scan(&totalPemasukan)

	// Total pengeluaran
	dc.db.Model(&models.Pengeluaran{}).
		Select("COALESCE(SUM(pengeluaran_nominal), 0)").
		Scan(&totalPengeluaran)

	// Hitung keuangan
	totalKeuangan := totalPemasukan - totalPengeluaran

	c.JSON(http.StatusOK, gin.H{
		"total_warga":    totalWarga,
		"keuangan":       totalKeuangan,
		"total_produk":   totalProduk,
		"total_kegiatan": totalKegiatan,
	})
}

// GetStatistikKeuangan godoc
// @Summary      Statistik Keuangan
// @Description  Mengambil detail statistik keuangan (pemasukan, pengeluaran, total dana)
// @Tags         Dashboard
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]string
// @Failure      403  {object}  map[string]string
// @Router       /api/dashboard/keuangan [get]
// @Security     ApiKeyAuth
func (dc *DashboardController) GetStatistikKeuangan(c *gin.Context) {
	var (
		totalPemasukan  float64
		totalPengeluaran float64
		jumlahPemasukan int64
		jumlahPengeluaran int64
	)

	// Total pemasukan
	dc.db.Model(&models.Pemasukan{}).
		Select("COALESCE(SUM(pemasukan_nominal), 0)").
		Scan(&totalPemasukan)

	// Total pengeluaran
	dc.db.Model(&models.Pengeluaran{}).
		Select("COALESCE(SUM(pengeluaran_nominal), 0)").
		Scan(&totalPengeluaran)

	// Jumlah Pemasukan (kali)
	dc.db.Model(&models.Pemasukan{}).
		Count(&jumlahPemasukan)

	// Jumlah Pengeluaran (kali)
	dc.db.Model(&models.Pengeluaran{}).
		Count(&jumlahPengeluaran)

	// Hitung total dana
	totalDana := totalPemasukan - totalPengeluaran

	c.JSON(http.StatusOK, gin.H{
		"total_pemasukan":   totalPemasukan,
		"total_pengeluaran": totalPengeluaran,
		"total_dana":        totalDana,
		"jumlah_pemasukan":  jumlahPemasukan,
		"jumlah_pengeluaran": jumlahPengeluaran,
	})
}
