package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rt-management/helper"
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ProdukController struct {
	db *gorm.DB
}

func NewProdukController(db *gorm.DB) *ProdukController {
	return &ProdukController{db: db}
}

type CreateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama" binding:"required"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok" binding:"required"`
	ProdukHarga      float64 `form:"produk_harga" binding:"required"`
	ProdukFoto       string  `form:"-"`
	KategoriProdukID uint    `form:"kategori_produk_id" binding:"required"`
	UserID           uint    `form:"user_id" binding:"required"`
}

type UpdateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok"`
	ProdukHarga      float64 `form:"produk_harga"`
	ProdukFoto       string  `form:"-"`
	KategoriProdukID uint    `form:"kategori_produk_id"`
	UserID           uint    `form:"user_id"`
}



// CreateProduk godoc
// @Summary      Membuat produk baru
// @Description  Membuat produk baru dengan upload foto
// @Tags         Produk
// @Accept       multipart/form-data
// @Produce      json
// @Param        produk_nama       formData string true "Nama Produk"
// @Param        produk_deskripsi  formData string false "Deskripsi Produk"
// @Param        produk_stok       formData int true "Stok Produk"
// @Param        produk_harga      formData number true "Harga Produk"
// @Param        kategori_produk_id formData int true "ID Kategori Produk"
// @Param        user_id           formData int true "ID User"
// @Param        produk_foto       formData file true "Foto Produk"
// @Success      201  {object}  models.Produk
// @Failure      400  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router /api/produk [post]
func (pc *ProdukController) CreateProduk(c *gin.Context) {
	var req CreateProdukRequest

	// Bind form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.ProdukNama = strings.TrimSpace(req.ProdukNama)
	req.ProdukDeskripsi = strings.TrimSpace(req.ProdukDeskripsi)

	// Validasi UserID (user harus ada)
	var user models.User
	if err := pc.db.First(&user, req.UserID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "User tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi user",
			})
		}
		return
	}

	// Validasi required fields
	if req.ProdukNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama produk harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.ProdukNama) < 2 || len(req.ProdukNama) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama produk harus 2-200 karakter",
		})
		return
	}

	// Validasi stok
	if req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Validasi harga
	if req.ProdukHarga <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harga produk harus lebih dari 0",
		})
		return
	}

	// Check if kategori produk exists
	var kategori models.KategoriProduk
	if err := pc.db.First(&kategori, req.KategoriProdukID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Kategori produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal memvalidasi kategori produk",
			})
		}
		return
	}

	// Handle file upload
	fotoPath, uploadErr := helper.HandleFileImageUpload(c, "produk_foto", "")
	if uploadErr != nil {
		if uploadErr == http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Foto produk harus diupload",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto",
				"details": uploadErr.Error(),
			})
		}
		return
	}

	// Buat produk baru (masukkan UserID)
	produk := models.Produk{
		ProdukNama:       req.ProdukNama,
		ProdukDeskripsi:  req.ProdukDeskripsi,
		ProdukStok:       req.ProdukStok,
		ProdukHarga:      req.ProdukHarga,
		ProdukFoto:       fotoPath,
		KategoriProdukID: req.KategoriProdukID,
		UserID:           req.UserID, // penting
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := pc.db.Create(&produk).Error; err != nil {
		// Hapus file yang sudah diupload jika gagal menyimpan ke database
		helper.DeleteOldPhoto(fotoPath, "produk_foto")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori + user
	if err := pc.db.Preload("KategoriProduk").Preload("User").First(&produk, produk.ProdukID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang dibuat",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Produk berhasil dibuat",
		"data":    produk,
	})
}

// ✅ UPDATE - Mengupdate produk dengan file upload

