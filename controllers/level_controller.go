// controllers/level_controller.go
package controllers

import (
	"net/http"
	"rt-management/models"

	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

type LevelController struct {
	db *gorm.DB
}

func NewLevelController(db *gorm.DB) *LevelController {
	return &LevelController{db: db}
}

type CreateLevelRequest struct {
	LevelKode string `json:"level_kode" binding:"required"`
	LevelNama string `json:"level_nama" binding:"required"`
}

type UpdateLevelRequest struct {
	LevelKode string `json:"level_kode"`
	LevelNama string `json:"level_nama"`
}

func (lc *LevelController) GetAllLevels(c *gin.Context) {
	var levels []models.Level
	if err := lc.db.Find(&levels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch levels"})
		return
	}

	c.JSON(http.StatusOK, levels)
}

func (lc *LevelController) GetLevelByID(c *gin.Context) {
	levelID := c.Param("id")

	var level models.Level
	if err := lc.db.First(&level, levelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Level not found"})
		return
	}

	c.JSON(http.StatusOK, level)
}

func (lc *LevelController) CreateLevel(c *gin.Context) {
	var req CreateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if level kode already exists
	var existingLevel models.Level
	if err := lc.db.Where("level_kode = ?", req.LevelKode).First(&existingLevel).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Level kode already exists"})
		return
	}

	level := models.Level{
		LevelKode: req.LevelKode,
		LevelNama: req.LevelNama,
	}

	if err := lc.db.Create(&level).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create level"})
		return
	}

	c.JSON(http.StatusCreated, level)
}

func (lc *LevelController) UpdateLevel(c *gin.Context) {
	levelID := c.Param("id")

	var req UpdateLevelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var level models.Level
	if err := lc.db.First(&level, levelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Level not found"})
		return
	}

	if req.LevelKode != "" {
		// Check if level kode already exists (excluding current level)
		var existingLevel models.Level
		if err := lc.db.Where("level_kode = ? AND level_id != ?", req.LevelKode, levelID).First(&existingLevel).Error; err == nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Level kode already exists"})
			return
		}
		level.LevelKode = req.LevelKode
	}

	if req.LevelNama != "" {
		level.LevelNama = req.LevelNama
	}

	if err := lc.db.Save(&level).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update level"})
		return
	}

	c.JSON(http.StatusOK, level)
}

func (lc *LevelController) DeleteLevel(c *gin.Context) {
	levelID := c.Param("id")

	var level models.Level
	if err := lc.db.First(&level, levelID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Level not found"})
		return
	}

	// Check if level is being used by any user
	var userCount int64
	if err := lc.db.Model(&models.User{}).Where("level_id = ?", levelID).Count(&userCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check level usage"})
		return
	}

	if userCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete level that is being used by users"})
		return
	}

	if err := lc.db.Delete(&level).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete level"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Level deleted successfully"})
}