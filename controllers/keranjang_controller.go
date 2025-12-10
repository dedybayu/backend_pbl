package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type KeranjangController struct {
	db *gorm.DB
}

func NewKeranjangController(db *gorm.DB) *KeranjangController {
	return &KeranjangController{db: db}
}

// Form data structs
type CreateKeranjangForm struct {
	ProdukID uint `form:"produk_id" binding:"required"`
	Jumlah   int  `form:"jumlah" binding:"required,min=1"`
}

type UpdateKeranjangForm struct {
	Jumlah int `form:"jumlah" binding:"required,min=1"`
}

// Response structs (tetap sama)
type KeranjangItemResponse struct {
	KeranjangID  uint          `json:"keranjang_id" example:"1"`
	UserID       uint          `json:"user_id" example:"1"`
	ProdukID     uint          `json:"produk_id" example:"1"`
	Jumlah       int           `json:"jumlah" example:"2"`
	HargaSatuan  float64       `json:"harga_satuan" example:"50000.00"`
	Subtotal     float64       `json:"subtotal" example:"100000.00"`
	Status       string        `json:"status" example:"active"`
	CreatedAt    time.Time     `json:"created_at"`
	UpdatedAt    time.Time     `json:"updated_at"`
	Produk       models.Produk `json:"produk"`
}

type CartSummaryResponse struct {
	Items       []KeranjangItemResponse `json:"items"`
	TotalHarga  float64                 `json:"total_harga" example:"150000.00"`
	TotalBarang int                     `json:"total_barang" example:"3"`
	TotalItem   int                     `json:"total_item" example:"2"`
}

