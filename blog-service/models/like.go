package models

import (
	"time"

	"github.com/google/uuid"
)

type Like struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index" json:"postId"`
	UserID    string    `gorm:"type:varchar(24);not null" json:"userId"`
	CreatedAt time.Time `gorm:"default:now();not null" json:"createdAt"`
}
