package models

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type TourExecutionStatus string

const (
	StatusInProgress TourExecutionStatus = "in_progress"
	StatusCompleted  TourExecutionStatus = "completed"
	StatusAbandoned  TourExecutionStatus = "abandoned"
)

type TourExecution struct {
	ID                 uuid.UUID           `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	UserID             string              `gorm:"type:varchar(24);not null;column:user_id" json:"userId"`
	TourID             uuid.UUID           `gorm:"type:uuid;not null;column:tour_id" json:"tourId"`
	Status             TourExecutionStatus `gorm:"type:varchar(20);default:'in_progress'" json:"status"`
	LastActivityAt     time.Time           `gorm:"not null;autoUpdateTime" json:"lastActivityAt"`
	CompletedKeyPoints []CompletedKeyPoint `gorm:"foreignKey:TourExecutionID" json:"completedKeyPoints"`
	CreatedAt          time.Time           `gorm:"autoCreateTime" json:"createdAt"`
}

func ParseTourExecutionStatus(statusStr string) (TourExecutionStatus, error) {
	switch statusStr {
	case string(StatusInProgress):
		return StatusInProgress, nil
	case string(StatusCompleted):
		return StatusCompleted, nil
	case string(StatusAbandoned):
		return StatusAbandoned, nil
	default:
		return "", fmt.Errorf("invalid status: %s", statusStr)
	}
}