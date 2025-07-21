package main

import (
	"fmt"
	"log"
	"net/http"

	"soa/blog-service/database"
	"soa/blog-service/handlers"

	"github.com/rs/cors"
)

func main() {
	database.InitDB()
	defer database.CloseDB()

	mux := http.NewServeMux()

	/*mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Blog Service is running!")
	})*/

	mux.HandleFunc("POST /posts", handlers.CreatePost)
	mux.HandleFunc("GET /posts", handlers.GetPosts)
	mux.HandleFunc("GET /posts/{id}", handlers.GetPostByID)
	mux.HandleFunc("POST /upload-image", handlers.UploadImage)

	fs := http.FileServer(http.Dir("static/uploads"))
	mux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:4200"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	})

	handler := c.Handler(mux)

	port := ":8081"
	fmt.Printf("Blog Service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, handler))
}