// UpdateProduk godoc
// @Summary      Update produk
// @Description  Update produk termasuk upload foto baru
// @Tags         Produk
// @Accept       multipart/form-data
// @Produce      json
// @Param        id                 path int true "Produk ID"
// @Param        produk_nama       formData string false "Nama Produk"
// @Param        produk_deskripsi  formData string false "Deskripsi Produk"
// @Param        produk_stok       formData int false "Stok Produk"
// @Param        produk_harga      formData number false "Harga Produk"
// @Param        kategori_produk_id formData int false "ID Kategori Produk"
// @Param        user_id           formData int false "ID User"
// @Param        produk_foto       formData file false "Foto Produk"
// @Success      200  {object}  models.Produk
// @Failure      400  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /api/produk/{id} [put]
func (pc *ProdukController) UpdateProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	var req UpdateProdukRequest
	// Bind form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Validasi jika UserID diupdate
	if req.UserID != 0 {
		var user models.User
		if err := pc.db.First(&user, req.UserID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "User tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi user",
				})
			}
			return
		}
	}

	// Handle file upload (pass current photo path for potential deletion)
	fotoPath := ""
	hasNewPhoto := false
	if file, fileErr := c.FormFile("produk_foto"); fileErr == nil && file != nil {
		// ada file dalam request -> gunakan helper untuk menyimpan (dengan old filename sebagai backup)
		newPath, upErr := helper.HandleFileImageUpload(c, "produk_foto", produk.ProdukFoto)
		if upErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto",
				"details": upErr.Error(),
			})
			return
		}
		fotoPath = newPath
		hasNewPhoto = true
	} else {
		// tidak ada file upload -> tidak apa-apa (fileErr biasanya http.ErrMissingFile)
		hasNewPhoto = false
	}

	// Sanitize input
	if req.ProdukNama != "" {
		req.ProdukNama = strings.TrimSpace(req.ProdukNama)
	}
	if req.ProdukDeskripsi != "" {
		req.ProdukDeskripsi = strings.TrimSpace(req.ProdukDeskripsi)
	}

	// Validasi jika nama diupdate
	if req.ProdukNama != "" {
		if len(req.ProdukNama) < 2 || len(req.ProdukNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama produk harus 2-200 karakter",
			})
			return
		}

		// Check duplicate name (exclude current)
		// var existingProduk models.Produk
		// if err := pc.db.Where("produk_nama = ? AND produk_id != ?", req.ProdukNama, produkID).
		// 	First(&existingProduk).Error; err == nil {
		// 	c.JSON(http.StatusBadRequest, gin.H{
		// 		"error": "Produk dengan nama tersebut sudah ada",
		// 	})
		// 	return
		// }
	}

	// Validasi stok jika diupdate
	if req.ProdukStok != 0 && req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Validasi harga jika diupdate
	if req.ProdukHarga != 0 && req.ProdukHarga <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Harga produk harus lebih dari 0",
		})
		return
	}

	// Validasi kategori produk jika diupdate
	if req.KategoriProdukID != 0 {
		var kategori models.KategoriProduk
		if err := pc.db.First(&kategori, req.KategoriProdukID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{
					"error": "Kategori produk tidak ditemukan",
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{
					"error": "Gagal memvalidasi kategori produk",
				})
			}
			return
		}
	}

	// Update fields menggunakan map
	updates := make(map[string]interface{})

	if req.ProdukNama != "" {
		updates["produk_nama"] = req.ProdukNama
	}
	if req.ProdukDeskripsi != "" {
		updates["produk_deskripsi"] = req.ProdukDeskripsi
	}
	if req.ProdukStok != 0 {
		updates["produk_stok"] = req.ProdukStok
	}
	if req.ProdukHarga != 0 {
		updates["produk_harga"] = req.ProdukHarga
	}
	if hasNewPhoto && fotoPath != "" {
		updates["produk_foto"] = fotoPath
	}
	if req.KategoriProdukID != 0 {
		updates["kategori_produk_id"] = req.KategoriProdukID
	}
	if req.UserID != 0 {
		updates["user_id"] = req.UserID
	}

	updates["updated_at"] = time.Now()

	if err := pc.db.Model(&produk).Updates(updates).Error; err != nil {
		// Hapus file baru yang sudah diupload jika gagal update database
		if hasNewPhoto && fotoPath != "" {
			helper.DeleteOldPhoto(fotoPath, "produk_foto")
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate produk",
			"details": err.Error(),
		})
		return
	}

	// Jika update berhasil dan ada foto baru -> hapus foto lama dari storage (helper bisa melakukan ini, contoh:)
	if hasNewPhoto && produk.ProdukFoto != "" {
		// produk.ProdukFoto saat ini masih berisi old filename dari sebelum update,
		// tetapi karena kita menjalankan Updates di atas, produk struct belum otomatis berubah.
		// Ambil produk terbaru dulu lalu hapus file lama jika perlu.
		var updatedProduk models.Produk
		if err := pc.db.First(&updatedProduk, produkID).Error; err == nil {
			// jika foto lama berbeda dengan foto baru, hapus file lama
			if updatedProduk.ProdukFoto != "" && updatedProduk.ProdukFoto != fotoPath {
				helper.DeleteOldPhoto(produk.ProdukFoto, "produk_foto") // produk.ProdukFoto masih old
			}
		}
	}

	// Reload dengan data terbaru termasuk kategori dan user
	if err := pc.db.Preload("KategoriProduk").Preload("User").First(&produk, produkID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil diupdate",
		"data":    produk,
	})
}


// ✅ DELETE - Menghapus produk beserta file fotonya

