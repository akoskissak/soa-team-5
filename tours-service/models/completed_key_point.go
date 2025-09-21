package models

import (
	"time"

	"github.com/google/uuid"
)

type CompletedKeyPoint struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	TourExecutionID uuid.UUID `gorm:"type:uuid;not null;index" json:"tourExecutionId"`
	KeyPointID      uuid.UUID `gorm:"type:uuid;not null" json:"keyPointId"`
	CompletedAt     time.Time `gorm:"not null" json:"completedAt"`
}