// AddToCart godoc
// @Summary      Tambah produk ke keranjang (Form Data)
// @Description  Menambahkan produk ke keranjang belanja user menggunakan form data
// @Tags         keranjang
// @Accept       multipart/form-data
// @Produce      json
// @Param        produk_id formData int true "ID Produk" minimum(1)
// @Param        jumlah formData int true "Jumlah Produk" minimum(1)
// @Security     BearerAuth
// @Success      201  {object}  map[string]interface{}  "Produk berhasil ditambahkan ke keranjang"
// @Failure      400  {object}  map[string]interface{}  "Invalid form data"
// @Failure      404  {object}  map[string]interface{}  "Produk tidak ditemukan"
// @Failure      409  {object}  map[string]interface{}  "Stok produk tidak mencukupi"
// @Failure      500  {object}  map[string]interface{}  "Gagal menambahkan ke keranjang"
// @Router       /api/keranjang [post]
func (kc *KeranjangController) AddToCart(c *gin.Context) {
	var form CreateKeranjangForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Debug log
	fmt.Printf("Form data received - ProdukID: %d, Jumlah: %d\n", form.ProdukID, form.Jumlah)

	// Validasi jumlah
	if form.Jumlah <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Jumlah harus lebih dari 0",
		})
		return
	}

	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Cek apakah produk exists
	var produk models.Produk
	if err := kc.db.First(&produk, form.ProdukID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi produk",
			})
		}
		return
	}

	// Cek stok produk
	if produk.ProdukStok < form.Jumlah {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Stok produk tidak mencukupi",
			"details": gin.H{
				"stok_tersedia": produk.ProdukStok,
				"jumlah_diminta": form.Jumlah,
			},
		})
		return
	}

	// Cek apakah produk sudah ada di keranjang user dengan status active
	var existingKeranjang models.Keranjang
	err := kc.db.Where("user_id = ? AND produk_id = ? AND status = 'active'", userID, form.ProdukID).First(&existingKeranjang).Error
	
	if err == nil {
		// Jika sudah ada, update jumlah
		newJumlah := existingKeranjang.Jumlah + form.Jumlah
		
		// Validasi stok untuk jumlah baru
		if produk.ProdukStok < newJumlah {
			c.JSON(http.StatusConflict, gin.H{
				"error": "Stok produk tidak mencukupi untuk jumlah baru",
				"details": gin.H{
					"stok_tersedia": produk.ProdukStok,
					"jumlah_diminta": newJumlah,
					"jumlah_sekarang": existingKeranjang.Jumlah,
				},
			})
			return
		}

		// Update keranjang
		existingKeranjang.Jumlah = newJumlah
		existingKeranjang.UpdatedAt = time.Now()

		if err := kc.db.Save(&existingKeranjang).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal update jumlah produk di keranjang",
				"details": err.Error(),
			})
			return
		}

		// Reload dengan data produk
		kc.db.Preload("Produk").Preload("Produk.KategoriProduk").First(&existingKeranjang, existingKeranjang.KeranjangID)

		// Hitung harga satuan dan subtotal dari produk
		hargaSatuan := existingKeranjang.Produk.ProdukHarga
		subtotal := float64(existingKeranjang.Jumlah) * hargaSatuan

		c.JSON(http.StatusOK, gin.H{
			"message": "Jumlah produk di keranjang berhasil diupdate",
			"data": KeranjangItemResponse{
				KeranjangID:  existingKeranjang.KeranjangID,
				UserID:       existingKeranjang.UserID,
				ProdukID:     existingKeranjang.ProdukID,
				Jumlah:       existingKeranjang.Jumlah,
				HargaSatuan:  hargaSatuan,
				Subtotal:     subtotal,
				Status:       existingKeranjang.Status,
				CreatedAt:    existingKeranjang.CreatedAt,
				UpdatedAt:    existingKeranjang.UpdatedAt,
				Produk:       existingKeranjang.Produk,
			},
		})
		return
	}

	// Jika belum ada, buat keranjang baru
	keranjang := models.Keranjang{
		UserID:    userID.(uint),
		ProdukID:  form.ProdukID,
		Jumlah:    form.Jumlah,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := kc.db.Create(&keranjang).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menambahkan produk ke keranjang",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data produk
	if err := kc.db.Preload("Produk").Preload("Produk.KategoriProduk").First(&keranjang, keranjang.KeranjangID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data keranjang yang dibuat",
		})
		return
	}

	// Hitung harga satuan dan subtotal dari produk
	hargaSatuan := keranjang.Produk.ProdukHarga
	subtotal := float64(keranjang.Jumlah) * hargaSatuan

	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil ditambahkan ke keranjang",
		"data": KeranjangItemResponse{
			KeranjangID:  keranjang.KeranjangID,
			UserID:       keranjang.UserID,
			ProdukID:     keranjang.ProdukID,
			Jumlah:       keranjang.Jumlah,
			HargaSatuan:  hargaSatuan,
			Subtotal:     subtotal,
			Status:       keranjang.Status,
			CreatedAt:    keranjang.CreatedAt,
			UpdatedAt:    keranjang.UpdatedAt,
			Produk:       keranjang.Produk,
		},
	})
}

// GetCartItems (tetap sama, tidak perlu perubahan)
func (kc *KeranjangController) GetCartItems(c *gin.Context) {
	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var keranjang []models.Keranjang

	// Query hanya item dengan status active
	if err := kc.db.
		Preload("Produk").
		Preload("Produk.KategoriProduk").
		Preload("Produk.User").
		Where("user_id = ? AND status = 'active'", userID).
		Order("created_at DESC").
		Find(&keranjang).Error; err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data keranjang",
		})
		return
	}

	// Hitung total harga, total barang, dan mapping response
	var totalHarga float64
	var totalBarang int
	var items []KeranjangItemResponse
	
	for _, item := range keranjang {
		// Hitung harga satuan dan subtotal dari produk
		hargaSatuan := item.Produk.ProdukHarga
		subtotal := float64(item.Jumlah) * hargaSatuan
		
		totalHarga += subtotal
		totalBarang += item.Jumlah
		
		items = append(items, KeranjangItemResponse{
			KeranjangID:  item.KeranjangID,
			UserID:       item.UserID,
			ProdukID:     item.ProdukID,
			Jumlah:       item.Jumlah,
			HargaSatuan:  hargaSatuan,
			Subtotal:     subtotal,
			Status:       item.Status,
			CreatedAt:    item.CreatedAt,
			UpdatedAt:    item.UpdatedAt,
			Produk:       item.Produk,
		})
	}

	response := CartSummaryResponse{
		Items:       items,
		TotalHarga:  totalHarga,
		TotalBarang: totalBarang,
		TotalItem:   len(keranjang),
	}

	c.JSON(http.StatusOK, response)
}

