package handlers

import (
	"net/http"
	"time"
	"tours-service/database"
	"tours-service/models"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func CreateTourExecution(c *gin.Context) {
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

	tourIDStr := c.Param("tourId")
	tourID, err := uuid.Parse(tourIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tour ID"})
		return
	}

	if !IsTourAvailable(tourID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "tour is not available"})
		return
	}

	newExecution := models.TourExecution{
		UserID:         userId,
		TourID:         tourID,
		Status:         models.StatusInProgress,
		LastActivityAt: time.Now(),
	}

	if err := database.GORM_DB.Create(&newExecution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create tour execution"})
		return
	}

	c.JSON(http.StatusOK, newExecution)
}

func UpdateTourExecutionStatus(c *gin.Context) {
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

	tourExecutionIDStr := c.Param("tourExecutionId")
	tourExecutionID, err := uuid.Parse(tourExecutionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid tour execution ID"})
		return
	}

	var body struct {
		Status string `json:"status"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	newStatus, err := models.ParseTourExecutionStatus(body.Status)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var execution models.TourExecution
	if err := database.GORM_DB.Where("id = ? AND user_id = ? AND status = ?", tourExecutionID, userId, models.StatusInProgress).First(&execution).Error; err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "cannot update tour execution"})
		return
	}

	execution.Status = newStatus

	if err := database.GORM_DB.Save(&execution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tour execution"})
		return
	}

	c.JSON(http.StatusOK, execution)
}
