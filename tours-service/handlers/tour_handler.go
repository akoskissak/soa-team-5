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
	"tours-service/services"
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
	transportation := c.PostForm("transportation")

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

	transType := models.TransportationType(transportation)
	if transType != models.Walking && transType != models.Bicycle && transType != models.Car {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid transportation type"})
		return
	}

	tour := models.Tour{
		Name:           name,
		UserID:         userId,
		Description:    description,
		Difficulty:     models.TourDifficulty(difficulty),
		Transportation: transType,
		Tags:           datatypes.JSON([]byte(tagsStr)),
		Status:         models.Draft,
		Price:          0,
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
		if err := services.UpdateTourDistanceAndTimes(tour.ID); err != nil {
			log.Printf("Failed to update tour distance and times for tour %s: %v", tour.ID, err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Greška prilikom računanja distance i vremena.",
				"details": err.Error(),
			})
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

func GetTourKeypoints(tourID uuid.UUID) []models.KeyPoint {
	var keypoints []models.KeyPoint

	if err := database.GORM_DB.Where("tour_id = ?", tourID).Find(&keypoints).Error; err != nil {
		return []models.KeyPoint{}
	}

	return keypoints
}

func PublishTour(c *gin.Context) {
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userId, _ := claims["userId"].(string)

	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var tour models.Tour
	if err := database.GORM_DB.First(&tour, "id = ?", tourId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	if tour.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the author of this tour"})
		return
	}

	if tour.Status != models.Draft {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour can only be published from Draft status"})
		return
	}

	if tour.Name == "" || tour.Description == "" || string(tour.Difficulty) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have name, description and difficulty"})
		return
	}
	if len(tour.Tags) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have at least one tag"})
		return
	}

	var keypointCount int64
	database.GORM_DB.Model(&models.KeyPoint{}).Where("tour_id = ?", tourId).Count(&keypointCount)
	if keypointCount < 2 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tour must have at least two keypoints"})
		return
	}

	var requiredTimes []models.RequiredTime
	database.GORM_DB.Where("tour_id = ?", tourId).Find(&requiredTimes)
	if len(requiredTimes) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one required time must be defined"})
		return
	}
	for _, rt := range requiredTimes {
		if rt.TimeInMinutes <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Required time must be greater than 0"})
			return
		}
	}

	tour.Status = models.Published
	now := time.Now()
	tour.PublishedAt = &now

	if err := database.GORM_DB.Save(&tour).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to publish tour"})
		return
	}

	c.JSON(http.StatusOK, tour)
}

func ArchiveTour(c *gin.Context) {
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userId, _ := claims["userId"].(string)

	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var tour models.Tour
	if err := database.GORM_DB.First(&tour, "id = ?", tourId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	if tour.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the author of this tour"})
		return
	}

	if tour.Status != models.Published {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only published tours can be archived"})
		return
	}

	tour.Status = models.Archived

	now := time.Now()
	tour.ArchivedAt = &now

	if err := database.GORM_DB.Save(&tour).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to archive tour"})
		return
	}

	c.JSON(http.StatusOK, tour)
}

func UnarchiveTour(c *gin.Context) {
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	userId, _ := claims["userId"].(string)

	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var tour models.Tour
	if err := database.GORM_DB.First(&tour, "id = ?", tourId).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tour not found"})
		return
	}

	if tour.UserID != userId {
		c.JSON(http.StatusForbidden, gin.H{"error": "You are not the author of this tour"})
		return
	}

	if tour.Status != models.Archived {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only archived tours can be unarchived"})
		return
	}

	tour.Status = models.Published
	now := time.Now()
	tour.PublishedAt = &now

	if err := database.GORM_DB.Save(&tour).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to unarchive tour"})
		return
	}

	c.JSON(http.StatusOK, tour)
}
