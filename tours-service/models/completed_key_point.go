package models

import (
	"time"

	"github.com/google/uuid"
)

type CompletedKeyPoint struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TourExecutionID uuid.UUID `gorm:"type:uuid;not null;index"`
	KeyPointID      uuid.UUID `gorm:"type:uuid;not null"`
	CompletedAt     time.Time `gorm:"not null"`
}
