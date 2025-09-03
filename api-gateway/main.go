package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"api-gateway/proto/blog"
	"api-gateway/proto/follower"
	"api-gateway/proto/stakeholders"
	"api-gateway/proto/tours"
)

const (
	port = ":8080"
)

// loggingHandler logs requests before they are handled by the mux.
func loggingHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("Request received: %s %s from %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	jsonpb := &runtime.JSONPb{
		MarshalOptions: protojson.MarshalOptions{
			UseProtoNames: false,
		},
		UnmarshalOptions: protojson.UnmarshalOptions{
			DiscardUnknown: true,
		},
	}

	mux := runtime.NewServeMux(
		runtime.WithMarshalerOption(runtime.MIMEWildcard, jsonpb),
	)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err := stakeholders.RegisterStakeholdersServiceHandlerFromEndpoint(ctx, mux, "localhost:8081", opts)
	if err != nil {
		log.Fatalf("failed to register stakeholders service: %v", err)
	}

	err = tours.RegisterToursServiceHandlerFromEndpoint(ctx, mux, "tours-service:8082", opts)
	if err != nil {
		log.Fatalf("failed to register tours service: %v", err)
	}

	err = blog.RegisterBlogServiceHandlerFromEndpoint(ctx, mux, "localhost:8083", opts)
	if err != nil {
		log.Fatalf("failed to register blog service: %v", err)
	}

	err = follower.RegisterFollowerServiceHandlerFromEndpoint(ctx, mux, "localhost:8084", opts)
	if err != nil {
		log.Fatalf("failed to register follower service: %v", err)
	}

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"})

	corsHandler := handlers.CORS(originsOk, headersOk, methodsOk)(mux)
	finalHandler := loggingHandler(corsHandler)

	log.Printf("server listening on port %s", port)
	if err := http.ListenAndServe(port, finalHandler); err != nil {
		log.Fatalf("could not start server: %v", err)
	}
}