// DeleteProduk godoc
// @Summary      Hapus produk
// @Description  Menghapus produk dan foto terkait
// @Tags         Produk
// @Produce      json
// @Param        id path int true "Produk ID"
// @Success      200 {object} map[string]string
// @Failure      400 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/produk/{id} [delete]
func (pc *ProdukController) DeleteProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	// Simpan path foto sebelum menghapus
	fotoPath := produk.ProdukFoto

	// Hapus dari database
	if err := pc.db.Delete(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus produk",
			"details": err.Error(),
		})
		return
	}

	// Hapus file foto
	if fotoPath != "" {
		helper.DeleteOldPhoto(fotoPath, "produk_foto")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil dihapus",
	})
}

// ✅ GET - Serve file foto produk

// GetProdukFoto godoc
// @Summary      Serve file foto produk
// @Description  Mengirim file foto produk
// @Tags         Produk
// @Produce      octet-stream
// @Param        filename path string true "Nama file foto"
// @Success      200 {file} file
// @Failure      400 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/produk/image/{filename} [get]
func (pc *ProdukController) GetProdukFoto(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	// Gunakan helper function GetFileByFileName
	file, err := helper.GetFileByFileName("produk_foto", filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File foto tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal membuka file",
				"details": err.Error(),
			})
		}
		return
	}
	defer file.Close()

	// Dapatkan file info untuk Content-Type
	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendapatkan info file",
		})
		return
	}

	// Set header yang sesuai
	ext := filepath.Ext(filename)
	contentType := helper.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	// Serve file
	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}

// ✅ READ - Mendapatkan semua produk (TETAP SAMA - GET request)

// GetAllProduk godoc
// @Summary      Mendapatkan semua produk
// @Description  Mendapatkan semua produk
// @Tags         Produk
// @Produce      json
// @Param        search query string false "Cari produk berdasarkan nama atau deskripsi"
// @Param        kategori_id query string false "Filter produk berdasarkan ID kategori"
// @Param        stok_min query string false "Filter produk berdasarkan stok minimum"
// @Param        stok_max query string false "Filter produk berdasarkan stok maksimum"
// @Param        harga_min query string false "Filter produk berdasarkan harga minimum"
// @Param        harga_max query string false "Filter produk berdasarkan harga maksimum"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/produk [get]
func (pc *ProdukController) GetAllProduk(c *gin.Context) {
	var produk []models.Produk

	// Parameter filter
	search := c.Query("search")
	kategoriID := c.Query("kategori_id")
	stokMin := c.Query("stok_min")
	stokMax := c.Query("stok_max")
	hargaMin := c.Query("harga_min")
	hargaMax := c.Query("harga_max")

	// Build base query
	query := pc.db.Model(&models.Produk{}).
		Preload("KategoriProduk").
		Preload("User")

	// === FILTERS ===

	// Search
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("produk_nama LIKE ? OR produk_deskripsi LIKE ?",
			"%"+searchSafe+"%", "%"+searchSafe+"%")
	}

	// Kategori
	if kategoriID != "" {
		if safeID, err := strconv.Atoi(kategoriID); err == nil {
			query = query.Where("kategori_produk_id = ?", safeID)
		}
	}

	// Stok minimum
	if stokMin != "" {
		if safe, err := strconv.Atoi(stokMin); err == nil {
			query = query.Where("produk_stok >= ?", safe)
		}
	}

	// Stok maksimum
	if stokMax != "" {
		if safe, err := strconv.Atoi(stokMax); err == nil {
			query = query.Where("produk_stok <= ?", safe)
		}
	}

	// Harga minimum
	if hargaMin != "" {
		if safe, err := strconv.ParseFloat(hargaMin, 64); err == nil {
			query = query.Where("produk_harga >= ?", safe)
		}
	}

	// Harga maksimum
	if hargaMax != "" {
		if safe, err := strconv.ParseFloat(hargaMax, 64); err == nil {
			query = query.Where("produk_harga <= ?", safe)
		}
	}

	// === EXECUTE ===
	if err := query.
		Order("created_at DESC").
		Find(&produk).Error; err != nil {

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk",
		})
		return
	}

	// === OUTPUT ===
	c.JSON(http.StatusOK, gin.H{
		"data": produk,
	})
}



// ✅ READ - Mendapatkan produk by ID (TETAP SAMA - GET request)

// GetProdukByID godoc
// @Summary      Mendapatkan produk berdasarkan ID
// @Description  Mendapatkan produk berdasarkan ID
// @Tags         Produk
// @Produce      json
// @Param        id path int true "Produk ID"
// @Success      200 {object} map[string]interface{}
// @Failure      400 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/produk/{id} [get]
func (pc *ProdukController) GetProdukByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data produk",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": produk,
	})
}
// ✅ GET - Produk terbaru (TETAP SAMA - GET request)

