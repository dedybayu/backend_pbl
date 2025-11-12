package controllers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// Request structs - ubah binding untuk file upload
type CreateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama" binding:"required"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok" binding:"required"`
	ProdukHarga      float64 `form:"produk_harga" binding:"required"`
	ProdukFoto       string  `form:"-"` // Tidak binding dari form, akan dihandle secara manual
	KategoriProdukID uint    `form:"kategori_produk_id" binding:"required"`
}

type UpdateProdukRequest struct {
	ProdukNama       string  `form:"produk_nama"`
	ProdukDeskripsi  string  `form:"produk_deskripsi"`
	ProdukStok       int     `form:"produk_stok"`
	ProdukHarga      float64 `form:"produk_harga"`
	ProdukFoto       string  `form:"-"` // Tidak binding dari form, akan dihandle secara manual
	KategoriProdukID uint    `form:"kategori_produk_id"`
}

// Helper function untuk membuat direktori jika belum ada
func ensureDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// Helper function untuk menghapus file foto lama
func deleteOldPhoto(filePath string) error {
	if filePath != "" {
		fullPath := filepath.Join("storage", "images", "produk", filepath.Base(filePath))
		if _, err := os.Stat(fullPath); err == nil {
			return os.Remove(fullPath)
		}
	}
	return nil
}

// Helper function untuk handle file upload
func handleFileUpload(c *gin.Context, fieldName string, oldPhotoPath string) (string, error) {
	file, header, err := c.Request.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			// Jika tidak ada file baru diupload, return path foto lama
			return oldPhotoPath, nil
		}
		return "", err
	}
	defer file.Close()

	// Validasi tipe file
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %v", err)
	}

	contentType := http.DetectContentType(buffer)
	if !allowedTypes[contentType] {
		return "", fmt.Errorf("tipe file tidak diizinkan. Hanya JPEG, JPG, PNG, GIF, dan WebP yang diperbolehkan")
	}

	// Kembali ke awal file
	file.Seek(0, 0)

	// Buat nama file unik
	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().Format("20060102150405")
	randomStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	filename := fmt.Sprintf("produk_%s_%s%s", timestamp, randomStr[len(randomStr)-6:], ext)

	// Path penyimpanan
	storageDir := "storage/images/produk"
	if err := ensureDirectory(storageDir); err != nil {
		return "", fmt.Errorf("gagal membuat direktori: %v", err)
	}

	filePath := filepath.Join(storageDir, filename)

	// Buat file baru
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file: %v", err)
	}
	defer out.Close()

	// Salin konten file
	_, err = io.Copy(out, file)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file: %v", err)
	}

	// Hapus foto lama jika ada file baru diupload dan foto lama ada
	if oldPhotoPath != "" {
		deleteOldPhoto(oldPhotoPath)
	}

	return filename, nil
}

// ✅ CREATE - Membuat produk baru dengan file upload
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

	// Handle file upload
	fotoPath, err := handleFileUpload(c, "produk_foto", "")
	if err != nil {
		if err == http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Foto produk harus diupload",
			})
		} else {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto",
				"details": err.Error(),
			})
		}
		return
	}

	// Sanitize input
	req.ProdukNama = strings.TrimSpace(req.ProdukNama)
	req.ProdukDeskripsi = strings.TrimSpace(req.ProdukDeskripsi)

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

	// Check if produk dengan nama yang sama sudah ada
	var existingProduk models.Produk
	if err := pc.db.Where("produk_nama = ?", req.ProdukNama).First(&existingProduk).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Produk dengan nama tersebut sudah ada",
		})
		return
	}

	// Buat produk baru
	produk := models.Produk{
		ProdukNama:       req.ProdukNama,
		ProdukDeskripsi:  req.ProdukDeskripsi,
		ProdukStok:       req.ProdukStok,
		ProdukHarga:      req.ProdukHarga,
		ProdukFoto:       fotoPath, // Hanya menyimpan nama file
		KategoriProdukID: req.KategoriProdukID,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := pc.db.Create(&produk).Error; err != nil {
		// Hapus file yang sudah diupload jika gagal menyimpan ke database
		deleteOldPhoto(fotoPath)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data kategori
	if err := pc.db.Preload("KategoriProduk").First(&produk, produk.ProdukID).Error; err != nil {
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

	// Handle file upload (pass current photo path for potential deletion)
	fotoPath, err := handleFileUpload(c, "produk_foto", produk.ProdukFoto)
	if err != nil && err != http.ErrMissingFile {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Gagal mengupload foto",
			"details": err.Error(),
		})
		return
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
		var existingProduk models.Produk
		if err := pc.db.Where("produk_nama = ? AND produk_id != ?", req.ProdukNama, produkID).
			First(&existingProduk).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Produk dengan nama tersebut sudah ada",
			})
			return
		}
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
	if fotoPath != "" && err != http.ErrMissingFile {
		updates["produk_foto"] = fotoPath
	}
	if req.KategoriProdukID != 0 {
		updates["kategori_produk_id"] = req.KategoriProdukID
	}
	
	updates["updated_at"] = time.Now()

	if err := pc.db.Model(&produk).Updates(updates).Error; err != nil {
		// Hapus file baru yang sudah diupload jika gagal update database
		if fotoPath != "" && err != http.ErrMissingFile {
			deleteOldPhoto(fotoPath)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal mengupdate produk",
			"details": err.Error(),
		})
		return
	}

	// Reload dengan data terbaru termasuk kategori
	if err := pc.db.Preload("KategoriProduk").First(&produk, produkID).Error; err != nil {
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
		deleteOldPhoto(fotoPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Produk berhasil dihapus",
	})
}

// ✅ GET - Serve file foto produk
func (pc *ProdukController) GetProdukFoto(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	filePath := filepath.Join("storage", "images", "produk", filename)

	// Validasi path untuk mencegah directory traversal
	cleanPath := filepath.Clean(filePath)
	if !strings.HasPrefix(cleanPath, "storage/images/produk") {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path file tidak valid",
		})
		return
	}

	// Cek apakah file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File foto tidak ditemukan",
		})
		return
	}

	// Serve file
	c.File(filePath)
}

// ✅ READ - Mendapatkan semua produk (TETAP SAMA - GET request)
func (pc *ProdukController) GetAllProduk(c *gin.Context) {
	var produk []models.Produk

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Filter parameters
	kategoriID := c.Query("kategori_id")
	search := c.Query("search")
	stokMin := c.Query("stok_min")
	stokMax := c.Query("stok_max")
	hargaMin := c.Query("harga_min")
	hargaMax := c.Query("harga_max")

	// Build query dengan GORM (AMAN - parameterized queries)
	query := pc.db.Model(&models.Produk{}).Preload("KategoriProduk")

	// Apply filters
	if search != "" {
		searchSafe := strings.TrimSpace(search)
		query = query.Where("produk_nama LIKE ? OR produk_deskripsi LIKE ?", 
			"%"+searchSafe+"%", "%"+searchSafe+"%")
	}

	if kategoriID != "" {
		kategoriIDSafe, err := strconv.ParseUint(kategoriID, 10, 32)
		if err == nil {
			query = query.Where("kategori_produk_id = ?", kategoriIDSafe)
		}
	}

	if stokMin != "" {
		if stokMinSafe, err := strconv.Atoi(stokMin); err == nil {
			query = query.Where("produk_stok >= ?", stokMinSafe)
		}
	}

	if stokMax != "" {
		if stokMaxSafe, err := strconv.Atoi(stokMax); err == nil {
			query = query.Where("produk_stok <= ?", stokMaxSafe)
		}
	}

	if hargaMin != "" {
		if hargaMinSafe, err := strconv.ParseFloat(hargaMin, 64); err == nil {
			query = query.Where("produk_harga >= ?", hargaMinSafe)
		}
	}

	if hargaMax != "" {
		if hargaMaxSafe, err := strconv.ParseFloat(hargaMax, 64); err == nil {
			query = query.Where("produk_harga <= ?", hargaMaxSafe)
		}
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&produk).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data produk",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": produk,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan produk by ID (TETAP SAMA - GET request)
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