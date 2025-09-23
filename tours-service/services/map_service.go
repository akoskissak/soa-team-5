package services

import (
	"log"
	"math"
	"tours-service/database"
	"tours-service/models"

	"github.com/google/uuid"
)

func getDistanceAndDurationLocally(kp1, kp2 models.KeyPoint, transportType models.TransportationType) (float64, float64) {
	distanceInKm := haversineDistance(kp1.Latitude, kp1.Longitude, kp2.Latitude, kp2.Longitude)

	var averageSpeedKmH float64
	switch transportType {
	case models.Walking:
		averageSpeedKmH = 5
	case models.Bicycle:
		averageSpeedKmH = 15
	case models.Car:
		averageSpeedKmH = 40
	default:
		averageSpeedKmH = 40
	}

	durationInHours := distanceInKm / averageSpeedKmH
	durationInMinutes := durationInHours * 60

	return distanceInKm, durationInMinutes
}

func UpdateTourDistanceAndTimes(tourID uuid.UUID) error {
	var keypoints []models.KeyPoint
	if err := database.GORM_DB.Where("tour_id = ?", tourID).Order("position asc").Find(&keypoints).Error; err != nil {
		return err
	}

	if len(keypoints) < 2 {
		tx := database.GORM_DB.Begin()
		if err := tx.Model(&models.Tour{}).Where("id = ?", tourID).Update("distance", 0).Error; err != nil {
			tx.Rollback()
			return err
		}
		if err := tx.Where("tour_id = ?", tourID).Delete(&models.RequiredTime{}).Error; err != nil {
			tx.Rollback()
			return err
		}
		return tx.Commit().Error
	}

	var totalDistance, totalDurationWalking, totalDurationBicycle, totalDurationCar float64

	for i := 0; i < len(keypoints)-1; i++ {
		distance, durationWalk := getDistanceAndDurationLocally(keypoints[i], keypoints[i+1], models.Walking)
		totalDurationWalking += durationWalk

		_, durationBike := getDistanceAndDurationLocally(keypoints[i], keypoints[i+1], models.Bicycle)
		totalDurationBicycle += durationBike

		_, durationCar := getDistanceAndDurationLocally(keypoints[i], keypoints[i+1], models.Car)
		totalDurationCar += durationCar

		totalDistance += distance
	}

	tx := database.GORM_DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.Tour{}).Where("id = ?", tourID).Update("distance", totalDistance).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tour_id = ?", tourID).Delete(&models.RequiredTime{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	requiredTimes := []models.RequiredTime{
		{TourID: tourID, Transportation: models.Walking, TimeInMinutes: int(math.Round(totalDurationWalking))},
		{TourID: tourID, Transportation: models.Bicycle, TimeInMinutes: int(math.Round(totalDurationBicycle))},
		{TourID: tourID, Transportation: models.Car, TimeInMinutes: int(math.Round(totalDurationCar))},
	}

	if err := tx.Create(&requiredTimes).Error; err != nil {
		tx.Rollback()
		return err
	}
	log.Printf("Successfully updated tour %s with local distance calculation.", tourID)
	return tx.Commit().Error
}
