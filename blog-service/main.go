package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"

	blogproto "soa/blog-service/proto/blog"
	followerproto "soa/blog-service/proto/follower"

	"soa/blog-service/database"
	"soa/blog-service/handlers"

	cors "github.com/gorilla/handlers"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
)

func main() {
	if os.Getenv("DATABASE_URL") == "" {
		err := godotenv.Load("../.env")
		if err != nil {
			log.Println("Error loading .env file:", err)
		}
	}

	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		log.Fatal("DATABASE_URL not set in .env file or environment.")
	}

	database.InitDB(connStr)
	defer database.CloseDB()

	followerServiceAddress := "localhost:8084"
	followerConn, err := grpc.Dial(
		followerServiceAddress,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("Failed to dial follower service: %v", err)
	}
	defer followerConn.Close()

	followerClient := followerproto.NewFollowerServiceClient(followerConn)
	handlers.InitFollowerClient(followerClient)

	go func() {
		httpPort := "8086"
		httpMux := http.NewServeMux()

		fs := http.FileServer(http.Dir("static/uploads"))
		httpMux.Handle("/uploads/", http.StripPrefix("/uploads/", fs))

		httpMux.HandleFunc("/upload-image", handlers.HandleImageUpload)

		headersOk := cors.AllowedHeaders([]string{
			"X-Requested-With", "Content-Type", "Authorization",
		})
		originsOk := cors.AllowedOrigins([]string{"*"})
		methodsOk := cors.AllowedMethods([]string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		})

		corsHandler := cors.CORS(originsOk, headersOk, methodsOk)(httpMux)

		log.Printf("Blog Service HTTP static server listening on port %s", httpPort)
		if err := http.ListenAndServe(":"+httpPort, corsHandler); err != nil {
			log.Fatalf("Failed to start HTTP static server: %v", err)
		}
	}()

	port := "8087"
	lis, err := net.Listen("tcp", fmt.Sprintf(":%s", port))
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	blogServer := handlers.NewBlogServer()
	blogproto.RegisterBlogServiceServer(grpcServer, blogServer)

	reflection.Register(grpcServer)

	log.Printf("Blog gRPC service listening on port %s", port)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
