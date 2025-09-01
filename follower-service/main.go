package main

import (
	"follower-service/db"
	"follower-service/handlers"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	localhost := "0.0.0.0"

	uri := os.Getenv("NEO4J_URI")
	user := os.Getenv("NEO4J_USER")
	pass := os.Getenv("NEO4J_PASS")
	port := "8082"

	if uri == "" || user == "" || pass == "" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Println("No .env file found or failed to load it:", err)
		}
		uri = os.Getenv("NEO4J_URI")
		user = os.Getenv("NEO4J_USER")
		pass = os.Getenv("NEO4J_PASS")
		localhost = "localhost"
	}

	db.ConnectNeo4j(uri, user, pass)

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
		api.POST("/follow", handlers.Follow)
		// api.DELETE("/follow", handlers.Unfollow)
		api.DELETE("/follow/:to", handlers.Unfollow)
		api.GET("/following/:username", handlers.GetFollowing)
		api.GET("/followers/:username", handlers.GetFollowers)
		api.GET("/recommend", handlers.Recommend)
	}

	log.Printf("follower-service listening on :%s", port)
	r.Run(localhost + ":" + port)
}
