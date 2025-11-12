package helper

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	// "golang.org/x/mod/sumdb/storage"
)

// Helper function untuk membuat direktori jika belum ada
func EnsureDirectory(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, 0755)
	}
	return nil
}

// Helper function untuk menghapus file foto lama
func DeleteOldPhoto(filename string, fieldName string) error {
	if filename != "" {
		fullPath := filepath.Join("storage", "images", "produk", filename)

		switch fieldName {
		case "produk_foto":
			fullPath = filepath.Join("storage", "images", "produk", filename)
		case "pengeluaran_bukti":
			fullPath = filepath.Join("storage", "images", "pengeluaran", filename)
		case "pemasukan_bukti":
			fullPath = filepath.Join("storage", "images", "pemasukan", filename)
		case "broadcast_foto":
			fullPath = filepath.Join("storage", "images", "broadcast", filename)
		default:
		}

		// Langsung gabung dengan base directory

		// Cek apakah file ada sebelum menghapus
		if _, err := os.Stat(fullPath); err == nil {
			return os.Remove(fullPath)
		}
		// Jika file tidak ditemukan, tidak perlu error
	}
	return nil
}

// Helper function untuk handle file upload (diekspor dengan huruf besar)
func HandleFileImageUpload(c *gin.Context, fieldName string, oldPhotoPath string) (string, error) {
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
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("gagal reset file pointer: %v", err)
	}

	var storageDir string
	var filePrefix string

	switch fieldName {
	case "produk_foto":
		storageDir = "storage/images/produk"
		filePrefix = "produk"
	case "pengeluaran_bukti":
		storageDir = "storage/images/pengeluaran"
		filePrefix = "pengeluaran"
	case "pemasukan_bukti":
		storageDir = "storage/images/pemasukan"
		filePrefix = "pemasukan"
	case "broadcast_foto":
		storageDir = "storage/images/broadcast"
		filePrefix = "broadcast"
	default:
		storageDir = "storage/images/default"
		filePrefix = "default"
	}

	// Buat nama file unik
	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().Format("20060102150405")
	randomStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	filename := fmt.Sprintf("%s_%s_%s%s",filePrefix, timestamp, randomStr[len(randomStr)-6:], ext)

	if err := EnsureDirectory(storageDir); err != nil {
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
		DeleteOldPhoto(oldPhotoPath, fieldName)
	}

	return filename, nil
}

// Helper function untuk menghapus file dokumen lama
func DeleteOldDocument(filename string, fieldName string) error {
	if filename != "" {
		var fullPath string

		switch fieldName {
		case "broadcast_dokumen":
			fullPath = filepath.Join("storage", "dokumen", "broadcast", filename)
		default:
			fullPath = filepath.Join("storage", "dokumen", filename)
		}

		// Cek apakah file ada sebelum menghapus
		if _, err := os.Stat(fullPath); err == nil {
			return os.Remove(fullPath)
		}
		// Jika file tidak ditemukan, tidak perlu error
	}
	return nil
}

