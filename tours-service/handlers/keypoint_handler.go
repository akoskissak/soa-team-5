package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"time"
	"tours-service/database"
	"tours-service/models"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateKeyPoint(c *gin.Context) {
	_, err := utils.VerifyJWT(c)
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
	keypoint := models.KeyPoint{
		Name:        input.Name,
		Description: input.Description,
		Latitude:    input.Latitude,
		Longitude:   input.Longitude,
		ImagePath:   savePath,
		TourID:      tourID,
	}

	if err := database.GORM_DB.Create(&keypoint).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save keypoint"})
		return
	}

	c.JSON(http.StatusCreated, keypoint)
}