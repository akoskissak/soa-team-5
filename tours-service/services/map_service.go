package services

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"tours-service/database"
	"tours-service/models"

	"github.com/google/uuid"
)

type OSRMResponse struct {
	Routes []struct {
		Distance float64 `json:"distance"`
		Duration float64 `json:"duration"`
	} `json:"routes"`
}

func getDistanceAndDurationBetweenPoints(kp1, kp2 models.KeyPoint, transportType models.TransportationType) (float64, float64, error) {
	var profile string
	switch transportType {
	case models.Walking:
		profile = "foot"
	case models.Bicycle:
		profile = "bike"
	case models.Car:
		profile = "driving"
	default:
		profile = "driving"
	}

	url := fmt.Sprintf("http://router.project-osrm.org/route/v1/%s/%f,%f;%f,%f?overview=false",
		profile, kp1.Longitude, kp1.Latitude, kp2.Longitude, kp2.Latitude)

	resp, err := http.Get(url)
	if err != nil {
		return 0, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, 0, fmt.Errorf("OSRM API returned non-200 status code: %d for profile %s", resp.StatusCode, profile)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, 0, err
	}

	var osrmResponse OSRMResponse
	if err := json.Unmarshal(body, &osrmResponse); err != nil {
		return 0, 0, err
	}

	if len(osrmResponse.Routes) > 0 {
		distanceInKm := osrmResponse.Routes[0].Distance / 1000
		durationInMinutes := osrmResponse.Routes[0].Duration / 60
		return distanceInKm, durationInMinutes, nil
	}

	return 0, 0, fmt.Errorf("no route found in OSRM response for profile %s", profile)
}

func UpdateTourDistanceAndTimes(tourID uuid.UUID) error {
	var keypoints []models.KeyPoint
	if err := database.GORM_DB.Where("tour_id = ?", tourID).Find(&keypoints).Error; err != nil {
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

	var totalDistanceCar, totalDurationWalking, totalDurationBicycle, totalDurationCar float64

	for i := 0; i < len(keypoints)-1; i++ {
		_, durationWalk, err := getDistanceAndDurationBetweenPoints(keypoints[i], keypoints[i+1], models.Walking)
		if err != nil {
			log.Printf("Greška pri dobavljanju trajanja za pešačenje: %v", err)
		}
		totalDurationWalking += durationWalk

		_, durationBike, err := getDistanceAndDurationBetweenPoints(keypoints[i], keypoints[i+1], models.Bicycle)
		if err != nil {
			log.Printf("Greška pri dobavljanju trajanja za bicikl: %v", err)
		}
		totalDurationBicycle += durationBike

		distanceCar, durationCar, err := getDistanceAndDurationBetweenPoints(keypoints[i], keypoints[i+1], models.Car)
		if err != nil {
			log.Printf("Greška pri dobavljanju distance/trajanja za auto: %v", err)
		}
		totalDistanceCar += distanceCar
		totalDurationCar += durationCar
	}

	tx := database.GORM_DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	if err := tx.Model(&models.Tour{}).Where("id = ?", tourID).Update("distance", totalDistanceCar).Error; err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Where("tour_id = ?", tourID).Delete(&models.RequiredTime{}).Error; err != nil {
		tx.Rollback()
		return err
	}

	requiredTimes := []models.RequiredTime{
		{TourID: tourID, Transportation: models.Walking, TimeInMinutes: int(totalDurationWalking)},
		{TourID: tourID, Transportation: models.Bicycle, TimeInMinutes: int(totalDurationBicycle)},
		{TourID: tourID, Transportation: models.Car, TimeInMinutes: int(totalDurationCar)},
	}

	if err := tx.Create(&requiredTimes).Error; err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit().Error
}
