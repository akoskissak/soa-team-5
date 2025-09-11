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

func CreateReview(c *gin.Context) {
	claims, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	touristId, ok := claims["userId"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId in token"})
		return
	}

	username, ok := claims["username"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid username in token"})
		return
	}

	tourIdStr := c.PostForm("tourId")
	ratingStr := c.PostForm("rating")
	comment := c.PostForm("comment")
	visitedDateStr := c.PostForm("visitedDate")

	if tourIdStr == "" || ratingStr == "" || visitedDateStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "tourId, rating, and visitedDate are required"})
		return
	}

	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tourId format"})
		return
	}

	rating := 0
	if _, err := fmt.Sscanf(ratingStr, "%d", &rating); err != nil || rating < 1 || rating > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid rating format or value (1-5)"})
		return
	}

	visitedDate, err := time.Parse("2006-01-02", visitedDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid visited date format, use YYYY-MM-DD"})
		return
	}

	review := models.Review{
		TourID:         tourId,
		TouristID:      touristId,
		Username:       username,
		Rating:         rating,
		Comment:        comment,
		SubmissionDate: time.Now(),
		VisitedDate:    visitedDate,
	}

	if err := database.GORM_DB.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save review"})
		return
	}

	form, _ := c.MultipartForm()
	files := form.File["images"]

	var reviewImages []models.ReviewImage
	for _, file := range files {
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), filepath.Base(file.Filename))

		savePath := filepath.Join("static", "uploads", filename)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save review image"})
			return
		}
		imagePath := "/uploads/" + filename

		reviewImages = append(reviewImages, models.ReviewImage{
			ImagePath: imagePath,
			ReviewID:  review.ID,
		})
	}

	if len(reviewImages) > 0 {
		if err := database.GORM_DB.Create(&reviewImages).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save review images"})
			return
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "Review created successfully",
		"review":  review,
		"images":  reviewImages,
	})
}

func GetReviewsByTourId(c *gin.Context) {
	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var reviews []models.Review
	if err := database.GORM_DB.Where("tour_id = ?", tourId).Preload("ReviewImages").Find(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch reviews"})
		return
	}

	c.JSON(http.StatusOK, reviews)
}
