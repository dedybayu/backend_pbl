package controllers

import (
	"net/http"
	"strconv"

	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DetailTransaksiController struct {
	db *gorm.DB
}

func NewDetailTransaksiController(db *gorm.DB) *DetailTransaksiController {
	return &DetailTransaksiController{db: db}
}

// GetDetailByTransaksiID godoc
// @Summary      Dapatkan detail transaksi
// @Description  Mendapatkan semua detail transaksi berdasarkan ID transaksi
// @Tags         detail-transaksi
// @Produce      json
// @Param        transaksi_id   path      int  true  "ID Transaksi"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar detail transaksi"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Transaksi tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data detail"
// @Router       /api/detail-transaksi/transaksi/{transaksi_id} [get]
func (dtc *DetailTransaksiController) GetDetailByTransaksiID(c *gin.Context) {
	transaksiIDStr := c.Param("transaksi_id")

	// Validasi ID
	transaksiID, err := strconv.ParseUint(transaksiIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID transaksi tidak valid",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	levelID, _ := c.Get("levelID")

	// Cek apakah transaksi exists dan user memiliki akses
	var transaksi models.Transaksi
	query := dtc.db.Model(&models.Transaksi{}) // PERBAIKAN: ganti tc.db menjadi dtc.db

	if levelID.(uint) != 1 { // Bukan admin
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&transaksi, transaksiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Transaksi tidak ditemukan atau Anda tidak memiliki akses",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi transaksi",
			})
		}
		return
	}

	var details []models.DetailTransaksi
	if err := dtc.db.
		Preload("Produk").
		Preload("Produk.KategoriProduk").
		Where("transaksi_id = ?", transaksiID).
		Order("created_at ASC").
		Find(&details).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data detail transaksi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":        details,
		"transaksi":   transaksi,
		"total_items": len(details),
	})
}

// GetDetailByID godoc
// @Summary      Dapatkan detail berdasarkan ID
// @Description  Mendapatkan detail transaksi berdasarkan ID detail
// @Tags         detail-transaksi
// @Produce      json
// @Param        id   path      int  true  "ID Detail Transaksi"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Detail transaksi"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Detail tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data detail"
// @Router       /api/detail-transaksi/{id} [get]
func (dtc *DetailTransaksiController) GetDetailByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	detailID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID detail transaksi tidak valid",
		})
		return
	}

	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	levelID, _ := c.Get("levelID")

	var detail models.DetailTransaksi
	query := dtc.db.
		Preload("Transaksi").
		Preload("Produk").
		Preload("Produk.KategoriProduk").
		First(&detail, detailID)

	if err := query.Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Detail transaksi tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data detail",
			})
		}
		return
	}

	// Cek akses: admin atau pemilik transaksi
	if levelID.(uint) != 1 && detail.Transaksi.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Anda tidak memiliki akses ke detail ini",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": detail,
	})
}