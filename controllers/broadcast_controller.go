package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"rt-management/helper" // Import package helper
	"rt-management/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type BroadcastController struct {
	db *gorm.DB
}

func NewBroadcastController(db *gorm.DB) *BroadcastController {
	return &BroadcastController{db: db}
}

// Request structs untuk form-data
type CreateBroadcastRequest struct {
	BroadcastNama      string `form:"broadcast_nama" binding:"required"`
	BroadcastDeskripsi string `form:"broadcast_deskripsi"`
}

type UpdateBroadcastRequest struct {
	BroadcastNama      string `form:"broadcast_nama"`
	BroadcastDeskripsi string `form:"broadcast_deskripsi"`
}

// ✅ CREATE - Membuat broadcast baru dengan form-data
func (bc *BroadcastController) CreateBroadcast(c *gin.Context) {
	var req CreateBroadcastRequest
	
	// Bind form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	req.BroadcastNama = strings.TrimSpace(req.BroadcastNama)
	req.BroadcastDeskripsi = strings.TrimSpace(req.BroadcastDeskripsi)

	// Validasi required fields
	if req.BroadcastNama == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama broadcast harus diisi",
		})
		return
	}

	// Validasi panjang nama
	if len(req.BroadcastNama) < 2 || len(req.BroadcastNama) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama broadcast harus 2-200 karakter",
		})
		return
	}

	// Check if broadcast dengan nama yang sama sudah ada
	var existingBroadcast models.Broadcast
	if err := bc.db.Where("broadcast_nama = ?", req.BroadcastNama).First(&existingBroadcast).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Broadcast dengan nama tersebut sudah ada",
		})
		return
	}

	// Handle file upload untuk foto
	broadcastFoto, err := helper.HandleFileImageUpload(c, "broadcast_foto", "")
	if err != nil {
		if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto broadcast",
				"details": err.Error(),
			})
			return
		}
		// Jika tidak ada file foto yang diupload, broadcastFoto akan kosong
	}

	// Handle file upload untuk dokumen
	broadcastDokumen, err := helper.HandleFileDokumenUpload(c, "broadcast_dokumen", "")
	if err != nil {
		if err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload dokumen broadcast",
				"details": err.Error(),
			})
			return
		}
		// Jika tidak ada file dokumen yang diupload, broadcastDokumen akan kosong
	}

	// Buat broadcast baru
	broadcast := models.Broadcast{
		BroadcastNama:      req.BroadcastNama,
		BroadcastDeskripsi: req.BroadcastDeskripsi,
		BroadcastFoto:      broadcastFoto,
		BroadcastDokumen:   broadcastDokumen,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}

	if err := bc.db.Create(&broadcast).Error; err != nil {
		// Rollback file upload jika gagal menyimpan ke database
		if broadcastFoto != "" {
			helper.DeleteOldPhoto(broadcastFoto, "broadcast_foto")
		}
		if broadcastDokumen != "" {
			helper.DeleteOldDocument(broadcastDokumen, "broadcast_dokumen")
		}
		
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal membuat broadcast",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Broadcast berhasil dibuat",
		"data":    broadcast,
	})
}

// ✅ UPDATE - Mengupdate broadcast dengan form-data
func (bc *BroadcastController) UpdateBroadcast(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan broadcast",
			})
		}
		return
	}

	var req UpdateBroadcastRequest
	
	// Bind form data
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid form data",
			"details": err.Error(),
		})
		return
	}

	// Sanitize input
	if req.BroadcastNama != "" {
		req.BroadcastNama = strings.TrimSpace(req.BroadcastNama)
	}
	if req.BroadcastDeskripsi != "" {
		req.BroadcastDeskripsi = strings.TrimSpace(req.BroadcastDeskripsi)
	}

	// Validasi jika nama diupdate
	if req.BroadcastNama != "" {
		if len(req.BroadcastNama) < 2 || len(req.BroadcastNama) > 200 {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Nama broadcast harus 2-200 karakter",
			})
			return
		}

		// Check duplicate name (exclude current)
		var existingBroadcast models.Broadcast
		if err := bc.db.Where("broadcast_nama = ? AND broadcast_id != ?", req.BroadcastNama, broadcastID).
			First(&existingBroadcast).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": "Broadcast dengan nama tersebut sudah ada",
			})
			return
		}
	}

	// Handle file upload untuk foto (jika ada file baru)
	var newFoto string
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File["broadcast_foto"] != nil {
		newFoto, err = helper.HandleFileImageUpload(c, "broadcast_foto", broadcast.BroadcastFoto)
		if err != nil && err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload foto broadcast",
				"details": err.Error(),
			})
			return
		}
	}

	// Handle file upload untuk dokumen (jika ada file baru)
	var newDokumen string
	if c.Request.MultipartForm != nil && c.Request.MultipartForm.File["broadcast_dokumen"] != nil {
		newDokumen, err = helper.HandleFileDokumenUpload(c, "broadcast_dokumen", broadcast.BroadcastDokumen)
		if err != nil && err != http.ErrMissingFile {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Gagal mengupload dokumen broadcast",
				"details": err.Error(),
			})
			return
		}
	}

	// Update fields menggunakan map
	updates := make(map[string]interface{})
	
	if req.BroadcastNama != "" {
		updates["broadcast_nama"] = req.BroadcastNama
	}
	if req.BroadcastDeskripsi != "" {
		updates["broadcast_deskripsi"] = req.BroadcastDeskripsi
	}
	if newFoto != "" {
		updates["broadcast_foto"] = newFoto
	}
	if newDokumen != "" {
		updates["broadcast_dokumen"] = newDokumen
	}
	
	updates["updated_at"] = time.Now()

	if len(updates) > 0 {
		if err := bc.db.Model(&broadcast).Updates(updates).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Gagal mengupdate broadcast",
				"details": err.Error(),
			})
			return
		}
	}

	// Reload dengan data terbaru
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal memuat data broadcast yang diupdate",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast berhasil diupdate",
		"data":    broadcast,
	})
}

