package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"time"
	"tours-service/database"
	"tours-service/models"
	"tours-service/opentelemetery"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/datatypes"
)

func CreateTour(c *gin.Context) {
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
	traceContext, span := opentelemetery.TraceProvider.Tracer(opentelemetery.ServiceName).Start(c, "tours-get-all")
	defer func() { span.End() }()

	span.AddEvent("Getting claims from gin context")
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		httpErrorUnathorized(err, span, c)
		return
	}

	userId, ok := claims["userId"].(string)
	if !ok {
		httpErrorBadRequest(err, span, c)
		return
	}

	span.AddEvent("Retrieving tours from the database")
	var tours []models.Tour

	if err := database.GORM_DB.WithContext(traceContext).Where("user_id = ?", userId).Find(&tours).Error; err != nil {
		httpErrorInternalServerError(err, span, c)
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

func IsTourAvailable(id uuid.UUID) bool {
	var tour models.Tour
	if err := database.GORM_DB.First(&tour, "id = ?", id).Error; err != nil {
		return false
	}

	return tour.Status == models.Published || tour.Status == models.Archived
}

// Tracing error messages
func httpErrorBadRequest(err error, span trace.Span, c *gin.Context) {
	httpError(err, span, c, http.StatusBadRequest)
}

func httpErrorUnathorized(err error, span trace.Span, c *gin.Context) {
	httpError(err, span, c, http.StatusUnauthorized)
}

func httpErrorInternalServerError(err error, span trace.Span, c *gin.Context) {
	httpError(err, span, c, http.StatusInternalServerError)
}

func httpError(err error, span trace.Span, c *gin.Context, status int) {
	log.Println(err.Error())
	span.RecordError(err)
	span.SetStatus(codes.Error, err.Error())
	c.String(status, err.Error())
}
