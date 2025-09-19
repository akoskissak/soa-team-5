package database

import (
	"log"
	"tours-service/models"

	"github.com/uptrace/opentelemetry-go-extra/otelgorm"
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

	if err := db.AutoMigrate(&models.Tour{}, &models.KeyPoint{}, &models.Review{}, &models.ReviewImage{}, &models.TourExecution{}, &models.RequiredTime{}, &models.CompletedKeyPoint{}); err != nil {
		log.Fatal("Failed to migrate database: ", err)
	}

	if err := db.Use(otelgorm.NewPlugin()); err != nil {
		log.Fatal("Failed to use otelgorm: ", err)
	}

	GORM_DB = db
}