// ✅ DELETE - Menghapus broadcast
func (bc *BroadcastController) DeleteBroadcast(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menemukan broadcast",
			})
		}
		return
	}

	// Simpan nama file untuk dihapus nanti
	fotoToDelete := broadcast.BroadcastFoto
	dokumenToDelete := broadcast.BroadcastDokumen

	// Delete dari database
	if err := bc.db.Delete(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Gagal menghapus broadcast",
			"details": err.Error(),
		})
		return
	}

	// Hapus file dari storage setelah berhasil delete dari database
	if fotoToDelete != "" {
		helper.DeleteOldPhoto(fotoToDelete, "broadcast_foto")
	}
	if dokumenToDelete != "" {
		helper.DeleteOldDocument(dokumenToDelete, "broadcast_dokumen")
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Broadcast berhasil dihapus",
	})
}

// ✅ READ - Mendapatkan semua broadcast
func (bc *BroadcastController) GetAllBroadcast(c *gin.Context) {
	var broadcast []models.Broadcast

	// Pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	// Search parameter
	search := strings.TrimSpace(c.Query("search"))

	// Build query dengan GORM (AMAN - parameterized queries)
	query := bc.db.Model(&models.Broadcast{})

	// Apply search filter jika ada
	if search != "" {
		query = query.Where("broadcast_nama LIKE ? OR broadcast_deskripsi LIKE ?", 
			"%"+search+"%", "%"+search+"%")
	}

	// Get total count for pagination
	var total int64
	query.Count(&total)

	// Execute query dengan pagination dan sorting
	if err := query.Offset(offset).
		// Limit(limit).
		Order("created_at DESC").
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data broadcast",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": broadcast,
		"pagination": gin.H{
			"page":  page,
			"limit": limit,
			"total": total,
		},
	})
}

// ✅ READ - Mendapatkan broadcast by ID
func (bc *BroadcastController) GetBroadcastByID(c *gin.Context) {
	id := c.Param("id")

	// Validasi ID (AMAN - dikonversi ke uint)
	broadcastID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "ID broadcast tidak valid",
		})
		return
	}

	var broadcast models.Broadcast
	if err := bc.db.First(&broadcast, broadcastID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Broadcast tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal mengambil data broadcast",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": broadcast,
	})
}


// ✅ GET - Mendapatkan broadcast terbaru
func (bc *BroadcastController) GetBroadcastTerbaru(c *gin.Context) {
	var broadcast []models.Broadcast

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "5"))

	// Query broadcast terbaru (AMAN)
	if err := bc.db.
		Order("created_at DESC").
		Limit(limit).
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mengambil data broadcast terbaru",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  broadcast,
		"total": len(broadcast),
	})
}

// ✅ GET - Pencarian broadcast
func (bc *BroadcastController) SearchBroadcast(c *gin.Context) {
	search := strings.TrimSpace(c.Query("q"))

	if search == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Parameter pencarian (q) harus diisi",
		})
		return
	}

	var broadcast []models.Broadcast

	// Execute search query dengan GORM (AMAN - parameterized)
	if err := bc.db.
		Where("broadcast_nama LIKE ? OR broadcast_deskripsi LIKE ?", 
			"%"+search+"%", "%"+search+"%").
		Order("created_at DESC").
		Limit(20).
		Find(&broadcast).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mencari broadcast",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":         broadcast,
		"total":        len(broadcast),
		"search_query": search,
	})
}

// ✅ GET - Statistik broadcast
func (bc *BroadcastController) GetStatistikBroadcast(c *gin.Context) {
	type StatistikResult struct {
		TotalBroadcast int64 `json:"total_broadcast"`
		BulanIni       int64 `json:"bulan_ini"`
		MingguIni      int64 `json:"minggu_ini"`
	}

	var statistik StatistikResult

	// Hitung total broadcast (AMAN)
	bc.db.Model(&models.Broadcast{}).Count(&statistik.TotalBroadcast)

	// Hitung broadcast bulan ini (AMAN)
	awalBulan := time.Now().AddDate(0, 0, -time.Now().Day()+1)
	bc.db.Model(&models.Broadcast{}).
		Where("created_at >= ?", awalBulan).
		Count(&statistik.BulanIni)

	// Hitung broadcast minggu ini (AMAN)
	awalMinggu := time.Now().AddDate(0, 0, -int(time.Now().Weekday())+1)
	bc.db.Model(&models.Broadcast{}).
		Where("created_at >= ?", awalMinggu).
		Count(&statistik.MingguIni)

	c.JSON(http.StatusOK, gin.H{
		"data": statistik,
	})
}



// ✅ GET - Serve file dokumen broadcast
func (bc *BroadcastController) GetBroadcastDokumen(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	file, err := helper.GetFileByFileName("broadcast_dokumen", filename)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File dokumen tidak ditemukan",
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

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendapatkan info file",
		})
		return
	}

	ext := filepath.Ext(filename)
	contentType := helper.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}

// ✅ GET - Serve file foto broadcast
func (bc *BroadcastController) GetBroadcastFoto(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file tidak valid",
		})
		return
	}

	file, err := helper.GetFileByFileName("broadcast_foto", filename)
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

	fileInfo, err := file.Stat()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal mendapatkan info file",
		})
		return
	}

	ext := filepath.Ext(filename)
	contentType := helper.GetContentType(ext)
	c.Header("Content-Type", contentType)
	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=\"%s\"", filename))

	http.ServeContent(c.Writer, c.Request, filename, fileInfo.ModTime(), file)
}