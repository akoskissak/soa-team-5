package main

import (
	"follower-service/db"
	"follower-service/handlers"
	"log"
	"net"
	"os"

	pb "follower-service/proto/follower"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
)

func main() {
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("Upozorenje: .env fajl nije pronađen ili se ne može učitati. Pokušavam koristiti sistemske varijable.")
	}

	uri := os.Getenv("NEO4J_URI")
	user := os.Getenv("NEO4J_USER")
	pass := os.Getenv("NEO4J_PASS")

	if uri == "" || user == "" || pass == "" {
		log.Fatalf("Greška: NEO4J varijable okruženja nisu postavljene. Provjerite .env fajl.")
	}

	db.ConnectNeo4j(uri, user, pass)

	lis, err := net.Listen("tcp", ":8084")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterFollowerServiceServer(grpcServer, handlers.NewFollowerServer())

	log.Println("Follower service listening on :8084")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
