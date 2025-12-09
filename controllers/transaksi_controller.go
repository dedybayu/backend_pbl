package controllers

import (
	// "fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"rt-management/models"
	"rt-management/helper"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TransaksiController struct {
	db *gorm.DB
}

func NewTransaksiController(db *gorm.DB) *TransaksiController {
	return &TransaksiController{db: db}
}

// Request structs untuk formdata
type CreateTransaksiRequest struct {
	MetodePembayaran string  `form:"metode_pembayaran" binding:"required" example:"tunai"`
	AlamatPengiriman string  `form:"alamat_pengiriman" binding:"required" example:"Jl. Contoh No. 123"`
	NamaPenerima     string  `form:"nama_penerima" binding:"required" example:"Budi Santoso"`
	NoTelpPenerima   string  `form:"no_telp_penerima" binding:"required" example:"081234567890"`
	Kurir            string  `form:"kurir" example:"JNE"`
	BiayaPengiriman  float64 `form:"biaya_pengiriman" example:"15000"`
	Diskon           float64 `form:"diskon" example:"0"`
	Catatan          string  `form:"catatan" example:"Tolong dikirim sebelum jam 5 sore"`
}

type UpdateTransaksiRequest struct {
	MetodePembayaran string     `form:"metode_pembayaran" example:"transfer_bank"`
	StatusPembayaran string     `form:"status_pembayaran" example:"dibayar"`
	StatusTransaksi  string     `form:"status_transaksi" example:"diproses"`
	AlamatPengiriman string     `form:"alamat_pengiriman" example:"Jl. Baru No. 456"`
	NamaPenerima     string     `form:"nama_penerima" example:"Siti Aminah"`
	NoTelpPenerima   string     `form:"no_telp_penerima" example:"081987654321"`
	Kurir            string     `form:"kurir" example:"TIKI"`
	NoResi           string     `form:"no_resi" example:"1234567890"`
	BiayaPengiriman  float64    `form:"biaya_pengiriman" example:"20000"`
	Diskon           float64    `form:"diskon" example:"5000"`
	Catatan          string     `form:"catatan" example:"Sudah dibayar via transfer"`
	TanggalBayar     *time.Time `form:"tanggal_bayar" example:"2024-01-15T10:30:00Z"`
	TanggalKirim     *time.Time `form:"tanggal_kirim" example:"2024-01-16T09:00:00Z"`
	TanggalSelesai   *time.Time `form:"tanggal_selesai" example:"2024-01-17T15:00:00Z"`
}

// CreateTransaksi godoc
// @Summary      Buat transaksi baru dari keranjang
// @Description  Membuat transaksi baru dengan mengambil semua item dari keranjang user
// @Tags         transaksi
// @Accept       multipart/form-data
// @Produce      json
// @Param        metode_pembayaran formData string true "Metode pembayaran (tunai, transfer_bank, e-wallet, qris)"
// @Param        alamat_pengiriman formData string true "Alamat pengiriman"
// @Param        nama_penerima formData string true "Nama penerima"
// @Param        no_telp_penerima formData string true "Nomor telepon penerima"
// @Param        kurir formData string false "Kurir pengiriman"
// @Param        biaya_pengiriman formData number false "Biaya pengiriman"
// @Param        diskon formData number false "Diskon"
// @Param        catatan formData string false "Catatan"
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "Transaksi berhasil dibuat"
// @Failure      400  {object}  map[string]interface{}  "Keranjang kosong atau data tidak valid"
// @Failure      409  {object}  map[string]interface{}  "Stok produk tidak mencukupi"
// @Failure      500  {object}  map[string]interface{}  "Gagal membuat transaksi"
// @Router       /api/transaksi [post]
func (tc *TransaksiController) CreateTransaksi(c *gin.Context) {
	var req CreateTransaksiRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Sanitize input
	req.MetodePembayaran = strings.TrimSpace(req.MetodePembayaran)
	req.AlamatPengiriman = strings.TrimSpace(req.AlamatPengiriman)
	req.NamaPenerima = strings.TrimSpace(req.NamaPenerima)
	req.NoTelpPenerima = strings.TrimSpace(req.NoTelpPenerima)
	req.Kurir = strings.TrimSpace(req.Kurir)
	req.Catatan = strings.TrimSpace(req.Catatan)

	// Validasi metode pembayaran
	validMetode := map[string]bool{
		"tunai":           true,
		"transfer_bank":   true,
		"e-wallet":        true,
		"qris":            true,
	}
	if !validMetode[req.MetodePembayaran] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Metode pembayaran tidak valid",
			"valid_options": []string{"tunai", "transfer_bank", "e-wallet", "qris"},
		})
		return
	}

	// Ambil semua item dari keranjang user
	var keranjang []models.Keranjang
	if err := tc.db.
		Preload("Produk").
		Where("user_id = ? AND status = 'active'", userID).
		Find(&keranjang).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data keranjang",
		})
		return
	}

	if len(keranjang) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Keranjang kosong, tidak bisa membuat transaksi",
		})
		return
	}

	// Validasi stok untuk semua item
	var totalHarga float64
	var stokErrors []gin.H

	for _, item := range keranjang {
		if item.Jumlah > item.Produk.ProdukStok {
			stokErrors = append(stokErrors, gin.H{
				"produk_id":     item.ProdukID,
				"produk_nama":   item.Produk.ProdukNama,
				"stok_tersedia": item.Produk.ProdukStok,
				"jumlah_diminta": item.Jumlah,
				"kekurangan":    item.Jumlah - item.Produk.ProdukStok,
			})
		}
		// Hitung total harga
		totalHarga += float64(item.Jumlah) * item.Produk.ProdukHarga
	}

	if len(stokErrors) > 0 {
		c.JSON(http.StatusConflict, gin.H{
			"error":        "Stok beberapa produk tidak mencukupi",
			"stok_errors":  stokErrors,
			"total_harga":  totalHarga,
		})
		return
	}

	// Hitung total bayar
	totalBayar := totalHarga - req.Diskon + req.BiayaPengiriman

	// Generate kode transaksi
	kodeTransaksi := helper.GenerateKodeTransaksi()

	// Mulai transaksi database
	tx := tc.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Buat transaksi
	transaksi := models.Transaksi{
		KodeTransaksi:    kodeTransaksi,
		UserID:           userID.(uint),
		TotalHarga:       totalHarga,
		Diskon:           req.Diskon,
		BiayaPengiriman:  req.BiayaPengiriman,
		TotalBayar:       totalBayar,
		MetodePembayaran: req.MetodePembayaran,
		StatusPembayaran: "menunggu",
		StatusTransaksi:  "menunggu",
		AlamatPengiriman: req.AlamatPengiriman,
		NamaPenerima:     req.NamaPenerima,
		NoTelpPenerima:   req.NoTelpPenerima,
		Kurir:            req.Kurir,
		Catatan:          req.Catatan,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := tx.Create(&transaksi).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat transaksi",
			"details": err.Error(),
		})
		return
	}

	// Buat detail transaksi dan update stok
	for _, item := range keranjang {
		hargaSatuan := item.Produk.ProdukHarga
		subtotal := float64(item.Jumlah) * hargaSatuan

		// Buat detail transaksi
		detail := models.DetailTransaksi{
			TransaksiID:  transaksi.TransaksiID,
			ProdukID:     item.ProdukID,
			Jumlah:       item.Jumlah,
			HargaSatuan:  hargaSatuan,
			Subtotal:     subtotal,
			CreatedAt:    time.Now(),
			UpdatedAt:    time.Now(),
		}

		if err := tx.Create(&detail).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal membuat detail transaksi",
				"details": err.Error(),
			})
			return
		}

		// Update stok produk
		newStok := item.Produk.ProdukStok - item.Jumlah
		if err := tx.Model(&models.Produk{}).
			Where("produk_id = ?", item.ProdukID).
			Update("produk_stok", newStok).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal update stok produk",
				"details": err.Error(),
			})
			return
		}
	}

	// Update status keranjang menjadi checked_out
	if err := tx.Model(&models.Keranjang{}).
		Where("user_id = ? AND status = 'active'", userID).
		Updates(map[string]interface{}{
			"status":     "checked_out",
			"updated_at": time.Now(),
		}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal update status keranjang",
			"details": err.Error(),
		})
		return
	}

	// Commit transaksi
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal commit transaksi",
			"details": err.Error(),
		})
		return
	}

	// Reload transaksi dengan detail
	tc.db.Preload("User").Preload("DetailTransaksi").Preload("DetailTransaksi.Produk").First(&transaksi, transaksi.TransaksiID)

	c.JSON(http.StatusCreated, gin.H{
		"message": "Transaksi berhasil dibuat",
		"data":    transaksi,
	})
}

