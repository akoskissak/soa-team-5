package models

import "github.com/google/uuid"

type RequiredTime struct {
	ID             uuid.UUID          `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	TourID         uuid.UUID          `gorm:"type:uuid;not null"`
	Transportation TransportationType `gorm:"type:varchar(20);not null"`
	TimeInMinutes  int                `gorm:"not null"`
}
