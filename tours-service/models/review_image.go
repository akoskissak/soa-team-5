package models

import "github.com/google/uuid"

type ReviewImage struct {
	ID        uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	ImagePath string    `json:"imagePath"`
	ReviewID  uuid.UUID `json:"reviewId" gorm:"not null"`
}
