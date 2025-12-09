package helper

import (
	"fmt"
	"time"
	"math/rand"
)

func GenerateKodeTransaksi() string {
	// Format: TRX + 7 digit angka unik (total 10 karakter)
	now := time.Now()
	
	// Bagian 1: Ambil 3 digit dari timestamp (detik + milidetik)
	timestampPart := now.Unix() % 1000
	
	// Bagian 2: Generate random 4 digit
	rand.Seed(now.UnixNano())
	randomPart := rand.Intn(10000)
	
	// Gabungkan dan format ke 7 digit
	codeNumber := (timestampPart * 10000) + int64(randomPart)
	codeNumber = codeNumber % 10000000 // Pastikan 7 digit
	
	// Format: TRX0000000 (10 karakter total)
	return fmt.Sprintf("TRX%07d", codeNumber)
}