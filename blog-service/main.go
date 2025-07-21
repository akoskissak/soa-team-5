package main

import (
	"fmt"
	"log"
	"net/http"

	"soa/blog-service/database"
	"soa/blog-service/handlers"
)

func main() {
	database.InitDB()
	defer database.CloseDB()

	mux := http.NewServeMux()

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Blog Service is running!")
	})

	mux.HandleFunc("POST /posts", handlers.CreatePost)
	mux.HandleFunc("GET /posts", handlers.GetPosts)
	mux.HandleFunc("GET /posts/{id}", handlers.GetPostByID)

	port := ":8081"
	fmt.Printf("Blog Service starting on port %s\n", port)
	log.Fatal(http.ListenAndServe(port, mux))
}
