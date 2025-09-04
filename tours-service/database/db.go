package database

import (
	"log"
	"tours-service/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var GORM_DB *gorm.DB

func Connect(connStr string) {
	if connStr == "" {
		log.Fatal("TOUR_DATABASE_URL is not set")
	}
	db, err := gorm.Open(postgres.Open(connStr), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	if err := db.AutoMigrate(&models.Tour{}, &models.KeyPoint{}, &models.Review{}, &models.ReviewImage{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	GORM_DB = db
}