// GetAllTransaksi godoc
// @Summary      Dapatkan semua transaksi
// @Description  Mendapatkan daftar semua transaksi dengan filter
// @Tags         transaksi
// @Produce      json
// @Param        page query int false "Halaman (default: 1)"
// @Param        limit query int false "Limit per halaman (default: 10)"
// @Param        status_pembayaran query string false "Filter status pembayaran"
// @Param        status_transaksi query string false "Filter status transaksi"
// @Param        start_date query string false "Filter tanggal mulai (YYYY-MM-DD)"
// @Param        end_date query string false "Filter tanggal sampai (YYYY-MM-DD)"
// @Param        search query string false "Pencarian kode transaksi atau nama penerima"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar transaksi"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data transaksi"
// @Router       /api/transaksi [get]
func (tc *TransaksiController) GetAllTransaksi(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Cek level user
	levelID, _ := c.Get("levelID")
	var transaksi []models.Transaksi

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	statusPembayaran := c.Query("status_pembayaran")
	statusTransaksi := c.Query("status_transaksi")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")
	search := c.Query("search")

	// Build query
	query := tc.db.Model(&models.Transaksi{}).
		Preload("User").
		Preload("DetailTransaksi").
		Preload("DetailTransaksi.Produk")

	// Admin bisa lihat semua, user biasa hanya lihat transaksi sendiri
	if levelID.(uint) != 1 { // Asumsi level 1 adalah admin
		query = query.Where("user_id = ?", userID)
	}

	// Apply filters
	if statusPembayaran != "" {
		query = query.Where("status_pembayaran = ?", statusPembayaran)
	}
	if statusTransaksi != "" {
		query = query.Where("status_transaksi = ?", statusTransaksi)
	}
	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}
	if search != "" {
		search = strings.TrimSpace(search)
		query = query.Where("kode_transaksi LIKE ? OR nama_penerima LIKE ?", 
			"%"+search+"%", "%"+search+"%")
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Execute query
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&transaksi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data transaksi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transaksi,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTransaksiByID godoc
// @Summary      Dapatkan transaksi berdasarkan ID
// @Description  Mendapatkan detail transaksi berdasarkan ID
// @Tags         transaksi
// @Produce      json
// @Param        id   path      int  true  "ID Transaksi"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Detail transaksi"
// @Failure      400  {object}  map[string]interface{}  "ID tidak valid"
// @Failure      404  {object}  map[string]interface{}  "Transaksi tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data transaksi"
// @Router       /api/transaksi/{id} [get]
func (tc *TransaksiController) GetTransaksiByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	transaksiID, err := strconv.ParseUint(id, 10, 32)
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

	var transaksi models.Transaksi
	query := tc.db.
		Preload("User").
		Preload("DetailTransaksi").
		Preload("DetailTransaksi.Produk").
		Preload("DetailTransaksi.Produk.KategoriProduk")

	// Admin bisa akses semua, user hanya transaksi sendiri
	if levelID.(uint) != 1 {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&transaksi, transaksiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Transaksi tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data transaksi",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transaksi,
	})
}