// GetProdukTerbaru godoc
// @Summary      Produk terbaru
// @Description  Mendapatkan daftar produk terbaru
// @Tags         Produk
// @Produce      json
// @Param        limit query int false "Limit produk"
// @Success      200  {array}  models.Produk
// @Failure      500  {object} map[string]interface{}
// @Router       /api/produk/terbaru [get]
func (pc *ProdukController) GetProdukTerbaru(c *gin.Context) {
	var produk []models.Produk

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "8"))

	// Query produk terbaru (AMAN)
	if err := pc.db.
		Preload("KategoriProduk").
		Order("created_at DESC").
		Limit(limit).
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk terbaru",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  produk,
		"total": len(produk),
	})
}

// ✅ GET - Produk dengan stok menipis (TETAP SAMA - GET request)

// GetProdukStokMenipis godoc
// @Summary      Produk stok menipis
// @Description  Mendapatkan produk dengan stok menipis (<10)
// @Tags         Produk
// @Produce      json
// @Success      200  {array}  models.Produk
// @Failure      500  {object} map[string]interface{}
// @Router       /api/produk/stok-menipis [get]
func (pc *ProdukController) GetProdukStokMenipis(c *gin.Context) {
	var produk []models.Produk

	// Query produk dengan stok menipis (AMAN)
	if err := pc.db.
		Preload("KategoriProduk").
		Where("produk_stok < 10").
		Order("produk_stok ASC").
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk stok menipis",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  produk,
		"total": len(produk),
	})
}

// ✅ GET - Statistik produk (TETAP SAMA - GET request)

// GetStatistikProduk godoc
// @Summary      Statistik produk
// @Description  Mendapatkan statistik produk (total, stok, nilai inventori, dll)
// @Tags         Produk
// @Produce      json
// @Success      200  {object} map[string]interface{}
// @Failure      500  {object} map[string]interface{}
// @Router       /api/produk/statistik [get]
func (pc *ProdukController) GetStatistikProduk(c *gin.Context) {
	type StatistikResult struct {
		TotalProduk      int64   `json:"total_produk"`
		TotalStok        int64   `json:"total_stok"`
		NilaiInventori   float64 `json:"nilai_inventori"`
		ProdukStokHabis  int64   `json:"produk_stok_habis"`
		ProdukStokMenipis int64  `json:"produk_stok_menipis"`
	}

	var statistik StatistikResult

	// Hitung total produk (AMAN)
	pc.db.Model(&models.Produk{}).Count(&statistik.TotalProduk)

	// Hitung total stok (AMAN)
	pc.db.Model(&models.Produk{}).
		Select("COALESCE(SUM(produk_stok), 0)").
		Row().
		Scan(&statistik.TotalStok)

	// Hitung nilai inventori (AMAN)
	pc.db.Model(&models.Produk{}).
		Select("COALESCE(SUM(produk_stok * produk_harga), 0)").
		Row().
		Scan(&statistik.NilaiInventori)

	// Hitung produk stok habis (AMAN)
	pc.db.Model(&models.Produk{}).
		Where("produk_stok = 0").
		Count(&statistik.ProdukStokHabis)

	// Hitung produk stok menipis (AMAN)
	pc.db.Model(&models.Produk{}).
		Where("produk_stok > 0 AND produk_stok < 10").
		Count(&statistik.ProdukStokMenipis)

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}

// ✅ PATCH - Update stok produk (FORM DATA)

// UpdateStokProduk godoc
// @Summary      Update stok produk
// @Description  Mengupdate stok produk (form data)
// @Tags         Produk
// @Accept       multipart/form-data
// @Produce      json
// @Param        id path int true "Produk ID"
// @Param        produk_stok formData int true "Stok baru"
// @Success      200 {object} models.Produk
// @Failure      400 {object} map[string]interface{}
// @Failure      404 {object} map[string]interface{}
// @Failure      500 {object} map[string]interface{}
// @Router       /api/produk/{id}/stok [patch]
func (pc *ProdukController) UpdateStokProduk(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	produkID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID produk tidak valid",
		})
		return
	}

	var produk models.Produk
	if err := pc.db.First(&produk, produkID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Produk tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan produk",
			})
		}
		return
	}

	type UpdateStokRequest struct {
		ProdukStok int `form:"produk_stok" binding:"required"`
	}

	var req UpdateStokRequest
	// Gunakan ShouldBind untuk form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Validasi stok
	if req.ProdukStok < 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Stok produk tidak boleh negatif",
		})
		return
	}

	// Update stok
	produk.ProdukStok = req.ProdukStok
	produk.UpdatedAt = time.Now()

	if err := pc.db.Save(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate stok produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data produk yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Stok produk berhasil diupdate",
		"data":    produk,
	})
}