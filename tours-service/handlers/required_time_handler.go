package handlers

import (
	"net/http"
	"tours-service/database"
	"tours-service/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateRequiredTime(c *gin.Context) {
	tourIdStr := c.Param("tourId")
	tourId, err := uuid.Parse(tourIdStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid tour ID"})
		return
	}

	var input struct {
		Transportation models.TransportationType `json:"transportation" binding:"required"`
		TimeInMinutes  int                       `json:"timeInMinutes" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	requiredTime := models.RequiredTime{
		TourID:         tourId,
		Transportation: input.Transportation,
		TimeInMinutes:  input.TimeInMinutes,
	}

	if err := database.GORM_DB.Create(&requiredTime).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save required time"})
		return
	}

	c.JSON(http.StatusCreated, requiredTime)
}
