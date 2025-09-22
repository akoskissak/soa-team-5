package handlers

import (
	"net/http"
	"time"
	"tours-service/database"
	"tours-service/rest_clients"
	"tours-service/models"
	"tours-service/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var purchaseClient *rest_clients.PurchaseClient

func InitPurchaseClient(baseURL string) {
	purchaseClient = rest_clients.NewPurchaseClient(baseURL)
}

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

	purchased, err := purchaseClient.HasPurchasedTour(userId, tourID.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify purchase"})
		return
	}
	if !purchased {
		c.JSON(http.StatusForbidden, gin.H{"error": "you have not purchased this tour"})
		return
	}

	if !IsTourAvailable(tourID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "tour is not available"})
		return
	}

	var existingExecution models.TourExecution
	err = database.GORM_DB.
		Where("user_id = ? AND tour_id = ? AND status = ?", userId, tourID, models.StatusInProgress).
		First(&existingExecution).Error

	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "you already have an in-progress execution for this tour"})
		return
	} else if err != gorm.ErrRecordNotFound {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing executions"})
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

	execution.LastActivityAt = time.Now()
	execution.Status = newStatus

	if err := database.GORM_DB.Save(&execution).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update tour execution"})
		return
	}

	c.JSON(http.StatusOK, execution)
}

func GetAllMyTourExecutions(c *gin.Context) {
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

	var executions []models.TourExecution
	if err := database.GORM_DB.Where("user_id = ?", userId).Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch executions"})
		return
	}

	c.JSON(http.StatusOK, executions)
}

func CheckTourLocation(c *gin.Context) {
	_, err := utils.GetClaimsFromGinContext2Args(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	executionIDStr := c.Param("tourExecutionId")
	executionID, err := uuid.Parse(executionIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid execution ID"})
		return
	}

	var body struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
	}
	if err := c.BindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	var execution models.TourExecution
	if err := database.GORM_DB.Preload("CompletedKeyPoints").First(&execution, "id = ?", executionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "execution not found"})
		return
	}

	checkpoints := GetTourKeypoints(execution.TourID)

	var newlyCompleted []models.CompletedKeyPoint
	for _, cp := range checkpoints {
		if IsNearby(body.Latitude, body.Longitude, cp.Latitude, cp.Longitude) {
			alreadyCompleted := false
			for _, completed := range execution.CompletedKeyPoints {
				if completed.KeyPointID == cp.ID {
					alreadyCompleted = true
					break
				}
			}

			if !alreadyCompleted {
				newCKeyPoint := models.CompletedKeyPoint{
					ID:              uuid.New(),
					TourExecutionID: execution.ID,
					KeyPointID:      cp.ID,
					CompletedAt:     time.Now(),
				}
				database.GORM_DB.Create(&newCKeyPoint)
				execution.CompletedKeyPoints = append(execution.CompletedKeyPoints, newCKeyPoint)

				newlyCompleted = append(newlyCompleted, newCKeyPoint)
			}
		}
	}

	execution.LastActivityAt = time.Now()
	if len(execution.CompletedKeyPoints) == len(checkpoints) && execution.Status != models.StatusCompleted {
		execution.Status = models.StatusCompleted
		if err := database.GORM_DB.Save(&execution).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update execution status"})
			return
		}
	} else {
		database.GORM_DB.Save(&execution)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":            "location checked",
		"newlyCompleted":     newlyCompleted,
		"completedKeyPoints": execution.CompletedKeyPoints,
		"status":             execution.Status,
	})
}

func IsNearby(lat1, lon1, lat2, lon2 float64) bool {
	const delta = 0.0001 // 11metara blizu
	latDiff := lat1 - lat2
	lonDiff := lon1 - lon2

	if latDiff < 0 {
		latDiff = -latDiff
	}
	if lonDiff < 0 {
		lonDiff = -lonDiff
	}

	return latDiff <= delta && lonDiff <= delta
}

func GetActiveTourExecution(c *gin.Context) {
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

	var execution models.TourExecution
	err = database.GORM_DB.Preload("CompletedKeyPoints").
		Where("user_id = ? AND status = ?", userId, models.StatusInProgress).
		First(&execution).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusOK, nil)
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch active execution"})
		return
	}

	c.JSON(http.StatusOK, execution)
}