// Helper function untuk handle file upload dokumen (diekspor dengan huruf besar)
func HandleFileDokumenUpload(c *gin.Context, fieldName string, oldDocumentPath string) (string, error) {
	file, header, err := c.Request.FormFile(fieldName)
	if err != nil {
		if err == http.ErrMissingFile {
			// Jika tidak ada file baru diupload, return path dokumen lama
			return oldDocumentPath, nil
		}
		return "", err
	}
	defer file.Close()

	// Validasi tipe file untuk DOKUMEN
	allowedTypes := map[string]bool{
		// Document types
		"application/pdf":    true,
		"application/msword": true,
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": true,
		"application/vnd.ms-excel": true,
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
		"application/vnd.ms-powerpoint":                                             true,
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": true,
		"text/plain":                   true,
		"text/csv":                     true,
		"application/zip":              true,
		"application/x-rar-compressed": true,
	}

	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return "", fmt.Errorf("gagal membaca file: %v", err)
	}

	contentType := http.DetectContentType(buffer)
	if !allowedTypes[contentType] {
		return "", fmt.Errorf("tipe file tidak diizinkan. Hanya PDF, DOC, DOCX, XLS, XLSX, PPT, PPTX, TXT, CSV, ZIP, RAR yang diperbolehkan")
	}

	// Kembali ke awal file
	_, err = file.Seek(0, 0)
	if err != nil {
		return "", fmt.Errorf("gagal reset file pointer: %v", err)
	}

	// Tentukan direktori penyimpanan dan prefix filename berdasarkan fieldName
	var storageDir, filePrefix string

	switch fieldName {
	case "broadcast_dokumen":
		storageDir = "storage/dokumen/broadcast"
		filePrefix = "broadcast"
	default:
		storageDir = "storage/dokumen/umum"
		filePrefix = "dokumen"
	}

	// Buat nama file unik dengan ekstensi asli
	ext := filepath.Ext(header.Filename)
	timestamp := time.Now().Format("20060102150405")
	randomStr := strconv.FormatInt(time.Now().UnixNano(), 10)
	filename := fmt.Sprintf("%s_%s_%s%s", filePrefix, timestamp, randomStr[len(randomStr)-6:], ext)

	// Buat direktori jika belum ada
	if err := EnsureDirectory(storageDir); err != nil {
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

	// Hapus dokumen lama jika ada file baru diupload dan dokumen lama ada
	if oldDocumentPath != "" {
		DeleteOldDocument(oldDocumentPath, fieldName)
	}

	return filename, nil
}

// Helper function untuk mendapatkan ekstensi file berdasarkan content type
func getFileExtension(contentType string) string {
	extensionMap := map[string]string{
		"application/pdf":    ".pdf",
		"application/msword": ".doc",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
		"application/vnd.ms-excel": ".xls",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         ".xlsx",
		"application/vnd.ms-powerpoint":                                             ".ppt",
		"application/vnd.openxmlformats-officedocument.presentationml.presentation": ".pptx",
		"text/plain":                   ".txt",
		"text/csv":                     ".csv",
		"application/zip":              ".zip",
		"application/x-rar-compressed": ".rar",
	}

	if ext, exists := extensionMap[contentType]; exists {
		return ext
	}
	return ""
}

// Helper function untuk mendapatkan mime type berdasarkan ekstensi file
func getMimeType(extension string) string {
	mimeMap := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".csv":  "text/csv",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
	}

	if mime, exists := mimeMap[extension]; exists {
		return mime
	}
	return "application/octet-stream"
}

// Helper function untuk menentukan Content-Type berdasarkan ekstensi
func GetContentType(ext string) string {
	switch strings.ToLower(ext) {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".doc":
		return "application/msword"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	case ".xls":
		return "application/vnd.ms-excel"
	case ".xlsx":
		return "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
	case ".txt":
		return "text/plain"
	default:
		return "application/octet-stream"
	}
}


// GetFileByFileName - Mendapatkan file berdasarkan fieldName dan fileName
func GetFileByFileName(fieldName string, fileName string) (*os.File, error) {
	var storageDir string

	switch fieldName {
	case "produk_foto":
		storageDir = "storage/images/produk"
	case "pengeluaran_bukti":
		storageDir = "storage/images/pengeluaran" // Fixed typo: pegeluaran -> pengeluaran
	case "pemasukan_bukti":
		storageDir = "storage/images/pemasukan"
	case "broadcast_foto":
		storageDir = "storage/images/broadcast"
	case "broadcast_dokumen":
		storageDir = "storage/dokumen/broadcast"
	case "pengeluaran_dokumen":
		storageDir = "storage/dokumen/pengeluaran"
	case "pemasukan_dokumen":
		storageDir = "storage/dokumen/pemasukan"
	case "produk_dokumen":
		storageDir = "storage/dokumen/produk"
	default:
		storageDir = "storage/images/default"
	}

	filePath := filepath.Join(storageDir, fileName)
	
	// Validasi path untuk mencegah directory traversal
	cleanPath := filepath.Clean(filePath)
	expectedPrefix := filepath.Clean(storageDir)
	if !strings.HasPrefix(cleanPath, expectedPrefix) {
		return nil, fmt.Errorf("path file tidak valid")
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	return file, nil
}