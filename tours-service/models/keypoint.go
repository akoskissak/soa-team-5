package models

import "github.com/google/uuid"

type KeyPoint struct {
	ID          uuid.UUID `json:"id" gorm:"type:uuid;default:gen_random_uuid();primaryKey"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Latitude    float64   `json:"latitude"`
	Longitude   float64   `json:"longitude"`
	ImagePath   string    `json:"imagePath"`
	TourID      uuid.UUID `json:"tourId"`
}