// UpdateTransaksi godoc
// @Summary      Update transaksi
// @Description  Mengupdate data transaksi (hanya admin)
// @Tags         transaksi
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "ID Transaksi"
// @Param        metode_pembayaran formData string false "Metode pembayaran"
// @Param        status_pembayaran formData string false "Status pembayaran"
// @Param        status_transaksi formData string false "Status transaksi"
// @Param        alamat_pengiriman formData string false "Alamat pengiriman"
// @Param        nama_penerima formData string false "Nama penerima"
// @Param        no_telp_penerima formData string false "No telepon penerima"
// @Param        kurir formData string false "Kurir"
// @Param        no_resi formData string false "No resi"
// @Param        biaya_pengiriman formData number false "Biaya pengiriman"
// @Param        diskon formData number false "Diskon"
// @Param        catatan formData string false "Catatan"
// @Param        tanggal_bayar formData string false "Tanggal bayar (RFC3339)"
// @Param        tanggal_kirim formData string false "Tanggal kirim (RFC3339)"
// @Param        tanggal_selesai formData string false "Tanggal selesai (RFC3339)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Transaksi berhasil diupdate"
// @Failure      400  {object}  map[string]interface{}  "Invalid request data"
// @Failure      403  {object}  map[string]interface{}  "Akses ditolak"
// @Failure      404  {object}  map[string]interface{}  "Transaksi tidak ditemukan"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengupdate transaksi"
// @Router       /api/transaksi/{id} [put]
func (tc *TransaksiController) UpdateTransaksi(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	transaksiID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID transaksi tidak valid",
		})
		return
	}

	// Hanya admin yang bisa update transaksi
	levelID, exists := c.Get("levelID")
	if !exists || levelID.(uint) != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Hanya admin yang bisa mengupdate transaksi",
		})
		return
	}

	var req UpdateTransaksiRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request data",
			"details": err.Error(),
		})
		return
	}

	var transaksi models.Transaksi
	if err := tc.db.First(&transaksi, transaksiID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Transaksi tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan transaksi",
			})
		}
		return
	}

	// Sanitize input
	if req.MetodePembayaran != "" {
		req.MetodePembayaran = strings.TrimSpace(req.MetodePembayaran)
		// Validasi metode pembayaran
		validMetode := map[string]bool{
			"tunai":         true,
			"transfer_bank": true,
			"e-wallet":      true,
			"qris":          true,
		}
		if !validMetode[req.MetodePembayaran] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Metode pembayaran tidak valid",
			})
			return
		}
	}

	if req.StatusPembayaran != "" {
		req.StatusPembayaran = strings.TrimSpace(req.StatusPembayaran)
		validStatus := map[string]bool{
			"menunggu":  true,
			"dibayar":   true,
			"ditolak":   true,
			"kadaluarsa": true,
		}
		if !validStatus[req.StatusPembayaran] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status pembayaran tidak valid",
			})
			return
		}
	}

	if req.StatusTransaksi != "" {
		req.StatusTransaksi = strings.TrimSpace(req.StatusTransaksi)
		validStatus := map[string]bool{
			"menunggu":    true,
			"diproses":    true,
			"dikirim":     true,
			"selesai":     true,
			"dibatalkan":  true,
		}
		if !validStatus[req.StatusTransaksi] {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Status transaksi tidak valid",
			})
			return
		}
	}

	// Update fields
	updates := make(map[string]interface{})
	
	if req.MetodePembayaran != "" {
		updates["metode_pembayaran"] = req.MetodePembayaran
	}
	if req.StatusPembayaran != "" {
		updates["status_pembayaran"] = req.StatusPembayaran
	}
	if req.StatusTransaksi != "" {
		updates["status_transaksi"] = req.StatusTransaksi
	}
	if req.AlamatPengiriman != "" {
		updates["alamat_pengiriman"] = strings.TrimSpace(req.AlamatPengiriman)
	}
	if req.NamaPenerima != "" {
		updates["nama_penerima"] = strings.TrimSpace(req.NamaPenerima)
	}
	if req.NoTelpPenerima != "" {
		updates["no_telp_penerima"] = strings.TrimSpace(req.NoTelpPenerima)
	}
	if req.Kurir != "" {
		updates["kurir"] = strings.TrimSpace(req.Kurir)
	}
	if req.NoResi != "" {
		updates["no_resi"] = strings.TrimSpace(req.NoResi)
	}
	if req.BiayaPengiriman != 0 {
		updates["biaya_pengiriman"] = req.BiayaPengiriman
		// Recalculate total bayar
		updates["total_bayar"] = transaksi.TotalHarga - transaksi.Diskon + req.BiayaPengiriman
	}
	if req.Diskon != 0 {
		updates["diskon"] = req.Diskon
		// Recalculate total bayar
		updates["total_bayar"] = transaksi.TotalHarga - req.Diskon + transaksi.BiayaPengiriman
	}
	if req.Catatan != "" {
		updates["catatan"] = strings.TrimSpace(req.Catatan)
	}
	if req.TanggalBayar != nil {
		updates["tanggal_bayar"] = req.TanggalBayar
	}
	if req.TanggalKirim != nil {
		updates["tanggal_kirim"] = req.TanggalKirim
	}
	if req.TanggalSelesai != nil {
		updates["tanggal_selesai"] = req.TanggalSelesai
	}
	
	updates["updated_at"] = time.Now()

	if err := tc.db.Model(&transaksi).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate transaksi",
			"details": err.Error(),
		})
		return
	}

	// Reload data
	tc.db.Preload("User").Preload("DetailTransaksi").Preload("DetailTransaksi.Produk").First(&transaksi, transaksiID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Transaksi berhasil diupdate",
		"data":    transaksi,
	})
}

