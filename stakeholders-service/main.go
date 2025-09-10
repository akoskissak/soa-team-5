package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"stakeholders-service/db"
	"stakeholders-service/handlers"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	stakeproto "stakeholders-service/proto/stakeholders"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found or failed to load it:", err)
	}

	mongoUri := os.Getenv("MONGODB_URI")
	mongoClient, err := db.ConnectMongoDB(mongoUri)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Printf("Failed to disconnect MongoDB client: %v", err)
		}
	}()

	/*neo4jURI := os.Getenv("NEO4J_URI")
	neo4jUser := os.Getenv("NEO4J_USER")
	neo4jPass := os.Getenv("NEO4J_PASS")
	db.ConnectNeo4j(neo4jURI, neo4jUser, neo4jPass)*/

	port := os.Getenv("STAKEHOLDERS_SERVICE_PORT")
	if port == "" {
		port = "8081"
	}

	go func() {
		httpPort := os.Getenv("STAKEHOLDERS_STATIC_PORT")
		if httpPort == "" {
			httpPort = "8085"
		}

		httpMux := http.NewServeMux()

		httpMux.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir("./static/uploads"))))

		log.Printf("Stakeholders HTTP static server listening at :%v", httpPort)
		if err := http.ListenAndServe(":"+httpPort, httpMux); err != nil {
			log.Fatalf("Failed to start HTTP static server: %v", err)
		}
	}()

	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	stakeholdersServer := handlers.NewStakeholdersServer(mongoClient)

	stakeproto.RegisterStakeholdersServiceServer(grpcServer, stakeholdersServer)

	reflection.Register(grpcServer)

	log.Printf("Stakeholders gRPC service listening at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
