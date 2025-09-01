package models

import (
	"github.com/google/uuid"
	
	"gorm.io/datatypes"
)

type Tour struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `gorm:"type:varchar(24);not null;column:user_id"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Difficulty  TourDifficulty `json:"difficulty"`
	Tags        datatypes.JSON `gorm:"type:jsonb" json:"tags"`
	Status      TourStatus     `json:"status"`
	Price       float32        `json:"price"`
}