// UpdateCartItem godoc
// @Summary      Update item keranjang (Form Data)
// @Description  Mengupdate jumlah item di keranjang menggunakan form data
// @Tags         keranjang
// @Accept       multipart/form-data
// @Produce      json
// @Param        id   path      int  true  "ID Keranjang"
// @Param        jumlah formData int true "Jumlah baru" minimum(1)
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}  "Item keranjang berhasil diupdate"
// @Failure      400  {object}  map[string]interface{}  "Invalid form data"
// @Failure      404  {object}  map[string]interface{}  "Item keranjang tidak ditemukan"
// @Failure      409  {object}  map[string]interface{}  "Stok produk tidak mencukupi"
// @Failure      500  {object}  map[string]interface{}  "Gagal mengupdate item keranjang"
// @Router       /api/keranjang/{id} [put]
func (kc *KeranjangController) UpdateCartItem(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	keranjangID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID keranjang tidak valid",
		})
		return
	}

	var form UpdateKeranjangForm
	if err := c.ShouldBind(&form); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Validasi jumlah
	if form.Jumlah <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Jumlah harus lebih dari 0",
		})
		return
	}

	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Cari item keranjang
	var keranjang models.Keranjang
	if err := kc.db.Preload("Produk").First(&keranjang, keranjangID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Item keranjang tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan item keranjang",
			})
		}
		return
	}

	// Cek apakah item milik user yang bersangkutan
	if keranjang.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Anda tidak memiliki akses ke item ini",
		})
		return
	}

	// Cek status
	if keranjang.Status != "active" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Item keranjang tidak aktif",
			"details": gin.H{
				"status": keranjang.Status,
			},
		})
		return
	}

	// Cek stok produk
	if keranjang.Produk.ProdukStok < form.Jumlah {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Stok produk tidak mencukupi",
			"details": gin.H{
				"stok_tersedia": keranjang.Produk.ProdukStok,
				"jumlah_diminta": form.Jumlah,
			},
		})
		return
	}

	// Update keranjang
	keranjang.Jumlah = form.Jumlah
	keranjang.UpdatedAt = time.Now()

	if err := kc.db.Save(&keranjang).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate item keranjang",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data produk
	kc.db.Preload("Produk").Preload("Produk.KategoriProduk").First(&keranjang, keranjangID)

	// Hitung harga satuan dan subtotal dari produk
	hargaSatuan := keranjang.Produk.ProdukHarga
	subtotal := float64(keranjang.Jumlah) * hargaSatuan

	c.JSON(http.StatusOK, gin.H{
		"message": "Item keranjang berhasil diupdate",
		"data": KeranjangItemResponse{
			KeranjangID:  keranjang.KeranjangID,
			UserID:       keranjang.UserID,
			ProdukID:     keranjang.ProdukID,
			Jumlah:       keranjang.Jumlah,
			HargaSatuan:  hargaSatuan,
			Subtotal:     subtotal,
			Status:       keranjang.Status,
			CreatedAt:    keranjang.CreatedAt,
			UpdatedAt:    keranjang.UpdatedAt,
			Produk:       keranjang.Produk,
		},
	})
}

// RemoveFromCart (tetap sama)
func (kc *KeranjangController) RemoveFromCart(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	keranjangID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID keranjang tidak valid",
		})
		return
	}

	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Cari item keranjang
	var keranjang models.Keranjang
	if err := kc.db.First(&keranjang, keranjangID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Item keranjang tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan item keranjang",
			})
		}
		return
	}

	// Cek apakah item milik user yang bersangkutan
	if keranjang.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "Anda tidak memiliki akses ke item ini",
		})
		return
	}

	// Soft delete dengan mengubah status
	keranjang.Status = "removed"
	keranjang.UpdatedAt = time.Now()

	if err := kc.db.Save(&keranjang).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus item dari keranjang",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Item berhasil dihapus dari keranjang",
	})
}

