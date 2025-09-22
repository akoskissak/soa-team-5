package main

import (
	// "api-gateway/utils"
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"

	"github.com/gorilla/handlers"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"

	"api-gateway/proto/blog"
	"api-gateway/proto/follower"
	stakeproto "api-gateway/proto/stakeholders"
	"api-gateway/utils"
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
	err := godotenv.Load("../.env")
	if err != nil {
		log.Println("No .env file found or failed to load it:", err)
	}

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
		runtime.WithMetadata(func(ctx context.Context, req *http.Request) metadata.MD {
			// za public rute ne salji
			if req.URL.Path == "/api/auth/login" || req.URL.Path == "/api/auth/register" {
				return nil
			}

			md, err := utils.AuthMetadata(req.Context())
			if err != nil {
				log.Printf("Could not create auth metadata: %v", err)
				return nil
			}
			return md
		}),
	)

	stakeholdersConn, err := grpc.Dial("stakeholders-service:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	//stakeholdersConn, err := grpc.Dial("localhost:8081", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect to stakeholders service for middleware: %v", err)
	}
	defer stakeholdersConn.Close()
	stakeholdersClient := stakeproto.NewStakeholdersServiceClient(stakeholdersConn)

	opts := []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}

	err = stakeproto.RegisterStakeholdersServiceHandlerFromEndpoint(ctx, mux, "stakeholders-service:8081", opts)
	//err = stakeproto.RegisterStakeholdersServiceHandlerFromEndpoint(ctx, mux, "localhost:8081", opts)
	if err != nil {
		log.Fatalf("failed to register stakeholders service: %v", err)
	}

	err = blog.RegisterBlogServiceHandlerFromEndpoint(ctx, mux, "blog-service:8087", opts)
	//err = blog.RegisterBlogServiceHandlerFromEndpoint(ctx, mux, "localhost:8087", opts)
	if err != nil {
		log.Fatalf("failed to register blog service: %v", err)
	}

	err = follower.RegisterFollowerServiceHandlerFromEndpoint(ctx, mux, "follower-service:8084", opts)
	//err = follower.RegisterFollowerServiceHandlerFromEndpoint(ctx, mux, "localhost:8084", opts)
	if err != nil {
		log.Fatalf("failed to register follower service: %v", err)
	}

	tourProxy := newReverseProxy("http://tours-service:8083")
	//tourProxy := newReverseProxy("http://localhost:8083")
	purchaseProxy := newReverseProxy("http://purchase-service:8088")
	//purchaseProxy := newReverseProxy("http://localhost:8088")

	proxyHandlerFunc := func(proxy http.Handler) runtime.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request, pathParams map[string]string) {
			proxy.ServeHTTP(w, r)
		}
	}

	mux.HandlePath("GET", "/api/tours", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/tours", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/keypoints", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/tours/{tourId}/start", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/reviews", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tours/{tourId}/reviews", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tours/published", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/uploads/{path=**}", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tours/{tourId}/keypoints", proxyHandlerFunc(tourProxy))
	mux.HandlePath("PUT", "/api/keypoints/{id}", proxyHandlerFunc(tourProxy))
	mux.HandlePath("DELETE", "/api/keypoints/{id}", proxyHandlerFunc(tourProxy))
	mux.HandlePath("PATCH", "/api/tour-executions/{tourExecutionId}/status", proxyHandlerFunc(tourProxy))
	mux.HandlePath("GET", "/api/tour-executions/active", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/tour-executions/{tourExecutionId}/check-location", proxyHandlerFunc(tourProxy))
	mux.HandlePath("PATCH", "/api/tours/{tourId}/publish", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/tours/{tourId}/required-times", proxyHandlerFunc(tourProxy))
	mux.HandlePath("PATCH", "/api/tours/{tourId}/archive", proxyHandlerFunc(tourProxy))
	mux.HandlePath("PATCH", "/api/tours/{tourId}/unarchive", proxyHandlerFunc(tourProxy))
	mux.HandlePath("POST", "/api/shopping-cart/{touristId}", proxyHandlerFunc(purchaseProxy))
	mux.HandlePath("POST", "/api/shopping-cart/{touristId}/items", proxyHandlerFunc(purchaseProxy))
	mux.HandlePath("GET", "/api/shopping-cart/{touristId}", proxyHandlerFunc(purchaseProxy))
	mux.HandlePath("DELETE", "/api/shopping-cart/{touristId}/items/{tourId}", proxyHandlerFunc(purchaseProxy))
	mux.HandlePath("POST", "/api/shopping-cart/{touristId}/checkout", proxyHandlerFunc(purchaseProxy))
	mux.HandlePath("GET", "/api/tourist/{touristId}/purchases", proxyHandlerFunc(purchaseProxy))

	// CORS
	/*headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})

	finalHandler := utils.JWTMiddleware(mux)
	corsHandler := handlers.CORS(originsOk, headersOk, methodsOk)(finalHandler)
	finalHandlerWithLogging := loggingHandler(corsHandler)

	log.Printf("server listening on port %s", port)
	if err := http.ListenAndServe(port, finalHandlerWithLogging); err != nil {
		log.Fatalf("could not start server: %v", err)
	}*/

	mainRouter := http.NewServeMux()

	headersOk := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{"*"})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"})

	apiHandler := utils.JWTMiddleware(mux, stakeholdersClient)
	apiHandler = handlers.CORS(originsOk, headersOk, methodsOk)(apiHandler)
	apiHandler = loggingHandler(apiHandler)

	// Svi zahtjevi koji NISU za slike idu na ovaj handler koji ima JWT
	mainRouter.Handle("/", apiHandler)

	// --- 2. Handler za SLIKE (bez JWT provjere) ---
	var uploadsHandler http.Handler = tourProxy

	uploadsHandler = handlers.CORS(originsOk, headersOk, methodsOk)(uploadsHandler)
	uploadsHandler = loggingHandler(uploadsHandler)

	mainRouter.Handle("/uploads/", uploadsHandler)

	log.Printf("server listening on port %s", port)
	if err := http.ListenAndServe(port, mainRouter); err != nil {
		log.Fatalf("could not start server: %v", err)
	}

}
