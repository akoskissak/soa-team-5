package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"path/filepath"
	"time"
	"tours-service/database"
	"tours-service/models"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/datatypes"
)

func CreateTour(c *gin.Context) {
	claims,  err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId in token"})
		return
	}

	name := c.PostForm("name")
	description := c.PostForm("description")
	difficulty := c.PostForm("difficulty")
	tagsStr := c.PostForm("tags")

	if name == "" || description == "" || difficulty == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "name, description and difficulty are required",
		})
		return
	}

	var tags []string
	if err := json.Unmarshal([]byte(tagsStr), &tags); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tags"})
		return
	}

	tour := models.Tour{
		Name:        name,
		UserID:      userId,
		Description: description,
		Difficulty:  models.TourDifficulty(difficulty),
		Tags:        datatypes.JSON([]byte(tagsStr)),
		Status:      models.Draft,
		Price:       0,
	}

	if err := database.GORM_DB.Create(&tour).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save tour"})
		return
	}

	var keypointsInput []struct {
		Name        string   `json:"name"`
		Description string   `json:"description"`
		Latitude    *float64 `json:"latitude"`
		Longitude   *float64 `json:"longitude"`
	}

	if err := json.Unmarshal([]byte(c.PostForm("keypoints")), &keypointsInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid keypoints"})
		return
	}

	var keypoints []models.KeyPoint
	for i, kp := range keypointsInput {
		file, _ := c.FormFile(fmt.Sprintf("image%d", i))

		if kp.Name == "" || kp.Latitude == nil || kp.Longitude == nil || file == nil {
			continue
		}

		var imagePath string
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))
		savePath := filepath.Join("static/uploads", filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save keypoint image"})
			return
		}
		imagePath = "/" + savePath

		keypoints = append(keypoints, models.KeyPoint{
			Name:        kp.Name,
			Description: kp.Description,
			Latitude:    *kp.Latitude,
			Longitude:   *kp.Longitude,
			ImagePath:   imagePath,
			TourID:      tour.ID,
		})
	}

	if len(keypoints) > 0 {
		if err := database.GORM_DB.Create(&keypoints).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save keypoints"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"tour":      tour,
		"keypoints": keypoints,
	})
}

func GetAllTours(c *gin.Context) {
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId in token"})
		return
	}

	var tours []models.Tour

	if err := database.GORM_DB.Where("user_id = ?", userId).Find(&tours).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}

func GetAllPublishedTours(c *gin.Context) {
	var tours []models.Tour

	if err := database.GORM_DB.Where("status = ?", models.Published).Find(&tours).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch published tours"})
		return
	}

	c.JSON(http.StatusOK, tours)
}
