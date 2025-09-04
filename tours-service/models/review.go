package models

import (
	"time"

	"github.com/google/uuid"
)

type Review struct {
	ID             uuid.UUID     `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	TourID         uuid.UUID     `json:"tourId" gorm:"not null"`
	TouristID      string        `json:"touristId" gorm:"type:varchar(24);not null"`
	Username       string        `json:"username" gorm:"type:varchar(255);not null"`
	Rating         int           `json:"rating" gorm:"not null"`
	Comment        string        `json:"comment"`
	SubmissionDate time.Time     `json:"submissionDate" gorm:"not null"`
	VisitedDate    time.Time     `json:"visitedDate"`
	ReviewImages   []ReviewImage `gorm:"foreignKey:ReviewID"`
}
