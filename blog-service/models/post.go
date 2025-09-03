package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
)

type Post struct {
	ID          uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID      string         `gorm:"type:varchar(24);not null"`
	Username    string         `gorm:"type:varchar(255);not null"`
	Title       string         `gorm:"type:varchar(255);not null"`
	Description string         `gorm:"type:text;not null"`
	CreatedAt   time.Time      `gorm:"default:now();not null"`
	ImageURLs   pq.StringArray `gorm:"type:text[]" json:"imageURLs"`
	LikesCount  int            `gorm:"default:0;not null"`
}
