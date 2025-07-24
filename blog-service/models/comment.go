package models

import (
	"time"

	"github.com/google/uuid"
)

type Comment struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	PostID    uuid.UUID `gorm:"type:uuid;not null;index" json:"postId"`
	UserID    uuid.UUID `gorm:"type:uuid;not null" json:"userId"`
	Username  string    `gorm:"type:varchar(255);not null" json:"username"`
	Text      string    `gorm:"type:text;not null" json:"text"`
	CreatedAt time.Time `gorm:"default:now();not null" json:"createdAt"`
	UpdatedAt time.Time `gorm:"default:now();not null" json:"updatedAt"`
}
