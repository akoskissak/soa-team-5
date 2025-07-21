package models

import (
	"time"

	"github.com/google/uuid" // Potreban za UUID tip
)

// Post predstavlja blog objavu.
type Post struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"` // UUID kao primarni ključ, automatski se generiše
	UserID      uuid.UUID `gorm:"type:uuid;not null"`                             // ID korisnika (od Stakeholders servisa)
	Username    string    `gorm:"type:varchar(255);not null"`                     // Korisničko ime (od Stakeholders servisa)
	Title       string    `gorm:"type:varchar(255);not null"`                     // Naslov objave
	Description string    `gorm:"type:text;not null"`                             // Sadržaj (opis) objave
	CreatedAt   time.Time `gorm:"default:now();not null"`                         // Vreme kreiranja
	ImageURLs   []string  `gorm:"type:text[]"`                                    // Opcija: lista URL-ova slika (kao PostgreSQL TEXT[] tip)
	LikesCount  int       `gorm:"default:0;not null"`                             // Broj lajkova (za sada bez logike, samo kolona)
}
