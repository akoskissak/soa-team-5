package main

import (
	"log"
	"stakeholders-service/db"
	"stakeholders-service/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println(err)
	}
	db.Connect()

	r := gin.Default()

	// cors zbog angulara
	r.Use(cors.New(cors.Config{
        AllowOrigins:     []string{"http://localhost:4200"},
        AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
        ExposeHeaders:    []string{"Content-Length"},
        AllowCredentials: true,
        MaxAge:           12 * time.Hour,
    }))

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.Register)
		api.GET("/admin/users", handlers.GetAllUsers)
		api.POST("/auth/login", handlers.Login)
	}
	
	r.Run(":8080")
}