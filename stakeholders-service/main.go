package main

import (
	"log"
	"os"
	"stakeholders-service/db"
	"stakeholders-service/handlers"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	localhost := "0.0.0.0"

	mongoUri := os.Getenv("MONGODB_URI")
	if mongoUri == "" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Println("No .env file found or failed to load it:", err)
		}

		mongoUri = os.Getenv("MONGODB_URI")
		localhost = "localhost"
	}
	db.ConnectMongoDB(mongoUri)

    neo4jURI := os.Getenv("NEO4J_URI")
    neo4jUser := os.Getenv("NEO4J_USER")
    neo4jPass := os.Getenv("NEO4J_PASS")
	db.ConnectNeo4j(neo4jURI, neo4jUser, neo4jPass)

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

	r.Static("/uploads", "./static/uploads")

	api := r.Group("/api")
	{
		api.POST("/auth/register", handlers.Register)
		api.GET("/admin/users", handlers.GetAllUsers)
		api.POST("/auth/login", handlers.Login)
		api.PUT("/admin/block-user", handlers.BlockUser)
		api.GET("/user/profile/:username", handlers.GetProfileByUsername)
		api.GET("/user/profile", handlers.GetProfile)
		api.PUT("/user/profile", handlers.UpdateProfile)
	}

	r.Run(localhost + ":8080")
}
