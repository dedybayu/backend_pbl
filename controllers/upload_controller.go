// upload_controller.go
package controllers

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadController struct {
	uploadPath string
}

func NewUploadController() *UploadController {
	// Buat direktori upload jika belum ada
	uploadPath := "storage/images/produk/"
	os.MkdirAll(uploadPath, 0755)
	
	return &UploadController{
		uploadPath: uploadPath,
	}
}

// UploadFile - Menangani upload file
func (uc *UploadController) UploadFile(c *gin.Context) {
	// Parse multipart form
	err := c.Request.ParseMultipartForm(10 << 20) // 10 MB max
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal parse form data",
		})
		return
	}

	// Get file dari form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "File tidak ditemukan dalam form data",
		})
		return
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Gagal membaca file",
		})
		return
	}

	fileType := http.DetectContentType(buffer)
	if !allowedTypes[fileType] {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Tipe file tidak diizinkan. Hanya JPEG, JPG, PNG, GIF, dan WEBP yang diperbolehkan",
		})
		return
	}

	// Kembali ke awal file
	file.Seek(0, 0)

	// Generate nama file unik
	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("produk_%s%s", timestamp, ext)
	filePath := filepath.Join(uc.uploadPath, uniqueFilename)

	// Buat file baru
	out, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal membuat file",
		})
		return
	}
	defer out.Close()

	// Copy file ke storage
	_, err = io.Copy(out, file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Gagal menyimpan file",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "File berhasil diupload",
		"filename": uniqueFilename,
		"filepath": filePath,
		"url":      "/" + filePath, // URL untuk diakses
	})
}

// DeleteFile - Menghapus file
func (uc *UploadController) DeleteFile(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file harus diisi",
		})
		return
	}

	// Security: hanya boleh menghapus file di direktori produk
	filePath := filepath.Join(uc.uploadPath, filename)
	if !strings.HasPrefix(filePath, uc.uploadPath) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Path file tidak valid",
		})
		return
	}

	err := os.Remove(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "File tidak ditemukan",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Gagal menghapus file",
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "File berhasil dihapus",
	})
}

// GetFile - Mengambil file (serve static file)
func (uc *UploadController) GetFile(c *gin.Context) {
	filename := c.Param("filename")
	
	if filename == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Nama file harus diisi",
		})
		return
	}

	filePath := filepath.Join(uc.uploadPath, filename)
	
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "File tidak ditemukan",
		})
		return
	}

	c.File(filePath)
}

// Tambahkan method ini ke UploadController

// UploadFileDirect - Upload file tanpa HTTP context (untuk dipanggil dari controller lain)
func (uc *UploadController) UploadFileDirect(file interface{}, header interface{}) (string, error) {
	// Type assertion untuk file dan header
	fileReader, ok := file.(io.Reader)
	if !ok {
		return "", fmt.Errorf("invalid file type")
	}

	fileHeader, ok := header.(*multipart.FileHeader)
	if !ok {
		return "", fmt.Errorf("invalid file header type")
	}

	// Validasi tipe file
	buffer := make([]byte, 512)
	_, err := fileReader.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file")
	}

	fileType := http.DetectContentType(buffer)
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !allowedTypes[fileType] {
		return "", fmt.Errorf("tipe file tidak diizinkan")
	}

	// Kembali ke awal file
	if seeker, ok := fileReader.(io.Seeker); ok {
		seeker.Seek(0, 0)
	}

	// Generate nama file unik
	ext := filepath.Ext(fileHeader.Filename)
	timestamp := time.Now().Format("20060102150405")
	uniqueFilename := fmt.Sprintf("produk_%s%s", timestamp, ext)
	filePath := filepath.Join(uc.uploadPath, uniqueFilename)

	// Buat file baru
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file")
	}
	defer out.Close()

	// Copy file ke storage
	_, err = io.Copy(out, fileReader)
	if err != nil {
		return "", fmt.Errorf("gagal menyimpan file")
	}

	return uniqueFilename, nil
}

// DeleteFileByName - Menghapus file by nama
func (uc *UploadController) DeleteFileByName(filename string) error {
	if filename == "" {
		return fmt.Errorf("nama file harus diisi")
	}

	filePath := filepath.Join(uc.uploadPath, filename)
	if !strings.HasPrefix(filePath, uc.uploadPath) {
		return fmt.Errorf("path file tidak valid")
	}

	err := os.Remove(filePath)
	if err != nil {
		return err
	}

	return nil
}