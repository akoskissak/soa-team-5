package main

import (
	"log"
	"os"
	"tours-service/database"
	"tours-service/handlers"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	localhost := "0.0.0.0"

	connStr := os.Getenv("TOUR_DATABASE_URL")
	if connStr == "" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Println(err)
		}

		connStr = os.Getenv("TOUR_DATABASE_URL")
		localhost = "localhost"
	}

	database.Connect(connStr)

	r := gin.Default()

	r.Static("/uploads", "./static/uploads")

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:4200"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")

	api.POST("/tours", handlers.CreateTour)
	api.GET("/tours", handlers.GetAllTours)

	api.GET("/tours/published", handlers.GetAllPublishedTours)

	api.POST("/keypoints", handlers.CreateKeyPoint)
	api.GET("/tours/:tourId/keypoints", handlers.GetKeyPointsByTourId)
	api.PUT("/keypoints/:id", handlers.UpdateKeyPoint)
	api.DELETE("/keypoints/:id", handlers.DeleteKeyPoint)

	api.POST("/reviews", handlers.CreateReview)

	api.GET("/tours/:tourId/reviews", handlers.GetReviewsByTourId)

	localhost = "tours-service"
	r.Run(localhost + ":8083")
}
