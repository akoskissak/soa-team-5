package handlers

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"tours-service/database"
	"tours-service/models"
	"tours-service/services"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateKeyPoint(c *gin.Context) {
	_, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var input struct {
		Name        string  `form:"name" binding:"required"`
		Description string  `form:"description"`
		Latitude    float64 `form:"latitude" binding:"required"`
		Longitude   float64 `form:"longitude" binding:"required"`
		TourID      string  `form:"tourId" binding:"required"`
	}

	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	file, err := c.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image is required"})
		return
	}

	filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
	savePath := filepath.Join("static/uploads", filename)

	if err := c.SaveUploadedFile(file, savePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
		return
	}

	tourID, _ := uuid.Parse(input.TourID)
	var maxPosition int
	database.GORM_DB.Model(&models.KeyPoint{}).Where("tour_id = ?", tourID).Select("COALESCE(MAX(position), -1)").Row().Scan(&maxPosition)

	keypoint := models.KeyPoint{
		Name:        input.Name,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		ImagePath:   savePath,
		TourID:      tourID,
		Position:    maxPosition + 1,
	}

	if err := database.GORM_DB.Create(&keypoint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save keypoint"})
		return
	}

	go func(tourID uuid.UUID) {
		if err := services.UpdateTourDistanceAndTimes(tourID); err != nil {
			log.Printf("Failed to update tour distance for tour %s: %v", tourID, err)
		}
	}(tourID)

	c.JSON(http.StatusCreated, keypoint)
}

func GetKeyPointsByTourId(c *gin.Context) {
	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var keyPoints []models.KeyPoint
	if err := database.GORM_DB.Where("tour_id = ?", tourId).Find(&keyPoints).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch keypoints"})
		return
	}

	c.JSON(http.StatusOK, keyPoints)
}

func UpdateKeyPoint(c *gin.Context) {
	keyPointIDStr := c.Param("id")
	keyPointID, err := uuid.Parse(keyPointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keypoint ID"})
		return
	}

	var keyPointToUpdate models.KeyPoint
	if err := database.GORM_DB.First(&keyPointToUpdate, keyPointID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keypoint not found"})
		return
	}

	var input struct {
		Name        string  `form:"name"`
		Description string  `form:"description"`
		Latitude    float64 `form:"latitude"`
		Longitude   float64 `form:"longitude"`
	}

	if err := c.ShouldBind(&input); err == nil && input.Name != "" {
		keyPointToUpdate.Name = input.Name
		keyPointToUpdate.Description = input.Description
		keyPointToUpdate.Latitude = input.Latitude
		keyPointToUpdate.Longitude = input.Longitude

		if file, err := c.FormFile("image"); err == nil {
			filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
			savePath := filepath.Join("static/uploads", filename)

			if err := c.SaveUploadedFile(file, savePath); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save image"})
				return
			}

			keyPointToUpdate.ImagePath = savePath
		}
	} else {
		var updatedKeyPoint models.KeyPoint
		if err := c.ShouldBindJSON(&updatedKeyPoint); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		keyPointToUpdate.Name = updatedKeyPoint.Name
		keyPointToUpdate.Description = updatedKeyPoint.Description
		keyPointToUpdate.Latitude = updatedKeyPoint.Latitude
		keyPointToUpdate.Longitude = updatedKeyPoint.Longitude
	}

	if err := database.GORM_DB.Save(&keyPointToUpdate).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update keypoint"})
		return
	}

	tourID := keyPointToUpdate.TourID
	go func(tourID uuid.UUID) {
		if err := services.UpdateTourDistanceAndTimes(tourID); err != nil {
			log.Printf("Failed to update tour distance for tour %s: %v", tourID, err)
		}
	}(tourID)

	c.JSON(http.StatusOK, keyPointToUpdate)
}
func DeleteKeyPoint(c *gin.Context) {
	keyPointIDStr := c.Param("id")
	keyPointID, err := uuid.Parse(keyPointIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid keypoint ID"})
		return
	}

	var keyPoint models.KeyPoint
	if err := database.GORM_DB.First(&keyPoint, "id = ?", keyPointID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Keypoint not found"})
		return
	}
	tourID := keyPoint.TourID

	if err := database.GORM_DB.Delete(&models.KeyPoint{}, keyPointID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete keypoint"})
		return
	}

	go func(tourID uuid.UUID) {
		if err := services.UpdateTourDistanceAndTimes(tourID); err != nil {
			log.Printf("Failed to update tour distance for tour %s: %v", tourID, err)
		}
	}(tourID)

	c.JSON(http.StatusOK, gin.H{"message": "Keypoint deleted successfully"})
}
