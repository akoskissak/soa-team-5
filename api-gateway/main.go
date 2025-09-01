package main

import (
	"context"
	"log"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"api-gateway/proto/blog"
	"api-gateway/proto/follower"
	"api-gateway/proto/stakeholders"
	"api-gateway/proto/tours"
)

const (
	port = ":8080"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	mux := runtime.NewServeMux()

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := stakeholders.RegisterStakeholdersServiceHandlerFromEndpoint(ctx, mux, "stakeholders-service:8081", opts)
	if err != nil {
		log.Fatalf("failed to register stakeholders service: %v", err)
	}

	err = tours.RegisterToursServiceHandlerFromEndpoint(ctx, mux, "tours-service:8082", opts)
	if err != nil {
		log.Fatalf("failed to register tours service: %v", err)
	}

	err = blog.RegisterBlogServiceHandlerFromEndpoint(ctx, mux, "blog-service:8083", opts)
	if err != nil {
		log.Fatalf("failed to register blog service: %v", err)
	}

	err = follower.RegisterFollowerServiceHandlerFromEndpoint(ctx, mux, "follower-service:8084", opts)
	if err != nil {
		log.Fatalf("failed to register follower service: %v", err)
	}

	log.Printf("server listening on port %s", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
