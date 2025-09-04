package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/encoding/protojson"

	"api-gateway/proto/blog"
	"api-gateway/proto/follower"
	"api-gateway/proto/stakeholders"
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

func newReverseProxy(target string) http.Handler {
	u, err := url.Parse(target)
	if err != nil {
		log.Fatalf("could not parse target url %s: %v", target, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = u.Host
		if origin := req.Header.Get("Origin"); origin != "" {
			req.Header.Set("Origin", u.Scheme+"://"+u.Host)
		}
	}

	proxy.ErrorLog = log.New(os.Stderr, "PROXY ERR: ", log.LstdFlags)

	return proxy
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

	err = blog.RegisterBlogServiceHandlerFromEndpoint(ctx, mux, "localhost:8087", opts)
	if err != nil {
		log.Fatalf("failed to register blog service: %v", err)
	}

	err = follower.RegisterFollowerServiceHandlerFromEndpoint(ctx, mux, "localhost:8084", opts)
	if err != nil {
		log.Fatalf("failed to register follower service: %v", err)
	}

	tourProxy := newReverseProxy("http://localhost:8083")

	proxyHandlerFunc := func(proxy http.Handler) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			proxy.ServeHTTP(w, r)
		}
	}

	mux.HandlePath("GET", "/api/tours", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/tours", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/keypoints", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/reviews", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tours/{tourId}/reviews", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tours/published", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/uploads/{path=**}", proxyHandlerFunc(tourProxy))

	// CORS
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