// GetTransaksiByUser godoc
// @Summary      Dapatkan transaksi user
// @Description  Mendapatkan daftar transaksi milik user yang sedang login
// @Tags         transaksi
// @Produce      json
// @Param        page query int false "Halaman (default: 1)"
// @Param        limit query int false "Limit per halaman (default: 10)"
// @Param        status query string false "Filter status transaksi"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Daftar transaksi user"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil data transaksi"
// @Router       /api/transaksi/my [get]
func (tc *TransaksiController) GetTransaksiByUser(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var transaksi []models.Transaksi

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	status := c.Query("status")

	query := tc.db.Model(&models.Transaksi{}).
		Preload("User").
		Preload("DetailTransaksi").
		Preload("DetailTransaksi.Produk").
		Where("user_id = ?", userID)

	if status != "" {
		query = query.Where("status_transaksi = ?", status)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Execute query
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&transaksi).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data transaksi",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": transaksi,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// GetTransaksiStats godoc
// @Summary      Dapatkan statistik transaksi
// @Description  Mendapatkan statistik transaksi (hanya admin)
// @Tags         transaksi
// @Produce      json
// @Param        start_date query string false "Tanggal mulai (YYYY-MM-DD)"
// @Param        end_date query string false "Tanggal sampai (YYYY-MM-DD)"
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Statistik transaksi"
// @Failure      403  {object}  map[string]interface{}  "Akses ditolak"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengambil statistik"
// @Router       /api/transaksi/stats [get]
func (tc *TransaksiController) GetTransaksiStats(c *gin.Context) {
	// Hanya admin yang bisa akses statistik
	levelID, exists := c.Get("levelID")
	if !exists || levelID.(uint) != 1 {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Hanya admin yang bisa mengakses statistik",
		})
		return
	}

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// Query builder
	query := tc.db.Model(&models.Transaksi{})

	if startDate != "" {
		query = query.Where("DATE(created_at) >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("DATE(created_at) <= ?", endDate)
	}

	var stats struct {
		TotalTransaksi  int64   `json:"total_transaksi"`
		TotalPendapatan float64 `json:"total_pendapatan"`
		TransaksiBaru   int64   `json:"transaksi_baru"`
		TransaksiSelesai int64  `json:"transaksi_selesai"`
	}

	// Total transaksi
	query.Count(&stats.TotalTransaksi)

	// Total pendapatan (hanya transaksi dengan status pembayaran = dibayar)
	tc.db.Model(&models.Transaksi{}).
		Where("status_pembayaran = 'dibayar'").
		Select("COALESCE(SUM(total_bayar), 0)").
		Scan(&stats.TotalPendapatan)

	// Transaksi baru hari ini
	tc.db.Model(&models.Transaksi{}).
		Where("DATE(created_at) = ?", time.Now().Format("2006-01-02")).
		Count(&stats.TransaksiBaru)

	// Transaksi selesai hari ini
	tc.db.Model(&models.Transaksi{}).
		Where("status_transaksi = 'selesai' AND DATE(updated_at) = ?", time.Now().Format("2006-01-02")).
		Count(&stats.TransaksiSelesai)

	c.JSON(http.StatusOK, gin.H{
		"data": stats,
		"period": gin.H{
			"start_date": startDate,
			"end_date":   endDate,
		},
	})
}