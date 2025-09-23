package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"stakeholders-service/db"
	"stakeholders-service/handlers"
	"time"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	stakeproto "stakeholders-service/proto/stakeholders"

	"github.com/nats-io/nats.go"
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

	// natsURL := os.Getenv("NATS_URL")
	// if natsURL == "" {
	// 		natsURL = "nats://nats:4222"
	// }
	
	// fmt.Println("NATS: ", natsURL)

	var natsConn *nats.Conn
	natsURL := os.Getenv("NATS_URL")
	for i := 0; i < 10; i++ {
			natsConn, err = nats.Connect(natsURL)
			if err == nil {
					break
			}
			log.Printf("Waiting for NATS to be ready... (%v)", err)
			time.Sleep(2 * time.Second)
	}
	if err != nil {
			log.Fatalf("Failed to connect to NATS after retries: %v", err)
	}

	defer natsConn.Close()

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

	handlers.SubscribePurchaseCheckout(natsConn, stakeholdersServer)

	stakeproto.RegisterStakeholdersServiceServer(grpcServer, stakeholdersServer)

	reflection.Register(grpcServer)

	log.Printf("Stakeholders gRPC service listening at %v", lis.Addr())

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve gRPC server: %v", err)
	}
}