// ClearCart (tetap sama)
func (kc *KeranjangController) ClearCart(c *gin.Context) {
	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	// Update semua item dengan status active menjadi removed
	result := kc.db.Model(&models.Keranjang{}).
		Where("user_id = ? AND status = 'active'", userID).
		Updates(map[string]interface{}{
			"status":     "removed",
			"updated_at": time.Now(),
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengosongkan keranjang",
			"details": result.Error.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Keranjang berhasil dikosongkan",
		"deleted_count": result.RowsAffected,
	})
}

// GetCartCount (tetap sama)
func (kc *KeranjangController) GetCartCount(c *gin.Context) {
	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var totalBarang int64
	var totalItem int64

	// Hitung total barang
	if err := kc.db.Model(&models.Keranjang{}).
		Where("user_id = ? AND status = 'active'", userID).
		Select("COALESCE(SUM(jumlah), 0)").
		Scan(&totalBarang).Error; err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghitung total barang",
		})
		return
	}

	// Hitung total item (baris)
	if err := kc.db.Model(&models.Keranjang{}).
		Where("user_id = ? AND status = 'active'", userID).
		Count(&totalItem).Error; err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menghitung total item",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_barang": totalBarang,
		"total_item":   totalItem,
	})
}

// CheckoutPreview (tetap sama)
func (kc *KeranjangController) CheckoutPreview(c *gin.Context) {
	// Get user ID dari token
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	var keranjang []models.Keranjang

	// Query item dengan status active
	if err := kc.db.
		Preload("Produk").
		Preload("Produk.KategoriProduk").
		Preload("User").
		Where("user_id = ? AND status = 'active'", userID).
		Find(&keranjang).Error; err != nil {
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data keranjang untuk checkout",
		})
		return
	}

	if len(keranjang) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Keranjang kosong, tidak bisa checkout",
		})
		return
	}

	// Hitung ringkasan
	var totalHarga float64
	var totalBarang int
	var items []gin.H

	for _, item := range keranjang {
		// Hitung harga satuan dan subtotal dari produk
		hargaSatuan := item.Produk.ProdukHarga
		subtotal := float64(item.Jumlah) * hargaSatuan
		
		totalHarga += subtotal
		totalBarang += item.Jumlah
		
		items = append(items, gin.H{
			"keranjang_id":  item.KeranjangID,
			"produk_id":     item.ProdukID,
			"produk_nama":   item.Produk.ProdukNama,
			"produk_foto":   item.Produk.ProdukFoto,
			"jumlah":        item.Jumlah,
			"harga_satuan":  hargaSatuan,
			"subtotal":      subtotal,
			"stok_tersedia": item.Produk.ProdukStok,
		})
	}

	// Validasi stok untuk semua item
	var stokValid = true
	var stokErrors []gin.H

	for _, item := range keranjang {
		if item.Jumlah > item.Produk.ProdukStok {
			stokValid = false
			stokErrors = append(stokErrors, gin.H{
				"produk_id":     item.ProdukID,
				"produk_nama":   item.Produk.ProdukNama,
				"stok_tersedia": item.Produk.ProdukStok,
				"jumlah_diminta": item.Jumlah,
				"kekurangan":    item.Jumlah - item.Produk.ProdukStok,
			})
		}
	}

	if !stokValid {
		c.JSON(http.StatusConflict, gin.H{
			"error": "Stok beberapa produk tidak mencukupi",
			"stok_errors": stokErrors,
			"preview": gin.H{
				"total_harga":  totalHarga,
				"total_barang": totalBarang,
				"total_item":   len(keranjang),
				"items":        items,
			},
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"preview": gin.H{
			"total_harga":  totalHarga,
			"total_barang": totalBarang,
			"total_item":   len(keranjang),
			"items":        items,
		},
		"stok_valid": stokValid,
	})
}