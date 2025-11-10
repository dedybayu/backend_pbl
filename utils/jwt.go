// utils/jwt.go
package utils

import (
	"errors"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type JWTClaims struct {
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	LevelID  uint   `json:"level_id"`
	jwt.RegisteredClaims
}

type JWTUtils struct {
	secretKey string
	blacklist *sync.Map // Untuk menyimpan token yang di-blacklist
	mutex     sync.RWMutex
}

func NewJWTUtils(secretKey string) *JWTUtils {
	return &JWTUtils{
		secretKey: secretKey,
		blacklist: &sync.Map{},
	}
}

func (j *JWTUtils) GenerateToken(userID uint, username string, levelID uint) (string, error) {
	claims := JWTClaims{
		UserID:   userID,
		Username: username,
		LevelID:  levelID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

func (j *JWTUtils) ValidateToken(tokenString string) (*JWTClaims, error) {
	// Cek apakah token di-blacklist terlebih dahulu
	if j.IsBlacklisted(tokenString) {
		return nil, errors.New("token has been invalidated")
	}

	token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(j.secretKey), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// AddToBlacklist menambahkan token ke blacklist
func (j *JWTUtils) AddToBlacklist(tokenString string) error {
	// Parse token untuk mendapatkan expiry time tanpa validasi signature
	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &JWTClaims{})
	if err != nil {
		return err
	}

	if claims, ok := token.Claims.(*JWTClaims); ok {
		expiryTime := claims.ExpiresAt.Time
		j.blacklist.Store(tokenString, expiryTime)
		
		// Schedule cleanup ketika token expired
		go j.scheduleCleanup(tokenString, expiryTime)
		return nil
	}

	return errors.New("failed to parse token claims")
}

// IsBlacklisted mengecek apakah token di-blacklist
func (j *JWTUtils) IsBlacklisted(tokenString string) bool {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	expiry, exists := j.blacklist.Load(tokenString)
	if !exists {
		return false
	}

	// Jika token sudah expired, hapus dari blacklist
	if time.Now().After(expiry.(time.Time)) {
		j.mutex.RUnlock()
		j.mutex.Lock()
		j.blacklist.Delete(tokenString)
		j.mutex.Unlock()
		j.mutex.RLock()
		return false
	}

	return true
}

// scheduleCleanup menjadwalkan pembersihan token dari blacklist ketika expired
func (j *JWTUtils) scheduleCleanup(tokenString string, expiryTime time.Time) {
	duration := time.Until(expiryTime)
	if duration > 0 {
		time.AfterFunc(duration, func() {
			j.mutex.Lock()
			defer j.mutex.Unlock()
			j.blacklist.Delete(tokenString)
		})
	}
}

// CleanupExpiredTokens membersihkan semua token yang sudah expired dari blacklist
func (j *JWTUtils) CleanupExpiredTokens() {
	j.mutex.Lock()
	defer j.mutex.Unlock()

	j.blacklist.Range(func(key, value interface{}) bool {
		if time.Now().After(value.(time.Time)) {
			j.blacklist.Delete(key)
		}
		return true
	})
}

// GetBlacklistSize mengembalikan jumlah token di blacklist (untuk debugging)
func (j *JWTUtils) GetBlacklistSize() int {
	j.mutex.RLock()
	defer j.mutex.RUnlock()

	count := 0
	j.blacklist.Range(func(_, _ interface{}) bool {
		count++
		return true
	})
	return count
}

// RemoveFromBlacklist menghapus token dari blacklist (untuk testing)
func (j *JWTUtils) RemoveFromBlacklist(tokenString string) {
	j.mutex.Lock()
	defer j.mutex.Unlock()
	j.blacklist.Delete(tokenString)
}